// Package llms checks llms.txt indexes: each starts with one H1 project name,
// links to at least one document, and links local pages only to Markdown files
// that exist in the built site.
package llms

import (
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "llms"

// Options tune the check.
type Options struct {
	// File is the base name of an index.
	File string `json:"file"`
	// SiteRoot is the directory that a link starting with a slash is relative
	// to. An empty value takes the directory of the index.
	SiteRoot string `json:"siteRoot"`
}

func defaults() Options { return Options{File: "llms.txt"} }

var (
	markdownLink = regexp.MustCompile(`\[[^\]]+\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	projectName  = regexp.MustCompile(`^# [^#\s]`)
)

// Check is the llms check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "llms.txt indexes whose Markdown targets exist" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the indexes below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool { return path.Base(p) == options.File })
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		text, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		root := options.SiteRoot
		if root == "" {
			root = path.Dir(file)
		}
		validate(ctx, report, file, text, root)
	}
	return report.Findings(), nil
}

func validate(ctx check.Context, report *check.Reporter, file, text, siteRoot string) {
	lines := markdown.Lines(text)
	first := -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			first = i
			break
		}
	}
	if first == -1 || !projectName.MatchString(lines[first]) {
		line := 1
		if first != -1 {
			line = first + 1
		}
		report.Fail("llms.heading", file, line, "llms.txt must start with one H1 project name")
	}
	links := 0
	for i, line := range lines {
		for _, m := range markdownLink.FindAllStringSubmatch(line, -1) {
			url := m[1]
			links++
			if markdown.IsExternal(url) || strings.HasPrefix(url, "#") {
				continue
			}
			target, _ := markdown.Split(url)
			switch {
			case !strings.HasSuffix(target, ".md"):
				report.Fail("llms.target-format", file, i+1, "local llms.txt target must be Markdown: %s", url)
			case !ctx.Tree.IsFile(resolve(target, siteRoot, file)):
				report.Fail("llms.target-missing", file, i+1, "local llms.txt target does not exist in the built site: %s", url)
			}
		}
	}
	if links == 0 {
		report.Fail("llms.links", file, 1, "llms.txt contains no documentation links")
	}
}

func resolve(target, siteRoot, source string) string {
	if strings.HasPrefix(target, "/") {
		return path.Join(siteRoot, target)
	}
	return path.Join(path.Dir(source), target)
}
