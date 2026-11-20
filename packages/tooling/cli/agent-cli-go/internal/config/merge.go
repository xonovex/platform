package config

import (
	"fmt"
	"slices"
)

// Keys is a set of configuration keys.
type Keys map[Key]struct{}

// KeysOf returns the set of keys.
func KeysOf(keys ...Key) Keys {
	set := Keys{}
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

// Has reports whether key is in the set.
func (k Keys) Has(key Key) bool { _, ok := k[key]; return ok }

// Value is the content of one key in one layer: text for a replace key, items for
// an append key, a flag for a require key.
type Value struct {
	Text  string
	Items []string
	On    bool
}

// Entries are the keys one layer sets. A key a layer does not set is absent.
type Entries map[Key]Value

// Source says where a value came from: a layer and, for a file, its path.
type Source struct {
	Layer  Layer
	Origin string
}

// String renders the source as "layer origin".
func (s Source) String() string {
	if s.Origin == "" {
		return string(s.Layer)
	}
	return string(s.Layer) + " " + s.Origin
}

// Input is the entries of one layer with the place they were read from.
type Input struct {
	Layer   Layer
	Origin  string
	Entries Entries
}

func (i Input) source() Source { return Source{Layer: i.Layer, Origin: i.Origin} }

// Item is one value of an append key with its source.
type Item struct {
	Text   string
	Source Source
}

// Resolution is the merged value of one key and where each part came from.
type Resolution struct {
	Text       string
	TextSource Source
	Items      []Item
	On         bool
	OnSource   Source
}

// Resolved is the merge of every layer.
type Resolved struct {
	values  map[Key]Resolution
	sources []Source
}

// Has reports whether any layer set key.
func (r Resolved) Has(key Key) bool { _, ok := r.values[key]; return ok }

// Text returns the merged value of a replace key, empty when no layer set it.
func (r Resolved) Text(key Key) string { return r.values[key].Text }

// On returns the merged value of a require key.
func (r Resolved) On(key Key) bool { return r.values[key].On }

// Items returns the merged values of an append key.
func (r Resolved) Items(key Key) []string {
	items := r.values[key].Items
	if len(items) == 0 {
		return nil
	}
	texts := make([]string, len(items))
	for index, item := range items {
		texts[index] = item.Text
	}
	return texts
}

// Sourced returns the merged values of an append key, each with the layer that
// supplied it, in the order a consumer applies them.
func (r Resolved) Sourced(key Key) []Item { return slices.Clone(r.values[key].Items) }

// Sources lists the layers that took part in the merge, in precedence order from
// the lowest layer up.
func (r Resolved) Sources() []Source { return slices.Clone(r.sources) }

// Provenance describes, one line per value, which layer decided it. A line reads
// "key: value (layer origin)".
func (r Resolved) Provenance() []string {
	var lines []string
	for _, spec := range Specs {
		resolution, ok := r.values[spec.Key]
		if !ok {
			continue
		}
		switch spec.Rule {
		case RuleReplace:
			lines = append(lines, fmt.Sprintf("%s: %s (%s)", spec.Key, resolution.Text, resolution.TextSource))
		case RuleRequire:
			lines = append(lines, fmt.Sprintf("%s: required (%s)", spec.Key, resolution.OnSource))
		case RuleAppend:
			for _, item := range resolution.Items {
				lines = append(lines, fmt.Sprintf("%s: %s (%s)", spec.Key, item.Text, item.Source))
			}
		}
	}
	return lines
}

// Layers returns, for each key a layer set, the layers that decided its value. A
// run record keeps it in place of the values, which can hold a locator or a path.
func (r Resolved) Layers() map[Key][]Layer {
	layers := map[Key][]Layer{}
	for key, resolution := range r.values {
		switch spec, _ := specOf(key); spec.Rule {
		case RuleReplace:
			layers[key] = []Layer{resolution.TextSource.Layer}
		case RuleRequire:
			layers[key] = []Layer{resolution.OnSource.Layer}
		case RuleAppend:
			for _, item := range resolution.Items {
				if !slices.Contains(layers[key], item.Source.Layer) {
					layers[key] = append(layers[key], item.Source.Layer)
				}
			}
		}
	}
	return layers
}

// replaceOrder is the precedence of a replace key, strongest first.
var replaceOrder = []Layer{LayerManaged, LayerFlags, LayerProject, LayerUser}

// appendOrder is the order of an append key's items, weakest first.
var appendOrder = []Layer{LayerUser, LayerProject, LayerFlags, LayerManaged}

// Merge combines the layers by the rule of each key. It fails closed: a layer that
// sets a key it may not set, an unknown key, and a command line value that
// contradicts a managed one are errors, never silent overrides.
func Merge(inputs []Input) (Resolved, error) {
	for _, input := range inputs {
		for key := range input.Entries {
			spec, ok := specOf(key)
			if !ok {
				return Resolved{}, fmt.Errorf("%s sets the unknown key %q", input.source(), key)
			}
			if !spec.allows(input.Layer) {
				return Resolved{}, fmt.Errorf("%s sets %q, which the %s layer may not set", input.source(), key, input.Layer)
			}
		}
	}
	resolved := Resolved{values: map[Key]Resolution{}}
	for _, layer := range []Layer{LayerUser, LayerProject, LayerFlags, LayerManaged} {
		for _, input := range inputs {
			if input.Layer == layer {
				resolved.sources = append(resolved.sources, input.source())
			}
		}
	}
	for _, spec := range Specs {
		var err error
		switch spec.Rule {
		case RuleReplace:
			err = mergeReplace(&resolved, spec.Key, inputs)
		case RuleAppend:
			mergeAppend(&resolved, spec.Key, inputs)
		case RuleRequire:
			mergeRequire(&resolved, spec.Key, inputs)
		}
		if err != nil {
			return Resolved{}, err
		}
	}
	return resolved, nil
}

// entry returns the value a layer's input holds for key.
func entry(inputs []Input, layer Layer, key Key) (Value, Source, bool) {
	for _, input := range inputs {
		if input.Layer != layer {
			continue
		}
		if value, ok := input.Entries[key]; ok {
			return value, input.source(), true
		}
	}
	return Value{}, Source{}, false
}

func mergeReplace(resolved *Resolved, key Key, inputs []Input) error {
	for _, layer := range replaceOrder {
		value, source, ok := entry(inputs, layer, key)
		if !ok || value.Text == "" {
			continue
		}
		if layer == LayerManaged {
			if flag, flagSource, set := entry(inputs, LayerFlags, key); set && flag.Text != "" && flag.Text != value.Text {
				return fmt.Errorf("%s=%s on the %s contradicts %s=%s from %s; the managed value cannot be changed", key, flag.Text, flagSource, key, value.Text, source)
			}
		}
		resolved.values[key] = Resolution{Text: value.Text, TextSource: source}
		return nil
	}
	return nil
}

func mergeAppend(resolved *Resolved, key Key, inputs []Input) {
	var items []Item
	seen := map[string]bool{}
	for _, layer := range appendOrder {
		for _, input := range inputs {
			if input.Layer != layer {
				continue
			}
			for _, text := range input.Entries[key].Items {
				if seen[text] {
					continue
				}
				seen[text] = true
				items = append(items, Item{Text: text, Source: input.source()})
			}
		}
	}
	if len(items) > 0 {
		resolved.values[key] = Resolution{Items: items}
	}
}

func mergeRequire(resolved *Resolved, key Key, inputs []Input) {
	for _, layer := range replaceOrder {
		if value, source, ok := entry(inputs, layer, key); ok && value.On {
			resolved.values[key] = Resolution{On: true, OnSource: source}
			return
		}
	}
}
