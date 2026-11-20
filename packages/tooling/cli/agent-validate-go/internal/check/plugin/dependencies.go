package plugin

import (
	"path"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// handoffs returns each skill name that a SKILL.md or a reference of the
// plugin sets in bold, the one written form of a handoff to another skill.
func (r *run) handoffs(p *pkg) (map[string]bool, error) {
	names := map[string]bool{}
	for _, dir := range p.skills {
		files := []string{path.Join(dir, "SKILL.md")}
		entries, err := r.ctx.Tree.List(path.Join(dir, "references"))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir && strings.HasSuffix(e.Name, ".md") {
				files = append(files, path.Join(dir, "references", e.Name))
			}
		}
		for _, file := range files {
			text, err := r.ctx.Tree.Read(file)
			if err != nil {
				return nil, err
			}
			for _, name := range markdown.BoldSkillNames(text) {
				names[name] = true
			}
		}
	}
	return names, nil
}

// A plugin installs another plugin only because one of its skills hands work
// to a skill of that plugin, so each declared dependency needs a bold handoff.
func (r *run) checkUsage(p *pkg) error {
	handoffs, err := r.handoffs(p)
	if err != nil {
		return err
	}
	unused := func(plugin string) bool {
		dependency, ok := r.byName[plugin]
		if !ok {
			dir := path.Join(path.Dir(p.dir), r.options.DirectoryPrefix+strings.TrimPrefix(plugin, r.options.PluginPrefix))
			skills, _ := listSkills(r, dir)
			for _, s := range skills {
				if handoffs[s] {
					return false
				}
			}
			return true
		}
		for _, s := range dependency.skills {
			if handoffs[path.Base(s)] {
				return false
			}
		}
		return true
	}
	const reason = "is declared but no skill of this plugin hands off to one of its skills in bold"
	for _, dep := range p.claude.Dependencies {
		if strings.HasPrefix(dep, r.options.PluginPrefix) && unused(dep) {
			r.report.Fail("plugin.dependency-unused", p.claude.file, 1, "plugin dependency '%s' %s", dep, reason)
		}
	}
	if p.pkgJSON == nil {
		return nil
	}
	for _, name := range sortedKeys(p.pkgJSON.Dependencies) {
		if strings.HasPrefix(name, r.options.PackagePrefix) {
			plugin := r.options.PluginPrefix + strings.TrimPrefix(name, r.options.PackagePrefix)
			if unused(plugin) {
				r.report.Fail("plugin.package-dependency-unused", p.pkgJSON.file, 1, "package dependency '%s' %s", name, reason)
			}
		}
	}
	return nil
}

