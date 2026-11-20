// Package plugin is the pure model of an installable plugin: its manifest, the
// capabilities that need an explicit grant, the two marketplace formats, and the
// lock file that records what the user approved. It parses bytes it is given and
// never touches the machine.
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// NeutralManifestPath is the manifest a plugin carries for no harness in
// particular.
const NeutralManifestPath = "plugin.json"

// ManifestPaths lists the manifest locations of a plugin: the manifest of every
// harness that declares one, in the order of the harness names, then the
// neutral manifest. A plugin carries one or more, and every one that is present
// must agree on the name, the version and the dependencies.
func ManifestPaths() []string {
	return append(harnessPaths(func(layout types.PluginLayout) string { return layout.Manifest }), NeutralManifestPath)
}

// harnessPaths lists one path of the plugin layout of every harness that
// declares it, in the order of the harness names.
func harnessPaths(pick func(types.PluginLayout) string) []string {
	facts := agents.Embedded()
	var paths []string
	for _, agentType := range facts.Types() {
		if p := pick(facts.PluginLayout(agentType)); p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// Manifest is what the manifests of a plugin agree on.
type Manifest struct {
	Name         string
	Version      string
	Description  string
	Dependencies []string
	// Keys are the top-level keys of every manifest, for the capability check.
	Keys []string
}

type manifestFile struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
}

// ReadManifest reads the manifests of a plugin out of its files and fails closed
// when none exists, one is not valid JSON, the name is not a kebab-case name, the
// version is empty, a dependency is invalid, repeated or the plugin itself, or
// two manifests disagree.
func ReadManifest(tree artifact.Tree) (Manifest, error) {
	var manifest Manifest
	keys := map[string]bool{}
	found := ""
	for _, path := range ManifestPaths() {
		file, ok := tree.Get(path)
		if !ok {
			continue
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(file.Content, &raw); err != nil {
			return Manifest{}, fmt.Errorf("%s: %w", path, err)
		}
		var parsed manifestFile
		if err := json.Unmarshal(file.Content, &parsed); err != nil {
			return Manifest{}, fmt.Errorf("%s: %w", path, err)
		}
		for key := range raw {
			keys[key] = true
		}
		slices.Sort(parsed.Dependencies)
		current := Manifest{Name: parsed.Name, Version: parsed.Version, Description: parsed.Description, Dependencies: parsed.Dependencies}
		if found == "" {
			manifest, found = current, path
			continue
		}
		if manifest.Name != current.Name || manifest.Version != current.Version || !slices.Equal(manifest.Dependencies, current.Dependencies) {
			return Manifest{}, fmt.Errorf("%s and %s disagree on the name, the version or the dependencies", found, path)
		}
	}
	if found == "" {
		return Manifest{}, errors.New("the plugin has no manifest: expected .claude-plugin/plugin.json, .codex-plugin/plugin.json or plugin.json")
	}
	if err := check(manifest); err != nil {
		return Manifest{}, fmt.Errorf("%s: %w", found, err)
	}
	manifest.Keys = slices.Sorted(maps.Keys(keys))
	return manifest, nil
}

func check(manifest Manifest) error {
	if !artifact.ValidName(manifest.Name) {
		return fmt.Errorf("the plugin name %q is not a kebab-case name", manifest.Name)
	}
	if manifest.Version == "" {
		return errors.New("the plugin has no version")
	}
	for index, dependency := range manifest.Dependencies {
		switch {
		case !artifact.ValidName(dependency):
			return fmt.Errorf("the dependency %q is not a kebab-case name", dependency)
		case dependency == manifest.Name:
			return errors.New("the plugin depends on itself")
		case index > 0 && manifest.Dependencies[index-1] == dependency:
			return fmt.Errorf("the dependency %q is repeated", dependency)
		}
	}
	return nil
}
