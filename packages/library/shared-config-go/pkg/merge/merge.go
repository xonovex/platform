// Package merge combines the layers of the agentic configuration key by key,
// by the rule each key declares, and remembers which layer gave each value.
package merge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// Flag is one key the command line sets.
type Flag struct {
	Section string
	Pointer string
	Value   json.RawMessage
}

// Origin is where a value comes from: a layer and, for a file, its path.
type Origin struct {
	Layer spec.Layer
	Path  string
}

// Value is the merged value of one key.
type Value struct {
	Key spec.Key
	Raw json.RawMessage
	// Origins are the layers that gave the value.
	Origins []Origin
	// Entries holds the origin of each entry of a Merge key.
	Entries map[string]Origin
}

// Resolved is the merged configuration of the requested sections.
type Resolved struct {
	values []Value
}

// Values returns every merged value in declaration order.
func (r Resolved) Values() []Value { return append([]Value(nil), r.values...) }

// Source returns the merged value of a key.
func (r Resolved) Source(section, pointer string) (Value, bool) {
	for _, value := range r.values {
		if value.Key.Section == section && value.Key.Pointer == pointer {
			return value, true
		}
	}
	return Value{}, false
}

// Section renders the merged values of a section as one JSON object that the
// owning tool decodes.
func (r Resolved) Section(name string) json.RawMessage {
	root := map[string]any{}
	for _, value := range r.values {
		if value.Key.Section != name {
			continue
		}
		if value.Key.Pointer == "" {
			return value.Raw
		}
		tokens := strings.Split(strings.TrimPrefix(value.Key.Pointer, "/"), "/")
		node := root
		for _, token := range tokens[:len(tokens)-1] {
			child, ok := node[token].(map[string]any)
			if !ok {
				child = map[string]any{}
				node[token] = child
			}
			node = child
		}
		node[tokens[len(tokens)-1]] = value.Raw
	}
	// The tree holds only maps and raw values that were parsed as JSON.
	data, _ := json.Marshal(root)
	return data
}

// setting is the value one layer gives a key.
type setting struct {
	origin Origin
	raw    json.RawMessage
}

// replaceOrder is the precedence of Replace and Merge, the winner first.
var replaceOrder = [...]spec.Layer{spec.Managed, spec.Flags, spec.Project, spec.User}

// appendOrder is the order Append concatenates the layers in, the managed last.
var appendOrder = [...]spec.Layer{spec.User, spec.Project, spec.Flags, spec.Managed}

// Merge combines the documents and the flags for the sections. A document that
// sets a key its layer may not set, or a flag that names no key, is an error.
// The documents must be parsed and checked.
func Merge(sections []spec.Section, documents []layer.Document, flags []Flag) (Resolved, error) {
	if err := checkFlags(sections, flags); err != nil {
		return Resolved{}, err
	}
	var values []Value
	for _, section := range sections {
		for _, key := range section.Keys {
			settings, err := collect(key, documents, flags)
			if err != nil {
				return Resolved{}, err
			}
			value, set, err := combine(key, settings)
			if err != nil {
				return Resolved{}, err
			}
			if set {
				values = append(values, value)
			}
		}
	}
	return Resolved{values: values}, nil
}

func checkFlags(sections []spec.Section, flags []Flag) error {
	for _, flag := range flags {
		known := false
		for _, section := range sections {
			if _, ok := section.Key(flag.Pointer); ok && section.Name == flag.Section {
				known = true
			}
		}
		if !known {
			return fmt.Errorf("the command line sets %s%s, which is not a key of the requested sections", flag.Section, flag.Pointer)
		}
		if !json.Valid(flag.Value) {
			return fmt.Errorf("the command line sets %s%s to %q, which is not JSON", flag.Section, flag.Pointer, flag.Value)
		}
	}
	return nil
}

