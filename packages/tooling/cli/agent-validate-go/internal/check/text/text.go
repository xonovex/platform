// Package text checks the typography of repository prose and the line breaks
// of Markdown: no em dash, no ellipsis character, no typographic quotes, and
// one paragraph or list item per line.
package text

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/glob"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "text"

// Options tune the check.
type Options struct {
	// Extensions are the file types that carry prose.
	Extensions []string `json:"extensions"`
	// Skip names directories, by base name, that the walk does not enter.
	Skip []string `json:"skip"`
	// Exclude lists glob patterns of generated files that the check leaves out.
	Exclude []string `json:"exclude"`
	// Wrap switches the hard-wrapped Markdown rule on or off.
	Wrap bool `json:"wrap"`
	// WrapExtensions are the file types the hard-wrap rule reads.
	WrapExtensions []string `json:"wrapExtensions"`
}

func defaults() Options {
	return Options{
		Extensions:     []string{".json", ".jsonc", ".markdown", ".md", ".py", ".sh", ".template", ".txt", ".yaml", ".yml"},
		Skip:           []string{".moon", "coverage"},
		Exclude:        []string{"**/CHANGELOG.md"},
		Wrap:           true,
		WrapExtensions: []string{".md", ".markdown"},
	}
}

type rule struct {
	code    string
	label   string
	hint    string
	pattern *regexp.Regexp
}

// Each character also hides behind a JSON or JavaScript unicode escape, which
// a search for the literal character misses.
var rules = []rule{
	{"text.em-dash", "em dash", "use a comma, colon, or full stop", regexp.MustCompile(`\x{2014}`)},
	{"text.escaped-em-dash", "escaped em dash", "use a comma, colon, or full stop", regexp.MustCompile(`(?i)\\u2014`)},
	{"text.ellipsis", "ellipsis", "use three periods", regexp.MustCompile(`\x{2026}`)},
	{"text.escaped-ellipsis", "escaped ellipsis", "use three periods", regexp.MustCompile(`(?i)\\u2026`)},
	{"text.smart-single-quote", "smart single quote", "use a straight apostrophe", regexp.MustCompile(`[\x{2018}\x{2019}]`)},
	{"text.escaped-smart-single-quote", "escaped smart single quote", "use a straight apostrophe", regexp.MustCompile(`(?i)\\u201[89]`)},
	{"text.smart-double-quote", "smart double quote", "use straight double quotes", regexp.MustCompile(`[\x{201C}\x{201D}]`)},
	{"text.escaped-smart-double-quote", "escaped smart double quote", "use straight double quotes", regexp.MustCompile(`(?i)\\u201[cd]`)},
}

const excerptRunes = 80

// Check is the text check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "typographic punctuation in prose files and hard-wrapped Markdown"
}

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the prose files below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	excluded, err := compileAll(options.Exclude)
	if err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool {
		return hasExtension(p, options.Extensions) && !inSkipped(p, options.Skip) && !matchesAny(p, excluded)
	})
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		source, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		punctuation(report, file, source)
		if options.Wrap && hasExtension(file, options.WrapExtensions) {
			hardWraps(report, file, source)
		}
	}
	return report.Findings(), nil
}

func compileAll(patterns []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		re, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("exclude pattern %q: %w", pattern, err)
		}
		out[i] = re
	}
	return out, nil
}

func matchesAny(file string, all []*regexp.Regexp) bool {
	for _, re := range all {
		if re.MatchString(file) {
			return true
		}
	}
	return false
}

func hasExtension(file string, extensions []string) bool {
	ext := path.Ext(file)
	for _, e := range extensions {
		if e == ext {
			return true
		}
	}
	return false
}

func inSkipped(file string, skip []string) bool {
	for _, segment := range strings.Split(path.Dir(file), "/") {
		for _, s := range skip {
			if segment == s {
				return true
			}
		}
	}
	return false
}

func excerpt(line string) string {
	line = strings.TrimSpace(line)
	if utf8.RuneCountInString(line) <= excerptRunes {
		return line
	}
	return string([]rune(line)[:excerptRunes])
}

func punctuation(report *check.Reporter, file, source string) {
	for i, line := range markdown.Lines(source) {
		for _, r := range rules {
			if r.pattern.MatchString(line) {
				report.Fail(r.code, file, i+1, "%s: %s (%s): %s", r.label, excerpt(line), r.hint, file)
			}
		}
	}
}
