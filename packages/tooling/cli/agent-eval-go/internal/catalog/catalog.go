// Package catalog discovers the plugins and skills an eval can offer a model.
// It reads plugin manifests through the file-system port, accepts only plugins
// that hold skills and nothing executable, and resolves a plugin's dependencies
// from its sibling plugins.
package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// ManifestPath is where a plugin keeps its manifest, relative to the plugin.
const ManifestPath = ".claude-plugin/plugin.json"

// executable lists what a skill-only plugin must not contain: anything that runs
// code or changes settings on the machine that loads the plugin.
var executable = []string{"commands", "agents", "hooks", ".mcp.json", ".lsp.json", "settings.json", ".claude/settings.json"}

// allowed lists the manifest members of a skill-only plugin.
var allowed = map[string]bool{"author": true, "dependencies": true, "description": true, "name": true, "skills": true, "version": true}

// Skill is one skill of a plugin.
type Skill struct {
	// Name is the skill directory name.
	Name   string
	Plugin string
	Dir    string
	// TriggerFile and OutputFile are the eval documents of the skill, empty when
	// the skill has none.
	TriggerFile string
	OutputFile  string
}

// Plugin is one skill plugin.
type Plugin struct {
	Name         string
	Dir          string
	Dependencies []string
	Skills       []Skill
}

// Catalog is the set of plugins found below the project roots.
type Catalog struct {
	Plugins []Plugin
}

// Load discovers the plugins below the roots. A root is a plugin directory or a
// directory whose children are plugin directories.
func Load(fsys fs.FileSystem, roots []string) (Catalog, error) {
	var cat Catalog
	seen := map[string]string{}
	add := func(dir string) error {
		plugin, err := readPlugin(fsys, dir)
		if err != nil {
			return err
		}
		if other, dup := seen[plugin.Name]; dup {
			return fmt.Errorf("plugin %q is in both %s and %s", plugin.Name, other, dir)
		}
		seen[plugin.Name] = dir
		cat.Plugins = append(cat.Plugins, plugin)
		return nil
	}
	for _, root := range roots {
		isPlugin, err := fs.Exists(fsys, path.Join(root, ManifestPath))
		if err != nil {
			return Catalog{}, err
		}
		if isPlugin {
			if err := add(root); err != nil {
				return Catalog{}, err
			}
			continue
		}
		entries, err := fsys.ReadDir(root)
		if err != nil {
			return Catalog{}, fmt.Errorf("plugin root %s: %w", root, err)
		}
		for _, entry := range entries {
			dir := path.Join(root, entry.Name)
			if ok, err := fs.Exists(fsys, path.Join(dir, ManifestPath)); err != nil {
				return Catalog{}, err
			} else if entry.IsDir && ok {
				if err := add(dir); err != nil {
					return Catalog{}, err
				}
			}
		}
	}
	sort.Slice(cat.Plugins, func(i, j int) bool { return cat.Plugins[i].Name < cat.Plugins[j].Name })
	return cat, nil
}

func readPlugin(fsys fs.FileSystem, dir string) (Plugin, error) {
	file := path.Join(dir, ManifestPath)
	data, err := fsys.ReadFile(file)
	if err != nil {
		return Plugin{}, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Plugin{}, fmt.Errorf("invalid plugin manifest %s: %w", file, err)
	}
	for field := range raw {
		if !allowed[field] {
			return Plugin{}, fmt.Errorf("plugin %s is not skill-only: its manifest has %q", dir, field)
		}
	}
	var manifest struct {
		Name         string   `json:"name"`
		Dependencies []string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil || manifest.Name == "" {
		return Plugin{}, fmt.Errorf("plugin manifest %s needs a name and string dependencies", file)
	}
	for _, name := range executable {
		if ok, err := fs.Exists(fsys, path.Join(dir, name)); err != nil {
			return Plugin{}, err
		} else if ok {
			return Plugin{}, fmt.Errorf("plugin %s is not skill-only: it contains %s", dir, name)
		}
	}
	skills, err := readSkills(fsys, manifest.Name, dir)
	if err != nil {
		return Plugin{}, err
	}
	return Plugin{Name: manifest.Name, Dir: dir, Dependencies: manifest.Dependencies, Skills: skills}, nil
}

func readSkills(fsys fs.FileSystem, plugin, dir string) ([]Skill, error) {
	root := path.Join(dir, "skills")
	entries, err := fsys.ReadDir(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var skills []Skill
	for _, entry := range entries {
		skillDir := path.Join(root, entry.Name)
		if ok, err := fs.Exists(fsys, path.Join(skillDir, "SKILL.md")); err != nil {
			return nil, err
		} else if !entry.IsDir || !ok {
			continue
		}
		skill := Skill{Name: entry.Name, Plugin: plugin, Dir: skillDir}
		if ok, _ := fs.Exists(fsys, path.Join(skillDir, model.TriggerFile)); ok {
			skill.TriggerFile = path.Join(skillDir, model.TriggerFile)
		}
		if ok, _ := fs.Exists(fsys, path.Join(skillDir, model.OutputFile)); ok {
			skill.OutputFile = path.Join(skillDir, model.OutputFile)
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

// Plugin returns the plugin with the name.
func (c Catalog) Plugin(name string) (Plugin, bool) {
	for _, p := range c.Plugins {
		if p.Name == name {
			return p, true
		}
	}
	return Plugin{}, false
}

// Skill returns the skill with the short name. Two skills with one name in
// different plugins are ambiguous: the name must be qualified as plugin:skill.
func (c Catalog) Skill(name string) (Skill, error) {
	var found []Skill
	for _, p := range c.Plugins {
		for _, s := range p.Skills {
			if s.Name == name || p.Name+":"+s.Name == name {
				found = append(found, s)
			}
		}
	}
	switch len(found) {
	case 0:
		return Skill{}, fmt.Errorf("skill %q not found", name)
	case 1:
		return found[0], nil
	}
	return Skill{}, fmt.Errorf("skill %q is in %d plugins: qualify it as plugin:skill", name, len(found))
}

// Closure returns the plugin and the plugins it depends on, dependencies first.
// A dependency the catalog lacks, or a cycle, is an error.
func (c Catalog) Closure(name string) ([]Plugin, error) {
	var ordered []Plugin
	state := map[string]int{}
	var visit func(name, from string) error
	visit = func(name, from string) error {
		p, ok := c.Plugin(name)
		if !ok {
			return fmt.Errorf("local plugin dependency not found: %s -> %s", from, name)
		}
		switch state[name] {
		case 1:
			return fmt.Errorf("plugin dependency cycle at %s", name)
		case 2:
			return nil
		}
		state[name] = 1
		for _, dep := range p.Dependencies {
			if err := visit(dep, name); err != nil {
				return err
			}
		}
		state[name] = 2
		ordered = append(ordered, p)
		return nil
	}
	if err := visit(name, name); err != nil {
		return nil, err
	}
	return ordered, nil
}
