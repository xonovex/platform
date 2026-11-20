// Package plugin checks each plugin package of a catalog: its manifests, its
// skill packaging, the dependencies it declares to npm and to the harness, the
// README that lists its skills, and the handoffs that justify a dependency.
package plugin

import (
	"encoding/json"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "plugin"

// Options tune the check.
type Options struct {
	// DirectoryPrefix starts the directory name of a plugin package.
	DirectoryPrefix string `json:"directoryPrefix"`
	// PluginPrefix starts the name that the manifests give the plugin.
	PluginPrefix string `json:"pluginPrefix"`
	// PackagePrefix starts the npm name of a plugin package.
	PackagePrefix string `json:"packagePrefix"`
	// SkillsPath is the one skill path a manifest names.
	SkillsPath string `json:"skillsPath"`
	// OverlapContainment is the shingle containment at which a reference of a
	// dependent plugin counts as a copy of its dependency.
	OverlapContainment float64 `json:"overlapContainment"`
}

func defaults() Options {
	return Options{
		DirectoryPrefix:    "plugin-",
		PluginPrefix:       "xonovex-",
		PackagePrefix:      "@xonovex/plugin-",
		SkillsPath:         "./skills/",
		OverlapContainment: 0.3,
	}
}

// minShingles is the size below which a reference is too small to call a copy.
const minShingles = 20

// Check is the plugin check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "plugin manifests, skill packaging, dependencies, README catalog and handoffs"
}

// DefaultPaths returns the plugin root.
func (Check) DefaultPaths() []string { return []string{"packages/runtime/plugin"} }

type manifest struct {
	file         string
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Skills       any      `json:"skills"`
}

type packageJSON struct {
	file         string
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Dependencies map[string]string `json:"dependencies"`
}

// pkg is a plugin package as the rules read it.
type pkg struct {
	dir        string
	name       string // the directory name
	pluginName string // the name its manifests must carry
	pkgJSON    *packageJSON
	claude     *manifest
	codex      *manifest
	neutral    *manifest
	skills     []string
	hasClaude  bool
	hasCodex   bool
}

type run struct {
	ctx     check.Context
	report  *check.Reporter
	options Options
	plugins []*pkg
	byName  map[string]*pkg
}

// Run checks every plugin package below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	r := &run{ctx: ctx, report: check.NewReporter(Name), options: options, byName: map[string]*pkg{}}
	if err := r.load(); err != nil {
		return nil, err
	}
	for _, p := range r.plugins {
		if err := r.checkPackage(p); err != nil {
			return nil, err
		}
	}
	r.checkDependencies()
	if err := r.checkOverlap(); err != nil {
		return nil, err
	}
	return r.report.Findings(), nil
}

func (r *run) load() error {
	seen := map[string]bool{}
	for _, root := range r.ctx.Paths {
		entries, err := r.ctx.Tree.List(root)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir || seen[path.Join(root, entry.Name)] {
				continue
			}
			dir := path.Join(root, entry.Name)
			seen[dir] = true
			p, ok, err := r.read(dir, entry.Name)
			if err != nil {
				return err
			}
			if ok {
				r.plugins = append(r.plugins, p)
			}
		}
	}
	sort.Slice(r.plugins, func(i, j int) bool { return r.plugins[i].dir < r.plugins[j].dir })
	return nil
}

func (r *run) read(dir, name string) (*pkg, bool, error) {
	p := &pkg{dir: dir, name: name, pluginName: r.options.PluginPrefix + strings.TrimPrefix(name, r.options.DirectoryPrefix)}
	hasSkills := r.ctx.Tree.IsDir(path.Join(dir, "skills"))
	p.hasClaude = r.ctx.Tree.IsFile(path.Join(dir, ".claude-plugin", "plugin.json"))
	p.hasCodex = r.ctx.Tree.IsFile(path.Join(dir, ".codex-plugin", "plugin.json"))
	var err error
	if p.claude, err = r.manifest(path.Join(dir, ".claude-plugin", "plugin.json")); err != nil {
		return nil, false, err
	}
	if p.codex, err = r.manifest(path.Join(dir, ".codex-plugin", "plugin.json")); err != nil {
		return nil, false, err
	}
	if p.neutral, err = r.manifest(path.Join(dir, "plugin.json")); err != nil {
		return nil, false, err
	}
	if p.pkgJSON, err = r.packageJSON(path.Join(dir, "package.json")); err != nil {
		return nil, false, err
	}
	if !hasSkills && !p.hasClaude && !p.hasCodex {
		return nil, false, nil
	}
	if !strings.HasPrefix(name, r.options.DirectoryPrefix) {
		return nil, false, nil
	}
	if p.skills, err = catalog.SkillDirs(r.ctx.Tree, []string{path.Join(dir, "skills")}); err != nil {
		return nil, false, err
	}
	r.byName[p.pluginName] = p
	return p, true, nil
}

