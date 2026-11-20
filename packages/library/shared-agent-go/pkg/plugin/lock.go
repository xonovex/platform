package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

// LockVersion is the version of the lock file format.
const LockVersion = 1

// Locked is what the lock file records of one installed plugin: where it came
// from, the digest of the content the user approved and the capabilities the user
// granted.
type Locked struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Source       string   `json:"source"`
	Marketplace  string   `json:"marketplace,omitempty"`
	Resolved     string   `json:"resolved,omitempty"`
	Digest       string   `json:"digest"`
	Grants       []string `json:"grants,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// Lock is the set of approved plugins, sorted by name.
type Lock struct {
	Version int      `json:"version"`
	Plugins []Locked `json:"plugins"`
}

// EmptyLock returns a lock with no plugin.
func EmptyLock() Lock { return Lock{Version: LockVersion} }

// Find returns the record of the named plugin.
func (l Lock) Find(name string) (Locked, bool) {
	for _, plugin := range l.Plugins {
		if plugin.Name == name {
			return plugin, true
		}
	}
	return Locked{}, false
}

// With returns a lock that holds record in place of any record of its name.
func (l Lock) With(record Locked) Lock {
	plugins := slices.DeleteFunc(slices.Clone(l.Plugins), func(p Locked) bool { return p.Name == record.Name })
	plugins = append(plugins, record)
	slices.SortFunc(plugins, func(a, b Locked) int { return strings.Compare(a.Name, b.Name) })
	return Lock{Version: LockVersion, Plugins: plugins}
}

// Without returns a lock with the named plugin removed, and whether it was there.
func (l Lock) Without(name string) (Lock, bool) {
	if _, found := l.Find(name); !found {
		return l, false
	}
	return Lock{Version: LockVersion, Plugins: slices.DeleteFunc(slices.Clone(l.Plugins), func(p Locked) bool { return p.Name == name })}, true
}

// Dependents lists the plugins that name the plugin as a dependency.
func (l Lock) Dependents(name string) []string {
	var dependents []string
	for _, plugin := range l.Plugins {
		if slices.Contains(plugin.Dependencies, name) {
			dependents = append(dependents, plugin.Name)
		}
	}
	return dependents
}

// Closure returns the named plugins and everything they depend on, each once,
// with a plugin after its dependencies. It fails closed on a plugin or a
// dependency that is not locked.
func (l Lock) Closure(names []string) ([]Locked, error) {
	var ordered []Locked
	state := map[string]int{}
	var visit func(name, from string) error
	visit = func(name, from string) error {
		record, found := l.Find(name)
		if !found {
			if from == "" {
				return fmt.Errorf("plugin %q is not installed", name)
			}
			return fmt.Errorf("plugin %q depends on %q, which is not installed", from, name)
		}
		switch state[name] {
		case 1:
			return fmt.Errorf("plugin %q depends on itself through %q", name, from)
		case 2:
			return nil
		}
		state[name] = 1
		for _, dependency := range record.Dependencies {
			if err := visit(dependency, name); err != nil {
				return err
			}
		}
		state[name] = 2
		ordered = append(ordered, record)
		return nil
	}
	for _, name := range names {
		if err := visit(name, ""); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// EncodeLock returns the lock file bytes: indented JSON with a final newline.
func EncodeLock(lock Lock) ([]byte, error) {
	if err := validateLock(lock); err != nil {
		return nil, err
	}
	if lock.Plugins == nil {
		lock.Plugins = []Locked{}
	}
	return indent(lock), nil
}

// DecodeLock reads lock file bytes and fails closed on an unknown key, a version
// other than the one this code writes, an invalid record or a name that repeats.
func DecodeLock(data []byte) (Lock, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var lock Lock
	if err := decoder.Decode(&lock); err != nil {
		return Lock{}, fmt.Errorf("lock file: %w", err)
	}
	if decoder.More() {
		return Lock{}, errors.New("lock file: trailing data after the JSON document")
	}
	if err := validateLock(lock); err != nil {
		return Lock{}, err
	}
	return lock, nil
}

func validateLock(lock Lock) error {
	if lock.Version != LockVersion {
		return fmt.Errorf("lock file: version %d is not supported; expected %d", lock.Version, LockVersion)
	}
	seen := map[string]bool{}
	for _, plugin := range lock.Plugins {
		switch {
		case !artifact.ValidName(plugin.Name):
			return fmt.Errorf("lock file: the plugin name %q is not a kebab-case name", plugin.Name)
		case seen[plugin.Name]:
			return fmt.Errorf("lock file: plugin %q is listed twice", plugin.Name)
		case !artifact.ValidDigest(plugin.Digest):
			return fmt.Errorf("lock file: plugin %q has an invalid digest", plugin.Name)
		case plugin.Source == "":
			return fmt.Errorf("lock file: plugin %q has no source", plugin.Name)
		}
		seen[plugin.Name] = true
		for _, grant := range plugin.Grants {
			if _, err := ParseCapability(grant); err != nil {
				return fmt.Errorf("lock file: plugin %q: %w", plugin.Name, err)
			}
		}
		for _, dependency := range plugin.Dependencies {
			if !artifact.ValidName(dependency) {
				return fmt.Errorf("lock file: plugin %q has the invalid dependency %q", plugin.Name, dependency)
			}
		}
	}
	return nil
}

// Source is a marketplace the user registered.
type Source struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// Registry is the set of registered marketplaces, sorted by name.
type Registry struct {
	Version      int      `json:"version"`
	Marketplaces []Source `json:"marketplaces"`
}

// EmptyRegistry returns a registry with no marketplace.
func EmptyRegistry() Registry { return Registry{Version: LockVersion} }

// Find returns the marketplace of the name.
func (r Registry) Find(name string) (Source, bool) {
	for _, source := range r.Marketplaces {
		if source.Name == name {
			return source, true
		}
	}
	return Source{}, false
}

// With returns a registry that holds source in place of any of its name.
func (r Registry) With(source Source) Registry {
	marketplaces := slices.DeleteFunc(slices.Clone(r.Marketplaces), func(s Source) bool { return s.Name == source.Name })
	marketplaces = append(marketplaces, source)
	slices.SortFunc(marketplaces, func(a, b Source) int { return strings.Compare(a.Name, b.Name) })
	return Registry{Version: LockVersion, Marketplaces: marketplaces}
}

// Without returns a registry with the named marketplace removed, and whether it
// was there.
func (r Registry) Without(name string) (Registry, bool) {
	if _, found := r.Find(name); !found {
		return r, false
	}
	return Registry{Version: LockVersion, Marketplaces: slices.DeleteFunc(slices.Clone(r.Marketplaces), func(s Source) bool { return s.Name == name })}, true
}

// EncodeRegistry returns the registry file bytes.
func EncodeRegistry(registry Registry) ([]byte, error) {
	if err := validateRegistry(registry); err != nil {
		return nil, err
	}
	if registry.Marketplaces == nil {
		registry.Marketplaces = []Source{}
	}
	return indent(registry), nil
}

// DecodeRegistry reads registry file bytes and fails closed on an unknown key or
// an invalid record.
func DecodeRegistry(data []byte) (Registry, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var registry Registry
	if err := decoder.Decode(&registry); err != nil {
		return Registry{}, fmt.Errorf("marketplace registry: %w", err)
	}
	if decoder.More() {
		return Registry{}, errors.New("marketplace registry: trailing data after the JSON document")
	}
	if err := validateRegistry(registry); err != nil {
		return Registry{}, err
	}
	return registry, nil
}

func validateRegistry(registry Registry) error {
	if registry.Version != LockVersion {
		return fmt.Errorf("marketplace registry: version %d is not supported; expected %d", registry.Version, LockVersion)
	}
	seen := map[string]bool{}
	for _, source := range registry.Marketplaces {
		switch {
		case !artifact.ValidName(source.Name):
			return fmt.Errorf("marketplace registry: the name %q is not a kebab-case name", source.Name)
		case seen[source.Name]:
			return fmt.Errorf("marketplace registry: %q is listed twice", source.Name)
		case source.Source == "":
			return fmt.Errorf("marketplace registry: %q has no source", source.Name)
		}
		seen[source.Name] = true
	}
	return nil
}

// indent encodes value as indented JSON with a final newline. The values are
// plain strings, numbers and lists, which cannot fail to encode.
func indent(value any) []byte {
	data, _ := json.MarshalIndent(value, "", "  ")
	return append(data, '\n')
}
