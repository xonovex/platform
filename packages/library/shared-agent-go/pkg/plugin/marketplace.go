package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// MarketplacePaths lists the files a marketplace root may hold: the marketplace
// file of every harness that declares one, in the order of the harness names.
func MarketplacePaths() []string {
	return harnessPaths(func(layout types.PluginLayout) string { return layout.Marketplace })
}

// Entry is one plugin a marketplace lists.
type Entry struct {
	Name        string
	Description string
	// Source is the relative path of the plugin below the marketplace root, as
	// the file wrote it, or empty when the file gave a source of another form.
	Source string
}

// Path returns the directory of the plugin below the marketplace root. It fails
// closed on a source that is not a relative path inside the root.
func (e Entry) Path() (string, error) {
	relative, ok := strings.CutPrefix(e.Source, "./")
	if e.Source == "" || !ok {
		return "", fmt.Errorf("plugin %q has a source that is not a path starting with ./", e.Name)
	}
	if err := artifact.ConfinedPath(strings.TrimSuffix(relative, "/")); err != nil {
		return "", fmt.Errorf("plugin %q: %w", e.Name, err)
	}
	return strings.TrimSuffix(relative, "/"), nil
}

// Marketplace is the plugin listing of a marketplace root.
type Marketplace struct {
	Name    string
	Entries []Entry
}

// Find returns the entry of the named plugin.
func (m Marketplace) Find(name string) (Entry, bool) {
	for _, entry := range m.Entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return Entry{}, false
}

type marketplaceFile struct {
	Name    string `json:"name"`
	Plugins []struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Source      json.RawMessage `json:"source"`
	} `json:"plugins"`
}

// ReadMarketplace reads the listing out of the marketplace files of a root. When
// both formats are present they must agree on the name and on the source of every
// plugin they share; the result lists the plugins of both, sorted by name.
func ReadMarketplace(tree artifact.Tree) (Marketplace, error) {
	var result Marketplace
	found := ""
	byName := map[string]Entry{}
	for _, path := range MarketplacePaths() {
		file, ok := tree.Get(path)
		if !ok {
			continue
		}
		var parsed marketplaceFile
		if err := json.Unmarshal(file.Content, &parsed); err != nil {
			return Marketplace{}, fmt.Errorf("%s: %w", path, err)
		}
		if !artifact.ValidName(parsed.Name) {
			return Marketplace{}, fmt.Errorf("%s: the marketplace name %q is not a kebab-case name", path, parsed.Name)
		}
		if found != "" && result.Name != parsed.Name {
			return Marketplace{}, fmt.Errorf("%s and %s name different marketplaces", found, path)
		}
		result.Name, found = parsed.Name, path
		for _, plugin := range parsed.Plugins {
			entry := Entry{Name: plugin.Name, Description: plugin.Description}
			_ = json.Unmarshal(plugin.Source, &entry.Source)
			if !artifact.ValidName(entry.Name) {
				return Marketplace{}, fmt.Errorf("%s: the plugin name %q is not a kebab-case name", path, entry.Name)
			}
			if other, seen := byName[entry.Name]; seen && other.Source != entry.Source {
				return Marketplace{}, fmt.Errorf("%s: plugin %q has another source than in the other marketplace file", path, entry.Name)
			}
			byName[entry.Name] = entry
		}
	}
	if found == "" {
		return Marketplace{}, errors.New("the source has no marketplace file: expected " + strings.Join(MarketplacePaths(), " or "))
	}
	for _, entry := range byName {
		result.Entries = append(result.Entries, entry)
	}
	slices.SortFunc(result.Entries, func(a, b Entry) int { return strings.Compare(a.Name, b.Name) })
	return result, nil
}
