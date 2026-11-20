// Package marketplace checks the Claude and Codex marketplace files against
// the plugin packages they list: one release version, one description, one
// source, and no missing, unexpected or repeated entry.
package marketplace

import (
	"encoding/json"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

// Name is the name of the check.
const Name = "marketplace"

// Options tune the check.
type Options struct {
	// Claude is the Claude marketplace file.
	Claude string `json:"claude"`
	// Codex is the Codex marketplace file.
	Codex string `json:"codex"`
	// Lockfile is the npm lockfile that records each plugin version. An empty
	// value switches the rule off.
	Lockfile string `json:"lockfile"`
	// DirectoryPrefix starts the directory name of a plugin package.
	DirectoryPrefix string `json:"directoryPrefix"`
	// PluginPrefix starts the name of a plugin in a marketplace.
	PluginPrefix string `json:"pluginPrefix"`
	// PackagePrefix starts the npm name of a plugin package.
	PackagePrefix string `json:"packagePrefix"`
}

func defaults() Options {
	return Options{
		Claude:          ".claude-plugin/marketplace.json",
		Codex:           ".agents/plugins/marketplace.json",
		Lockfile:        "package-lock.json",
		DirectoryPrefix: "plugin-",
		PluginPrefix:    "xonovex-",
		PackagePrefix:   "@xonovex/plugin-",
	}
}

var semver = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// Check is the marketplace check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "Claude and Codex marketplaces against the plugin packages and the release version"
}

// DefaultPaths returns the plugin root.
func (Check) DefaultPaths() []string { return []string{"packages/runtime/plugin"} }

type entry struct {
	Name        string `json:"name"`
	Source      any    `json:"source"`
	Description string `json:"description"`
}

func (e entry) source() string {
	switch v := e.Source.(type) {
	case string:
		return v
	case map[string]any:
		s, _ := v["path"].(string)
		return s
	}
	return ""
}

type file struct {
	Metadata struct {
		Version string `json:"version"`
	} `json:"metadata"`
	Plugins []entry `json:"plugins"`
}

type manifest struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Description  string        `json:"description"`
	Skills       any           `json:"skills"`
	Dependencies dependencyMap `json:"dependencies"`
	// The other three maps of package.json also install a plugin package, so
	// each must name the release version.
	DevDependencies      dependencyMap `json:"devDependencies"`
	PeerDependencies     dependencyMap `json:"peerDependencies"`
	OptionalDependencies dependencyMap `json:"optionalDependencies"`
}

// installed merges the four dependency maps of package.json. A name in a later
// map replaces the same name in an earlier one, as npm resolves it.
func (m manifest) installed() map[string]string {
	merged := map[string]string{}
	for _, deps := range []dependencyMap{m.Dependencies, m.DevDependencies, m.PeerDependencies, m.OptionalDependencies} {
		for name, version := range deps {
			merged[name] = version
		}
	}
	return merged
}

// dependencyMap reads the dependencies of package.json, which name a version
// each. A plugin manifest lists dependencies as an array of plugin names
// instead, and holds no version to compare.
type dependencyMap map[string]string

func (d *dependencyMap) UnmarshalJSON(data []byte) error {
	var versions map[string]string
	if err := json.Unmarshal(data, &versions); err != nil {
		*d = nil
		return nil
	}
	*d = versions
	return nil
}

type pluginPackage struct{ dir string }

type run struct {
	ctx     check.Context
	report  *check.Reporter
	options Options
}

// Run checks both marketplaces.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	r := &run{ctx: ctx, report: check.NewReporter(Name), options: options}
	claude, okClaude, err := readJSON[file](r, options.Claude)
	if err != nil {
		return nil, err
	}
	codex, okCodex, err := readJSON[file](r, options.Codex)
	if err != nil {
		return nil, err
	}
	if !okClaude || !okCodex {
		return r.report.Findings(), nil
	}
	packages, err := r.packages()
	if err != nil {
		return nil, err
	}
	release := claude.Metadata.Version
	if !semver.MatchString(release) {
		r.report.Fail("marketplace.version", options.Claude, 1, "marketplace release must use a semantic version")
	}
	claudeEntries, codexEntries := index(claude), index(codex)
	var names []string
	for _, p := range packages {
		r.checkPackage(p, release, claudeEntries, codexEntries)
		names = append(names, r.expectedName(p.dir))
	}
	r.checkInventory("Claude marketplace", options.Claude, claude, names)
	r.checkInventory("Codex marketplace", options.Codex, codex, names)
	if options.Lockfile != "" {
		if err := r.checkLockfile(release, packages); err != nil {
			return nil, err
		}
	}
	return r.report.Findings(), nil
}

