package providers

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// Getenv resolves an environment variable and returns the empty string when it is unset.
type Getenv func(key string) string

// ErrNotDefined reports that no loaded document defines the provider.
var ErrNotDefined = errors.New("provider is not defined")

// ErrLocalOnly reports that a provider is not marked portable.
var ErrLocalOnly = errors.New("provider is local-only")

// Document is one provider document with the layer and the file it came from.
type Document struct {
	Preset Preset
	// Layer names the configuration layer that holds the file, such as user or
	// managed.
	Layer string
	// Origin names the file the document was read from.
	Origin string
	// Raw holds the bytes of the file, which an approval digest covers.
	Raw []byte
}

// Override replaces the id of one model slot of the provider a caller resolves.
type Override struct {
	Slot string
	ID   string
	// Source names the layer or the flag that decided the id.
	Source string
}

// Entry describes one defined provider for a listing.
type Entry struct {
	Name        string
	DisplayName string
	Agent       types.AgentType
	Extends     string
	Portable    bool
	Layer       string
	Origin      string
}

type key struct {
	agent types.AgentType
	name  string
}

// Catalog holds the provider documents a caller loaded. The zero value is an
// empty catalog.
type Catalog struct {
	documents map[key]Document
}

// NewCatalog builds a catalog from documents in ascending precedence: a later
// document replaces an earlier one that has the same agent and name.
func NewCatalog(documents ...Document) Catalog {
	catalog := Catalog{documents: make(map[key]Document, len(documents))}
	for _, document := range documents {
		catalog.documents[key{document.Preset.Agent, document.Preset.Name}] = document
	}
	return catalog
}

// Names returns the names of the providers defined for an agent type, sorted.
func (c Catalog) Names(agent types.AgentType) []string {
	var names []string
	for k := range c.documents {
		if k.agent == agent {
			names = append(names, k.name)
		}
	}
	slices.Sort(names)
	return names
}

// List returns every defined provider sorted by agent and name.
func (c Catalog) List() []Entry {
	entries := make([]Entry, 0, len(c.documents))
	for _, document := range c.documents {
		p := document.Preset
		entries = append(entries, Entry{
			Name: p.Name, DisplayName: p.DisplayName, Agent: p.Agent, Extends: p.Extends,
			Portable: p.Portable != nil && *p.Portable, Layer: document.Layer, Origin: document.Origin,
		})
	}
	slices.SortFunc(entries, func(a, b Entry) int {
		if a.Agent != b.Agent {
			return compareStrings(string(a.Agent), string(b.Agent))
		}
		return compareStrings(a.Name, b.Name)
	})
	return entries
}

func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// Resolve returns the provider of the given name and agent type with the
// overrides applied to its model slots and every placeholder expanded. It wraps
// ErrNotDefined when no document defines the provider.
func (c Catalog) Resolve(name string, agent types.AgentType, overrides ...Override) (*types.ModelProvider, error) {
	document, ok := c.documents[key{agent, name}]
	if !ok {
		return nil, fmt.Errorf("provider %q for agent %q: %w", name, agent, ErrNotDefined)
	}
	flat, sources, err := c.flatten(document, map[key]bool{})
	if err != nil {
		return nil, err
	}
	for _, override := range overrides {
		slot, defined := flat.Models[override.Slot]
		if !defined {
			return nil, fmt.Errorf("provider %q has no model slot %q; it defines %s", name, override.Slot, slotNames(flat.Models))
		}
		next := Slot{ID: override.ID, Effort: slot.Effort}
		if err := next.validate(); err != nil {
			return nil, fmt.Errorf("provider %q: model slot %q: %w", name, override.Slot, err)
		}
		flat.Models = overlay(flat.Models, map[string]Slot{override.Slot: next})
		sources[override.Slot] = override.Source
	}
	if err := flat.checkPlaceholders(); err != nil {
		return nil, err
	}
	return build(flat, sources)
}

