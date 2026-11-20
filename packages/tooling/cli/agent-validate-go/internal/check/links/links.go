// Package links checks the relative links of Markdown files, including the
// heading anchors they name, and the skill names that a document hands work to
// in bold.
package links

import (
	"path"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "links"

// Options tune the check.
type Options struct {
	// SkillRoots are the catalog roots whose skills a bold name may hand work to.
	// An empty list switches the handoff rule off.
	SkillRoots []string `json:"skillRoots"`
	// Anchors makes a link fragment a failure when the heading it names does
	// not exist in the target Markdown file.
	Anchors bool `json:"anchors"`
}

func defaults() Options {
	return Options{SkillRoots: []string{"packages/runtime/plugin"}, Anchors: true}
}

// Check is the links check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "relative Markdown links, their heading anchors, and bold skill handoffs"
}

// DefaultPaths returns the paths the check reads without a project setting.
func (Check) DefaultPaths() []string { return []string{"."} }

type run struct {
	ctx     check.Context
	report  *check.Reporter
	options Options
	anchors map[string]map[string]bool
}

// Run checks every Markdown file below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool { return strings.HasSuffix(p, ".md") })
	if err != nil {
		return nil, err
	}
	r := &run{ctx: ctx, report: check.NewReporter(Name), options: options, anchors: map[string]map[string]bool{}}
	for _, file := range files {
		text, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		r.checkLinks(file, text)
	}
	if err := r.checkHandoffs(); err != nil {
		return nil, err
	}
	return r.report.Findings(), nil
}

func (r *run) anchorsOf(file string) (map[string]bool, error) {
	if cached, ok := r.anchors[file]; ok {
		return cached, nil
	}
	text, err := r.ctx.Tree.Read(file)
	if err != nil {
		return nil, err
	}
	r.anchors[file] = markdown.Anchors(text)
	return r.anchors[file], nil
}

func (r *run) checkLinks(file, text string) {
	for _, link := range markdown.Links(text) {
		if markdown.IsExternal(link.Target) || markdown.Placeholder(link.Target) {
			continue
		}
		rawPath, fragment := markdown.Split(link.Target)
		target := file
		if rawPath != "" {
			target = path.Join(path.Dir(file), markdown.DecodePath(rawPath))
			if strings.HasPrefix(target, "../") || target == ".." {
				continue
			}
			if !r.ctx.Tree.Exists(target) {
				r.report.Fail("links.missing-target", file, link.Line, "link target '%s' does not exist", link.Target)
				continue
			}
		}
		if fragment == "" || !r.options.Anchors || !strings.HasSuffix(target, ".md") || !r.ctx.Tree.IsFile(target) {
			continue
		}
		anchors, err := r.anchorsOf(target)
		if err != nil {
			r.report.Fail("links.unreadable", file, link.Line, "link target '%s' cannot be read: %v", link.Target, err)
			continue
		}
		if !anchors[markdown.DecodeFragment(fragment)] {
			r.report.Fail("links.missing-anchor", file, link.Line, "link fragment '%s' names no heading of %s", link.Target, target)
		}
	}
}

// A bold skill name is the one written form of a handoff, so each one names a
// skill that exists.
func (r *run) checkHandoffs() error {
	if len(r.options.SkillRoots) == 0 {
		return nil
	}
	dirs, err := catalog.SkillDirs(r.ctx.Tree, r.options.SkillRoots)
	if err != nil {
		return err
	}
	known := map[string]bool{}
	documents := map[string]bool{}
	for _, dir := range dirs {
		known[path.Base(dir)] = true
		documents[path.Join(dir, catalog.SkillFile)] = true
		entries, err := r.ctx.Tree.List(path.Join(dir, "references"))
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir && strings.HasSuffix(e.Name, ".md") {
				documents[path.Join(dir, "references", e.Name)] = true
			}
		}
		plugin := path.Dir(path.Dir(dir))
		if readme := path.Join(plugin, "README.md"); r.ctx.Tree.IsFile(readme) {
			documents[readme] = true
		}
		docs, err := r.ctx.Tree.List(path.Join(plugin, "docs"))
		if err != nil {
			return err
		}
		for _, e := range docs {
			if !e.IsDir && strings.HasSuffix(e.Name, ".md") {
				documents[path.Join(plugin, "docs", e.Name)] = true
			}
		}
	}
	files := make([]string, 0, len(documents))
	for f := range documents {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, file := range files {
		text, err := r.ctx.Tree.Read(file)
		if err != nil {
			return err
		}
		for _, name := range markdown.BoldSkillNames(text) {
			if !known[name] {
				r.report.Fail("links.handoff-missing", file, boldLine(text, name), "skill handoff names missing **%s**", name)
			}
		}
	}
	return nil
}

func boldLine(text, name string) int {
	if index := strings.Index(text, "**"+name+"**"); index >= 0 {
		return markdown.LineOf(text, index)
	}
	return 1
}