// readJSON reads a JSON file into T. A missing or broken file is a finding.
func readJSON[T any](r *run, name string) (T, bool, error) {
	var value T
	text, ok, err := r.ctx.Tree.Optional(name)
	if err != nil {
		return value, false, err
	}
	if !ok {
		r.report.Fail("marketplace.missing", name, 1, "%s does not exist", name)
		return value, false, nil
	}
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		r.report.Fail("marketplace.invalid", name, 1, "%s is not valid: %v", name, err)
		return value, false, nil
	}
	return value, true, nil
}

type entries struct {
	sources      map[string]string
	descriptions map[string]string
}

func index(f file) entries {
	e := entries{sources: map[string]string{}, descriptions: map[string]string{}}
	for _, p := range f.Plugins {
		e.sources[p.Name] = p.source()
		e.descriptions[p.Name] = p.Description
	}
	return e
}

func (r *run) expectedName(dir string) string {
	return r.options.PluginPrefix + strings.TrimPrefix(path.Base(dir), r.options.DirectoryPrefix)
}

func (r *run) packages() ([]pluginPackage, error) {
	var out []pluginPackage
	seen := map[string]bool{}
	for _, root := range r.ctx.Paths {
		list, err := r.ctx.Tree.List(root)
		if err != nil {
			return nil, err
		}
		for _, e := range list {
			dir := path.Join(root, e.Name)
			if !e.IsDir || seen[dir] {
				continue
			}
			seen[dir] = true
			out = append(out, pluginPackage{dir: dir})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].dir < out[j].dir })
	return out, nil
}

// optional reads a manifest that must exist and parse; a failure is reported
// and nil returned.
func (r *run) manifest(dir, name, label string) *manifest {
	file := path.Join(dir, name)
	text, ok, err := r.ctx.Tree.Optional(file)
	if err != nil || !ok {
		r.report.Fail("marketplace.package-missing", dir, 1, "%s has %s", dir, label)
		return nil
	}
	var m manifest
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		r.report.Fail("marketplace.package-invalid", file, 1, "%s is not valid JSON: %v", file, err)
		return nil
	}
	return &m
}