// ResolvePortable is Resolve for a provider that is safe in a remote workload.
func (c Catalog) ResolvePortable(name string, agent types.AgentType, overrides ...Override) (*types.ModelProvider, error) {
	provider, err := c.Resolve(name, agent, overrides...)
	if err != nil {
		return nil, err
	}
	if !provider.Portable {
		return nil, fmt.Errorf("provider %q for agent %q: %w", name, agent, ErrLocalOnly)
	}
	return provider, nil
}

// Layers returns the layers of the documents that decide a provider: the
// document itself and each document it extends, nearest first. It is empty when
// the provider is not defined or its chain is broken.
func (c Catalog) Layers(name string, agent types.AgentType) []string {
	var layers []string
	seen := map[key]bool{}
	for current, ok := name, true; ok; {
		k := key{agent, current}
		document, found := c.documents[k]
		if !found || seen[k] {
			return layers
		}
		seen[k] = true
		layers = append(layers, document.Layer)
		current, ok = document.Preset.Extends, document.Preset.Extends != ""
	}
	return layers
}

// flatten applies the extends chain of a document and records which document
// decided each model slot.
func (c Catalog) flatten(document Document, seen map[key]bool) (Preset, map[string]string, error) {
	p := document.Preset
	k := key{p.Agent, p.Name}
	if seen[k] {
		return Preset{}, nil, fmt.Errorf("provider %q extends itself through a cycle", p.Name)
	}
	seen[k] = true
	sources := make(map[string]string, len(p.Models))
	if p.Extends == "" {
		for slot := range p.Models {
			sources[slot] = document.Origin
		}
		return p, sources, nil
	}
	parent, ok := c.documents[key{p.Agent, p.Extends}]
	if !ok {
		return Preset{}, nil, fmt.Errorf("provider %q extends %q, which is not defined for agent %q", p.Name, p.Extends, p.Agent)
	}
	base, sources, err := c.flatten(parent, seen)
	if err != nil {
		return Preset{}, nil, err
	}
	for slot := range p.Models {
		sources[slot] = document.Origin
	}
	return refine(base, p), sources, nil
}

func slotNames(slots map[string]Slot) string {
	if len(slots) == 0 {
		return "no slot"
	}
	return fmt.Sprintf("%q", slices.Sorted(maps.Keys(slots)))
}

func build(p Preset, sources map[string]string) (*types.ModelProvider, error) {
	provider := &types.ModelProvider{
		Name:        p.Name,
		DisplayName: p.DisplayName,
		AgentType:   p.Agent,
		Portable:    p.Portable != nil && *p.Portable,
	}
	if provider.DisplayName == "" {
		provider.DisplayName = p.Name
	}
	if p.Credential != nil {
		provider.CredentialSourceEnv = p.Credential.SourceEnv
		provider.CredentialTargetEnv = p.Credential.TargetEnv
	}
	for _, name := range slices.Sorted(maps.Keys(p.Models)) {
		slot := p.Models[name]
		provider.Slots = append(provider.Slots, types.ModelSlot{Name: name, ID: slot.ID, Effort: slot.Effort, Source: sources[name]})
	}
	if len(p.Env) > 0 {
		provider.Environment = make(map[string]string, len(p.Env))
		for k, v := range p.Env {
			expanded, err := expand(v, p.Models)
			if err != nil {
				return nil, fmt.Errorf("provider %q: environment %s: %w", p.Name, k, err)
			}
			provider.Environment[k] = expanded
		}
	}
	for index, arg := range p.Args {
		expanded, err := expand(arg, p.Models)
		if err != nil {
			return nil, fmt.Errorf("provider %q: args[%d]: %w", p.Name, index, err)
		}
		provider.CliArgs = append(provider.CliArgs, expanded)
	}
	return provider, nil
}

// GetProviderCliArgs gets CLI arguments for a provider
func GetProviderCliArgs(provider *types.ModelProvider) []string {
	return slices.Clone(provider.CliArgs)
}
