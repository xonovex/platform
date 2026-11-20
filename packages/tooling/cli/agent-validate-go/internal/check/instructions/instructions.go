// Package instructions checks the eval corpus of AGENTS.md files: each corpus
// matches its schema, names an instruction file that exists, anchors every case
// in that file, and covers the rules of the file.
package instructions

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

// Name is the name of the check.
const Name = "instructions"

// Options tune the check.
type Options struct {
	// Corpus is the file name of the corpus beside an instruction file.
	Corpus string `json:"corpus"`
	// Instruction is the file name of an instruction file.
	Instruction string `json:"instruction"`
	// Require makes a missing corpus a failure.
	Require bool `json:"require"`
	// Coverage reports a rule that no case anchors.
	Coverage bool `json:"coverage"`
}

func defaults() Options {
	return Options{Corpus: "instructions.eval.json", Instruction: "AGENTS.md", Coverage: true}
}

type anchor struct {
	Heading string `json:"heading"`
	Text    string `json:"text"`
}

type corpus struct {
	Instruction string `json:"instruction"`
	Cases       []struct {
		ID     string `json:"id"`
		Anchor anchor `json:"anchor"`
	} `json:"cases"`
}

// Check is the instructions check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string {
	return "instruction eval corpora: schema, anchors and rule coverage"
}

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the corpora below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool {
		base := path.Base(p)
		return base == options.Corpus || base == options.Instruction
	})
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		dir := path.Dir(file)
		switch path.Base(file) {
		case options.Instruction:
			if options.Require && !ctx.Tree.IsFile(path.Join(dir, options.Corpus)) {
				report.Fail("instructions.corpus-missing", file, 1, "%s has no %s beside it", file, options.Corpus)
			}
		case options.Corpus:
			if err := checkCorpus(ctx, report, options, file); err != nil {
				return nil, err
			}
		}
	}
	return report.Findings(), nil
}

func checkCorpus(ctx check.Context, report *check.Reporter, options Options, file string) error {
	text, err := ctx.Tree.Read(file)
	if err != nil {
		return err
	}
	violations, err := ctx.Validator.Validate(schemas.EvalInstruction, []byte(text))
	if err != nil {
		return err
	}
	for _, v := range violations {
		report.Fail("instructions.schema", file, 1, "%s%s: %s", path.Base(file), v.Pointer, v.Message)
	}
	if len(violations) > 0 {
		return nil
	}
	var parsed corpus
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		report.Fail("instructions.json", file, 1, "%s is not readable: %v", file, err)
		return nil
	}
	target := path.Join(path.Dir(file), parsed.Instruction)
	source, found, err := ctx.Tree.Optional(target)
	if err != nil {
		return err
	}
	if !found {
		report.Fail("instructions.target-missing", file, 1, "%s names %s, which does not exist", file, parsed.Instruction)
		return nil
	}
	headings := markdown.Headings(source)
	headingNames := map[string]bool{}
	for _, h := range headings {
		headingNames[strings.ToLower(h.Text)] = true
	}
	seen := map[string]bool{}
	var texts, sections []string
	for _, c := range parsed.Cases {
		if seen[c.ID] {
			report.Fail("instructions.duplicate-case", file, 1, "%s repeats case id %q", file, c.ID)
		}
		seen[c.ID] = true
		switch {
		case c.Anchor.Heading != "":
			if !headingNames[strings.ToLower(c.Anchor.Heading)] {
				report.Fail("instructions.anchor-heading", file, 1, "case %q anchors heading %q, which %s does not have", c.ID, c.Anchor.Heading, parsed.Instruction)
			}
			sections = append(sections, strings.ToLower(c.Anchor.Heading))
		case !strings.Contains(source, c.Anchor.Text):
			report.Fail("instructions.anchor-text", file, 1, "case %q quotes text that %s does not contain", c.ID, parsed.Instruction)
		default:
			texts = append(texts, c.Anchor.Text)
		}
	}
	if options.Coverage {
		unguarded(report, target, source, texts, sections)
	}
	return nil
}

// A rule is a top-level bullet. It is guarded when a case quotes it or anchors
// the section that holds it.
func unguarded(report *check.Reporter, file, source string, texts, sections []string) {
	section := ""
	masked := markdown.MaskFences(source)
	for i, line := range strings.Split(masked, "\n") {
		if m := headingLine(line); m != "" {
			section = strings.ToLower(m)
			continue
		}
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		if contains(sections, section) || quoted(texts, line) {
			continue
		}
		report.Warn("instructions.unguarded-rule", file, i+1, "no eval case guards this rule of %s", file)
	}
}

func headingLine(line string) string {
	for _, h := range markdown.Headings(line) {
		return h.Text
	}
	return ""
}

func contains(all []string, value string) bool {
	for _, a := range all {
		if a == value {
			return true
		}
	}
	return false
}

func quoted(texts []string, line string) bool {
	for _, t := range texts {
		if strings.Contains(line, t) {
			return true
		}
	}
	return false
}