// collect gathers the value each layer gives key, by layer.
func collect(key spec.Key, documents []layer.Document, flags []Flag) (map[spec.Layer]setting, error) {
	settings := map[spec.Layer]setting{}
	for _, document := range documents {
		raw, ok, err := lookup(document.Sections[key.Section], key.Pointer)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", document.Path, key.Name(), err)
		}
		if !ok {
			continue
		}
		if !key.Allows(document.Layer) {
			return nil, fmt.Errorf("%s: the %s file sets %s, which only the %s files may set", document.Path, document.Layer, key.Name(), layerList(key.Layers))
		}
		settings[document.Layer] = setting{origin: Origin{Layer: document.Layer, Path: document.Path}, raw: raw}
	}
	for _, flag := range flags {
		if flag.Section == key.Section && flag.Pointer == key.Pointer {
			settings[spec.Flags] = setting{origin: Origin{Layer: spec.Flags}, raw: flag.Value}
		}
	}
	return settings, nil
}

func layerList(layers []spec.Layer) string {
	names := make([]string, len(layers))
	for i, l := range layers {
		names[i] = string(l)
	}
	return strings.Join(names, " and ")
}

// lookup returns the value at pointer inside a section, which may be absent.
func lookup(section json.RawMessage, pointer string) (json.RawMessage, bool, error) {
	if section == nil {
		return nil, false, nil
	}
	if pointer == "" {
		return section, true, nil
	}
	current := section
	for _, token := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(current, &object); err != nil {
			return nil, false, fmt.Errorf("the parent of %s is not an object", token)
		}
		next, ok := object[token]
		if !ok {
			return nil, false, nil
		}
		current = next
	}
	return current, true, nil
}

func combine(key spec.Key, settings map[spec.Layer]setting) (Value, bool, error) {
	if len(settings) == 0 && key.Rule != spec.Cap {
		return Value{}, false, nil
	}
	switch key.Rule {
	case spec.Replace:
		value, err := replaceValue(key, settings)
		return value, true, err
	case spec.Append:
		value, err := appendValue(key, settings)
		return value, true, err
	case spec.Require:
		value, err := requireValue(key, settings)
		return value, true, err
	case spec.Cap:
		value, err := capValue(key, settings)
		return value, true, err
	case spec.Merge:
		value, err := mergeValue(key, settings)
		return value, true, err
	default:
		return Value{}, false, fmt.Errorf("%s has the unknown rule %q", key.Name(), key.Rule)
	}
}

func replaceValue(key spec.Key, settings map[spec.Layer]setting) (Value, error) {
	if err := conflict(key.Name(), settings[spec.Managed], settings[spec.Flags]); err != nil {
		return Value{}, err
	}
	var chosen setting
	for i := len(replaceOrder) - 1; i >= 0; i-- {
		if s, ok := settings[replaceOrder[i]]; ok {
			chosen = s
		}
	}
	return Value{Key: key, Raw: chosen.raw, Origins: []Origin{chosen.origin}}, nil
}

// conflict reports a command line value that differs from a managed one.
func conflict(name string, managed, flag setting) error {
	if managed.raw == nil || flag.raw == nil || equal(managed.raw, flag.raw) {
		return nil
	}
	return fmt.Errorf("the command line sets %s to %s, but the managed file %s sets it to %s", name, flag.raw, managed.origin.Path, managed.raw)
}

// equal compares two valid JSON values without their white space.
func equal(a, b json.RawMessage) bool {
	var ca, cb bytes.Buffer
	_ = json.Compact(&ca, a)
	_ = json.Compact(&cb, b)
	return bytes.Equal(ca.Bytes(), cb.Bytes())
}

func appendValue(key spec.Key, settings map[spec.Layer]setting) (Value, error) {
	items := []json.RawMessage{}
	var origins []Origin
	for _, l := range appendOrder {
		s, ok := settings[l]
		if !ok {
			continue
		}
		var layerItems []json.RawMessage
		if err := json.Unmarshal(s.raw, &layerItems); err != nil {
			return Value{}, fmt.Errorf("%s from %s is not an array", key.Name(), s.origin.Layer)
		}
		items = append(items, layerItems...)
		origins = append(origins, s.origin)
	}
	// The items were parsed from JSON, so they always encode.
	raw, _ := json.Marshal(items)
	return Value{Key: key, Raw: raw, Origins: origins}, nil
}

