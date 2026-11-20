package runconfig

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	providerssection "github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/providers"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/run"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

// FlagsOrigin names the command line in a source.
const FlagsOrigin = "command line"

// Source says where a value came from: a layer and, for a file, its path.
type Source struct {
	Layer  spec.Layer
	Origin string
}

// String renders the source as "layer origin".
func (s Source) String() string {
	if s.Origin == "" {
		return string(s.Layer)
	}
	return string(s.Layer) + " " + s.Origin
}

func sourceOf(origin merge.Origin) Source {
	if origin.Layer == spec.Flags {
		return Source{Layer: spec.Flags, Origin: FlagsOrigin}
	}
	return Source{Layer: origin.Layer, Origin: origin.Path}
}

// Item is one value of an Append key with its source.
type Item struct {
	Text   string
	Source Source
}

// Resolved is the merged run section with the source of each value.
type Resolved struct {
	merged  merge.Resolved
	sources []Source
}

// Sections lists the sections agent-cli reads.
func Sections() []spec.Section {
	return []spec.Section{run.Section(), providerssection.Section()}
}

// Merge combines the run section of the documents and the command line by the
// rule of each key. It fails closed: a file that sets a key its layer may not
// set and a command line value that contradicts a managed one are errors.
func Merge(documents []layer.Document, flags []merge.Flag) (Resolved, error) {
	merged, err := merge.Merge(Sections(), documents, flags)
	if err != nil {
		return Resolved{}, err
	}
	resolved := Resolved{merged: merged}
	for _, l := range []spec.Layer{spec.User, spec.Project, spec.Flags, spec.Managed} {
		if l == spec.Flags {
			resolved.sources = append(resolved.sources, Source{Layer: spec.Flags, Origin: FlagsOrigin})
		}
		for _, document := range documents {
			if document.Layer == l && setsAny(document) {
				resolved.sources = append(resolved.sources, Source{Layer: l, Origin: document.Path})
			}
		}
	}
	return resolved, nil
}

func setsAny(document layer.Document) bool {
	for _, section := range Sections() {
		if _, ok := document.Sections[section.Name]; ok {
			return true
		}
	}
	return false
}

// Merged returns the merge of every section agent-cli reads, for a report of
// the configuration and its sources.
func (r Resolved) Merged() merge.Resolved { return r.merged }

func (r Resolved) value(key Key) (merge.Value, bool) {
	return r.merged.Source(run.Name, key.pointer())
}

// Has reports whether any layer set key. A Require key counts only when a
// layer demands it.
func (r Resolved) Has(key Key) bool {
	value, ok := r.value(key)
	return ok && (value.Key.Rule != spec.Require || r.On(key))
}

// Text returns the merged value of a Replace key, empty when no layer set it.
func (r Resolved) Text(key Key) string {
	value, ok := r.value(key)
	if !ok {
		return ""
	}
	var text string
	// The schema of the run section makes every Replace key a string.
	_ = json.Unmarshal(value.Raw, &text)
	return text
}

// On returns the merged value of a Require key.
func (r Resolved) On(key Key) bool {
	value, ok := r.value(key)
	if !ok {
		return false
	}
	var on bool
	_ = json.Unmarshal(value.Raw, &on)
	return on
}

// Items returns the merged values of an Append key.
func (r Resolved) Items(key Key) []string {
	sourced := r.Sourced(key)
	if len(sourced) == 0 {
		return nil
	}
	texts := make([]string, len(sourced))
	for i, item := range sourced {
		texts[i] = item.Text
	}
	return texts
}

// Sourced returns the merged values of an Append key, each with the layer that
// supplied it, in the order a consumer applies them. A value that a weaker
// layer already gave keeps that source and appears once.
func (r Resolved) Sourced(key Key) []Item {
	value, ok := r.value(key)
	if !ok {
		return nil
	}
	var texts []string
	// The schema of the run section makes every Append key an array of
	// strings, and the command line gives one.
	_ = json.Unmarshal(value.Raw, &texts)
	var items []Item
	seen := map[string]bool{}
	for i, text := range texts {
		if seen[text] {
			continue
		}
		seen[text] = true
		items = append(items, Item{Text: text, Source: sourceOf(value.Items[i])})
	}
	return items
}

// Sources lists the layers that took part in the merge, in precedence order
// from the lowest layer up. The command line always takes part.
func (r Resolved) Sources() []Source { return slices.Clone(r.sources) }

// strength ranks the layers of a Replace key, the winner first.
var strength = []spec.Layer{spec.Managed, spec.Flags, spec.Project, spec.User}

// demander returns the strongest layer that demands a Require value.
func demander(value merge.Value) Source {
	best := value.Origins[0]
	for _, origin := range value.Origins[1:] {
		if slices.Index(strength, origin.Layer) < slices.Index(strength, best.Layer) {
			best = origin
		}
	}
	return sourceOf(best)
}

// Provenance describes, one line per value, which layer decided it. A line
// reads "key: value (layer origin)".
func (r Resolved) Provenance() []string {
	var lines []string
	for _, declared := range run.Section().Keys {
		key := keyOf(declared)
		value, ok := r.value(key)
		if !ok {
			continue
		}
		switch declared.Rule {
		case spec.Replace:
			lines = append(lines, fmt.Sprintf("%s: %s (%s)", key, r.Text(key), sourceOf(value.Origins[0])))
		case spec.Require:
			if r.On(key) {
				lines = append(lines, fmt.Sprintf("%s: required (%s)", key, demander(value)))
			}
		default:
			for _, item := range r.Sourced(key) {
				lines = append(lines, fmt.Sprintf("%s: %s (%s)", key, item.Text, item.Source))
			}
		}
	}
	return lines
}

// Layers returns, for each key a layer set, the layers that decided its value.
// A run record keeps it in place of the values, which can hold a locator or a
// path.
func (r Resolved) Layers() map[Key][]spec.Layer {
	layers := map[Key][]spec.Layer{}
	for _, declared := range run.Section().Keys {
		key := keyOf(declared)
		value, ok := r.value(key)
		if !ok {
			continue
		}
		switch declared.Rule {
		case spec.Replace:
			layers[key] = []spec.Layer{value.Origins[0].Layer}
		case spec.Require:
			if r.On(key) {
				layers[key] = []spec.Layer{demander(value).Layer}
			}
		default:
			for _, item := range r.Sourced(key) {
				if !slices.Contains(layers[key], item.Source.Layer) {
					layers[key] = append(layers[key], item.Source.Layer)
				}
			}
		}
	}
	return layers
}