// manifest reads a plugin manifest. A missing file is nil; a broken one is a
// finding and also nil.
func (r *run) manifest(file string) (*manifest, error) {
	text, ok, err := r.ctx.Tree.Optional(file)
	if err != nil || !ok {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		r.report.Fail("plugin.manifest-invalid", file, 1, "cannot parse JSON in %s: %v", file, err)
		return nil, nil
	}
	m.file = file
	if m.Name == "" {
		r.report.Fail("plugin.manifest-name", file, 1, "%s needs a string name", file)
		return nil, nil
	}
	return &m, nil
}

func (r *run) packageJSON(file string) (*packageJSON, error) {
	text, ok, err := r.ctx.Tree.Optional(file)
	if err != nil || !ok {
		return nil, err
	}
	var p packageJSON
	if err := json.Unmarshal([]byte(text), &p); err != nil {
		r.report.Fail("plugin.package-invalid", file, 1, "cannot parse JSON in %s: %v", file, err)
		return nil, nil
	}
	p.file = file
	return &p, nil
}

func skillPaths(m *manifest) []string {
	switch v := m.Skills.(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func sameStrings(a, b []string) bool {
	a, b = append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (r *run) checkPackage(p *pkg) error {
	if r.ctx.Tree.IsDir(path.Join(p.dir, "commands")) {
		r.report.Fail("plugin.commands-present", path.Join(p.dir, "commands"), 1, "a plugin package holds skills only, so its commands directory must not exist")
	}
	if !p.hasClaude || !p.hasCodex {
		r.report.Fail("plugin.manifest-missing", p.dir, 1, "%s needs both Claude and Codex manifests", p.dir)
		return r.checkReadme(p)
	}
	if p.claude == nil || p.codex == nil {
		return r.checkReadme(p)
	}
	for _, m := range []*manifest{p.claude, p.codex} {
		if m.Name != p.pluginName {
			r.report.Fail("plugin.name", m.file, 1, "manifests in %s must be named %s", p.dir, p.pluginName)
		}
	}
	if p.claude.Name != p.codex.Name {
		r.report.Fail("plugin.names-differ", p.dir, 1, "manifest names differ in %s (%s != %s)", p.dir, p.claude.Name, p.codex.Name)
		return r.checkReadme(p)
	}
	if !sameStrings(p.claude.Dependencies, p.codex.Dependencies) {
		r.report.Fail("plugin.dependencies-differ", p.dir, 1, "manifest dependencies differ for %s", p.claude.Name)
	}
	if !sameStrings(skillPaths(p.claude), skillPaths(p.codex)) {
		r.report.Fail("plugin.skills-differ", p.dir, 1, "manifest skill paths differ for %s", p.claude.Name)
	}
	if len(p.skills) == 0 {
		r.report.Fail("plugin.no-skills", p.dir, 1, "%s contains no skills", p.dir)
	}
	for label, m := range map[string]*manifest{"Claude": p.claude, "Codex": p.codex} {
		if !sameStrings(skillPaths(m), []string{r.options.SkillsPath}) {
			r.report.Fail("plugin.skills-path", m.file, 1, "%s manifest in %s must point directly to %s", label, p.dir, r.options.SkillsPath)
		}
	}
	if err := r.checkReadme(p); err != nil {
		return err
	}
	return r.checkUsage(p)
}

var (
	skillsHeading = regexp.MustCompile(`^##\s+Skills\s*$`)
	anyHeading    = regexp.MustCompile(`^##\s+\S`)
)

var readmeRow = regexp.MustCompile(`(?m)^\|\s*\[([a-z0-9-]+)\]\(skills/([a-z0-9-]+)/SKILL\.md\)\s*\|`)

// The skill tables sit under the README's Skills heading, so a table elsewhere
// in the file is not read as a skill listing.
func skillsSection(source string) string {
	lines := strings.Split(source, "\n")
	start := -1
	for i, line := range lines {
		if skillsHeading.MatchString(line) {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	rest := lines[start+1:]
	for i, line := range rest {
		if anyHeading.MatchString(line) {
			rest = rest[:i]
			break
		}
	}
	return strings.Join(rest, "\n")
}

func (r *run) checkReadme(p *pkg) error {
	file := path.Join(p.dir, "README.md")
	text, ok, err := r.ctx.Tree.Optional(file)
	if err != nil || !ok {
		return err
	}
	listed := map[string]bool{}
	for _, m := range readmeRow.FindAllStringSubmatch(skillsSection(text), -1) {
		if m[1] == m[2] {
			listed[m[1]] = true
		}
	}
	shipped := map[string]bool{}
	for _, dir := range p.skills {
		shipped[path.Base(dir)] = true
	}
	for _, name := range markdown.Sorted(keys(listed)) {
		if !shipped[name] {
			r.report.Fail("plugin.readme-skill-missing", file, 1, "README lists skill '%s' but 'skills/%s/SKILL.md' does not exist", name, name)
		}
	}
	for _, name := range markdown.Sorted(keys(shipped)) {
		if !listed[name] {
			r.report.Fail("plugin.readme-skill-unlisted", file, 1, "skill '%s' is not listed in the README skill tables", name)
		}
	}
	return nil
}

func keys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}