func listSkills(r *run, dir string) ([]string, error) {
	entries, err := r.ctx.Tree.List(path.Join(dir, "skills"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir && r.ctx.Tree.IsFile(path.Join(dir, "skills", e.Name, "SKILL.md")) {
			out = append(out, e.Name)
		}
	}
	return out, nil
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// npmName is the published name of a plugin.
func (r *run) npmName(plugin string) string {
	return r.options.PackagePrefix + strings.TrimPrefix(plugin, r.options.PluginPrefix)
}

// checkDependencies compares the dependencies of the manifests with the ones of
// package.json, pins each exactly, and looks for cycles.
func (r *run) checkDependencies() {
	graph := map[string][]string{}
	for _, p := range r.plugins {
		if p.claude == nil || p.codex == nil || p.claude.Name != p.codex.Name {
			continue
		}
		graph[p.pluginName] = p.codex.Dependencies
		declared := map[string]string{}
		if p.pkgJSON != nil {
			declared = p.pkgJSON.Dependencies
		}
		for _, npm := range sortedKeys(declared) {
			if !strings.HasPrefix(npm, r.options.PackagePrefix) {
				continue
			}
			plugin := r.options.PluginPrefix + strings.TrimPrefix(npm, r.options.PackagePrefix)
			if !contains(p.codex.Dependencies, plugin) {
				r.report.Fail("plugin.dependency-undeclared", p.pkgJSON.file, 1, "%s declares %s in package.json but omits %s from its plugin manifests", p.pluginName, npm, plugin)
			}
		}
		for _, dep := range p.codex.Dependencies {
			target, ok := r.byName[dep]
			if !ok {
				r.report.Fail("plugin.dependency-missing", p.codex.file, 1, "%s depends on missing %s", p.pluginName, dep)
				continue
			}
			version, pinned := declared[r.npmName(dep)]
			switch {
			case !pinned:
				r.report.Fail("plugin.dependency-unpinned", p.dir, 1, "%s declares %s in its plugin manifests but omits %s from package.json", p.pluginName, dep, r.npmName(dep))
			case target.pkgJSON == nil || target.pkgJSON.Version == "":
				r.report.Fail("plugin.dependency-version", target.dir, 1, "%s/package.json needs a string version", target.dir)
			case version != target.pkgJSON.Version:
				r.report.Fail("plugin.dependency-pin", p.pkgJSON.file, 1, "%s pins %s@%s; expected exact installed version %s", p.pluginName, r.npmName(dep), version, target.pkgJSON.Version)
			}
		}
	}
	for _, cycle := range cycles(graph) {
		r.report.Fail("plugin.dependency-cycle", "", 0, "dependency cycle %s", cycle)
	}
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// cycles returns each dependency cycle once, named from its smallest rotation.
func cycles(graph map[string][]string) []string {
	state := map[string]int{}
	found := map[string]bool{}
	var visit func(name string, stack []string)
	visit = func(name string, stack []string) {
		switch state[name] {
		case 2:
			return
		case 1:
			start := 0
			for i, n := range stack {
				if n == name {
					start = i
				}
			}
			found[canonical(stack[start:])] = true
			return
		}
		state[name] = 1
		deps := append([]string(nil), graph[name]...)
		sort.Strings(deps)
		for _, dep := range deps {
			if _, known := graph[dep]; known {
				visit(dep, append(stack, name))
			}
		}
		state[name] = 2
	}
	names := make([]string, 0, len(graph))
	for name := range graph {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		visit(name, nil)
	}
	out := make([]string, 0, len(found))
	for c := range found {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

func canonical(cycle []string) string {
	best := ""
	for i := range cycle {
		rotated := append(append([]string(nil), cycle[i:]...), cycle[:i]...)
		joined := strings.Join(rotated, "\x00")
		if best == "" || joined < best {
			best = joined
		}
	}
	parts := strings.Split(best, "\x00")
	return strings.Join(append(parts, parts[0]), " -> ")
}

// checkOverlap fails each reference of a dependent plugin that copies one of
// its dependency.
func (r *run) checkOverlap() error {
	for _, p := range r.plugins {
		if p.codex == nil {
			continue
		}
		own, err := r.references(p)
		if err != nil {
			return err
		}
		for _, dep := range p.codex.Dependencies {
			target, ok := r.byName[dep]
			if !ok {
				continue
			}
			theirs, err := r.references(target)
			if err != nil {
				return err
			}
			for _, a := range own {
				for _, b := range theirs {
					containment := containment(a.shingles, b.shingles)
					if containment >= r.options.OverlapContainment {
						r.report.Fail("plugin.reference-overlap", a.file, 1,
							"skill ownership: %s %s duplicates %s %s (%d%% five-word-shingle containment); keep only the dependent skill's specialization",
							p.pluginName, strings.TrimPrefix(a.file, p.dir+"/"), dep, strings.TrimPrefix(b.file, target.dir+"/"), int(containment*100+0.5))
					}
				}
			}
		}
	}
	return nil
}

type reference struct {
	file     string
	shingles map[string]struct{}
}

func (r *run) references(p *pkg) ([]reference, error) {
	var out []reference
	for _, dir := range p.skills {
		entries, err := r.ctx.Tree.List(path.Join(dir, "references"))
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir || !strings.HasSuffix(e.Name, ".md") {
				continue
			}
			file := path.Join(dir, "references", e.Name)
			text, err := r.ctx.Tree.Read(file)
			if err != nil {
				return nil, err
			}
			out = append(out, reference{file: file, shingles: markdown.Shingles(text, 5)})
		}
	}
	return out, nil
}

func containment(a, b map[string]struct{}) float64 {
	denominator := len(a)
	if len(b) < denominator {
		denominator = len(b)
	}
	if denominator < minShingles {
		return 0
	}
	shared := 0
	for s := range a {
		if _, ok := b[s]; ok {
			shared++
		}
	}
	return float64(shared) / float64(denominator)
}
