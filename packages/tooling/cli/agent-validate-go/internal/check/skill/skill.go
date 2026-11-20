// Package skill checks each skill of a catalog against the Agent Skills
// specification and the authoring rules of the repository: its frontmatter, its
// body, the references it links, its evals and the files it ships.
package skill

import (
	"path"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
)

// Name is the name of the check.
const Name = "skill"

// Options tune the check.
type Options struct {
	// DescriptionLimit is the longest description in characters. The
	// specification allows 1024; the catalog keeps a shorter limit.
	DescriptionLimit int `json:"descriptionLimit"`
	// BodyFailLines is the line count above which a body fails.
	BodyFailLines int `json:"bodyFailLines"`
	// BodyWarnLines is the line count above which a body draws a warning.
	BodyWarnLines int `json:"bodyWarnLines"`
	// VendorAdapters are skill names that carry a vendor name by design.
	VendorAdapters []string `json:"vendorAdapters"`
	// RequireEvals makes a missing eval-output.json or eval-trigger.json a failure.
	RequireEvals bool `json:"requireEvals"`
	// ForbiddenExtensions are the file extensions a skill must not ship as
	// implementation files.
	ForbiddenExtensions []string `json:"forbiddenExtensions"`
}

func defaults() Options {
	return Options{
		DescriptionLimit:    820,
		BodyFailLines:       500,
		BodyWarnLines:       400,
		RequireEvals:        true,
		ForbiddenExtensions: []string{".ts", ".mjs"},
	}
}

// Check is the skill check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "Agent Skills specification and catalog rules for each SKILL.md"
}

// DefaultPaths returns the paths the check reads without a project setting.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks every skill below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	dirs, err := catalog.SkillDirs(ctx.Tree, ctx.Paths)
	if err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	for _, dir := range dirs {
		if err := checkSkill(ctx, report, options, dir); err != nil {
			return nil, err
		}
	}
	return report.Findings(), nil
}

func checkSkill(ctx check.Context, report *check.Reporter, options Options, dir string) error {
	file := path.Join(dir, catalog.SkillFile)
	text, err := ctx.Tree.Read(file)
	if err != nil {
		return err
	}
	doc, err := document.Parse(text)
	if err != nil {
		report.Fail("skill.frontmatter-invalid", file, 1, "%v", err)
		return nil
	}
	references, err := readReferences(ctx, dir)
	if err != nil {
		return err
	}
	checkFrontmatter(report, options, doc, file, path.Base(dir))
	checkLoaderQuoting(report, doc, file)
	checkBody(report, options, doc, file)
	checkSourcePlacement(report, doc, file, references)
	checkReferenceLinks(ctx, report, doc, file, dir)
	checkReferenceTocs(report, references)
	checkReferenceFileLinks(ctx, report, dir, references)
	checkHarnessNeutrality(report, doc, file)
	checkCredentials(report, file, text, references)
	if err := checkImplementationFiles(ctx, report, options, doc, file, dir); err != nil {
		return err
	}
	return checkEvals(ctx, report, options, doc, dir)
}

// reference is one Markdown file directly under references/.
type reference struct {
	name string // path inside references/, slash-separated
	path string // repository path
	text string
}

// readReferences returns every Markdown file below references/, sorted.
func readReferences(ctx check.Context, dir string) ([]reference, error) {
	root := path.Join(dir, "references")
	files, err := ctx.Tree.Walk(root, func(p string) bool { return strings.HasSuffix(p, ".md") })
	if err != nil {
		return nil, err
	}
	out := make([]reference, 0, len(files))
	for _, file := range files {
		text, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		out = append(out, reference{name: strings.TrimPrefix(file, root+"/"), path: file, text: text})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, nil
}

func topLevel(all []reference) []reference {
	var out []reference
	for _, r := range all {
		if !strings.Contains(r.name, "/") {
			out = append(out, r)
		}
	}
	return out
}

func typed(value any) string { return document.TypeName(value) }