func requireValue(key spec.Key, settings map[spec.Layer]setting) (Value, error) {
	value := Value{Key: key, Raw: json.RawMessage("false")}
	for _, l := range appendOrder {
		s, ok := settings[l]
		if !ok {
			continue
		}
		var demanded bool
		if err := json.Unmarshal(s.raw, &demanded); err != nil {
			return Value{}, fmt.Errorf("%s from %s is not a boolean", key.Name(), s.origin.Layer)
		}
		if demanded {
			value.Raw = json.RawMessage("true")
			value.Origins = append(value.Origins, s.origin)
		}
	}
	return value, nil
}

// limit is one candidate value of a Cap key.
type limit struct {
	setting
	n float64
}

func capValue(key spec.Key, settings map[spec.Layer]setting) (Value, error) {
	limits := map[spec.Layer]limit{}
	for l, s := range settings {
		var n float64
		if err := json.Unmarshal(s.raw, &n); err != nil {
			return Value{}, fmt.Errorf("%s from %s is not a number", key.Name(), s.origin.Layer)
		}
		limits[l] = limit{setting: s, n: n}
	}
	ceiling, _ := json.Marshal(key.Ceiling)
	chosen := limit{setting: setting{origin: Origin{Layer: spec.Default}, raw: ceiling}, n: key.Ceiling}
	if base, ok := first(limits, spec.User, spec.Managed); ok {
		chosen = base
	}
	if flag, ok := limits[spec.Flags]; ok {
		chosen = flag
	} else if project, ok := limits[spec.Project]; ok && project.n < chosen.n {
		chosen = project
	}
	if managed, ok := limits[spec.Managed]; ok && chosen.n > managed.n {
		if chosen.origin.Layer == spec.Flags {
			return Value{}, fmt.Errorf("the command line sets %s to %s, above the ceiling %s of the managed file %s", key.Name(), chosen.raw, managed.raw, managed.origin.Path)
		}
		chosen = managed
	}
	return Value{Key: key, Raw: chosen.raw, Origins: []Origin{chosen.origin}}, nil
}

func first(limits map[spec.Layer]limit, layers ...spec.Layer) (limit, bool) {
	for _, l := range layers {
		if found, ok := limits[l]; ok {
			return found, true
		}
	}
	return limit{}, false
}

func mergeValue(key spec.Key, settings map[spec.Layer]setting) (Value, error) {
	entries := map[spec.Layer]map[string]json.RawMessage{}
	for l, s := range settings {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(s.raw, &object); err != nil || object == nil {
			return Value{}, fmt.Errorf("%s from %s is not an object", key.Name(), s.origin.Layer)
		}
		entries[l] = object
	}
	merged := map[string]json.RawMessage{}
	origins := map[string]Origin{}
	for _, l := range replaceOrder {
		for name, raw := range entries[l] {
			if _, taken := merged[name]; taken {
				continue
			}
			if l == spec.Managed {
				if flagRaw, ok := entries[spec.Flags][name]; ok {
					if err := conflict(key.Name()+"."+name, setting{origin: settings[l].origin, raw: raw}, setting{raw: flagRaw}); err != nil {
						return Value{}, err
					}
				}
			}
			merged[name] = raw
			origins[name] = settings[l].origin
		}
	}
	// The entries were parsed from JSON, so they always encode.
	raw, _ := json.Marshal(merged)
	return Value{Key: key, Raw: raw, Origins: distinct(origins), Entries: origins}, nil
}

// distinct lists the origins of the entries once each, in precedence order.
func distinct(entries map[string]Origin) []Origin {
	seen := map[Origin]bool{}
	var out []Origin
	for _, origin := range entries {
		if !seen[origin] {
			seen[origin] = true
			out = append(out, origin)
		}
	}
	rank := map[spec.Layer]int{}
	for i, l := range replaceOrder {
		rank[l] = i
	}
	sort.Slice(out, func(i, j int) bool { return rank[out[i].Layer] < rank[out[j].Layer] })
	return out
}