func (r *run) checkPackage(p pluginPackage, release string, claudeEntries, codexEntries entries) {
	pkg := r.manifest(p.dir, "package.json", "package.json")
	neutral := r.manifest(p.dir, "plugin.json", "a neutral plugin manifest")
	claude := r.manifest(p.dir, ".claude-plugin/plugin.json", "a Claude plugin manifest")
	codex := r.manifest(p.dir, ".codex-plugin/plugin.json", "a Codex plugin manifest")
	if pkg == nil || neutral == nil || claude == nil || codex == nil {
		return
	}
	source := "./" + p.dir
	if neutral.Name != claude.Name || neutral.Version != release || neutral.Description != pkg.Description || !isSkillsDirectory(neutral.Skills) {
		r.report.Fail("marketplace.neutral", path.Join(p.dir, "plugin.json"), 1, "%s neutral manifest matches the plugin identity and skill directory", pkg.Name)
	}
	if pkg.Version != release {
		r.report.Fail("marketplace.version-package", path.Join(p.dir, "package.json"), 1, "%s package version matches %s", pkg.Name, release)
	}
	if claude.Version != release {
		r.report.Fail("marketplace.version-claude", path.Join(p.dir, ".claude-plugin/plugin.json"), 1, "%s Claude manifest version matches %s", claude.Name, release)
	}
	if codex.Version != release {
		r.report.Fail("marketplace.version-codex", path.Join(p.dir, ".codex-plugin/plugin.json"), 1, "%s Codex manifest version matches %s", codex.Name, release)
	}
	if claudeEntries.sources[claude.Name] != source {
		r.report.Fail("marketplace.source-claude", r.options.Claude, 1, "%s has the expected Claude marketplace source %s", claude.Name, source)
	}
	if codexEntries.sources[codex.Name] != source {
		r.report.Fail("marketplace.source-codex", r.options.Codex, 1, "%s has the expected Codex marketplace source %s", codex.Name, source)
	}
	if claude.Name != codex.Name {
		r.report.Fail("marketplace.names", p.dir, 1, "%s manifest names match", pkg.Name)
	}
	// package.json is the single description source: a plugin package declares
	// one, and every manifest and marketplace entry repeats it verbatim.
	if pkg.Description == "" {
		r.report.Fail("marketplace.description-missing", path.Join(p.dir, "package.json"), 1, "%s package declares a description", pkg.Name)
	}
	for _, c := range []struct {
		label, got, file string
	}{
		{"Claude manifest", claude.Description, path.Join(p.dir, ".claude-plugin/plugin.json")},
		{"Claude marketplace", claudeEntries.descriptions[claude.Name], r.options.Claude},
		{"Codex manifest", codex.Description, path.Join(p.dir, ".codex-plugin/plugin.json")},
		{"Codex marketplace", codexEntries.descriptions[codex.Name], r.options.Codex},
	} {
		if c.got != pkg.Description {
			r.report.Fail("marketplace.description", c.file, 1, "%s package and %s descriptions match", pkg.Name, c.label)
		}
	}
	installed := pkg.installed()
	names := make([]string, 0, len(installed))
	for name := range installed {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if strings.HasPrefix(name, r.options.PackagePrefix) && installed[name] != release {
			r.report.Fail("marketplace.dependency-version", path.Join(p.dir, "package.json"), 1, "%s dependency %s matches %s", pkg.Name, name, release)
		}
	}
}

func isSkillsDirectory(skills any) bool {
	switch v := skills.(type) {
	case string:
		return v == "./skills/"
	}
	return false
}

func (r *run) checkInventory(label, file string, f file, expected []string) {
	counts := map[string]int{}
	for _, p := range f.Plugins {
		counts[p.Name]++
	}
	want := map[string]bool{}
	for _, n := range expected {
		want[n] = true
	}
	var missing, unexpected, duplicates []string
	for _, n := range expected {
		if counts[n] == 0 {
			missing = append(missing, n)
		}
	}
	for n, c := range counts {
		if !want[n] {
			unexpected = append(unexpected, n)
		}
		if c > 1 {
			duplicates = append(duplicates, n)
		}
	}
	sort.Strings(missing)
	sort.Strings(unexpected)
	sort.Strings(duplicates)
	if len(missing) > 0 {
		r.report.Fail("marketplace.missing-entry", file, 1, "%s is missing plugin entries: %s", label, strings.Join(missing, ", "))
	}
	if len(unexpected) > 0 {
		r.report.Fail("marketplace.unexpected-entry", file, 1, "%s has unexpected plugin entries: %s", label, strings.Join(unexpected, ", "))
	}
	if len(duplicates) > 0 {
		r.report.Fail("marketplace.duplicate-entry", file, 1, "%s has duplicate plugin entries: %s", label, strings.Join(duplicates, ", "))
	}
}

func (r *run) checkLockfile(release string, packages []pluginPackage) error {
	text, ok, err := r.ctx.Tree.Optional(r.options.Lockfile)
	if err != nil {
		return err
	}
	if !ok {
		r.report.Fail("marketplace.lockfile-missing", r.options.Lockfile, 1, "%s does not exist", r.options.Lockfile)
		return nil
	}
	var lock struct {
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.Unmarshal([]byte(text), &lock); err != nil {
		r.report.Fail("marketplace.lockfile-invalid", r.options.Lockfile, 1, "%s is not valid JSON: %v", r.options.Lockfile, err)
		return nil
	}
	for _, p := range packages {
		if lock.Packages[p.dir].Version != release {
			r.report.Fail("marketplace.lockfile", r.options.Lockfile, 1, "package-lock records %s@%s", p.dir, release)
		}
	}
	return nil
}
