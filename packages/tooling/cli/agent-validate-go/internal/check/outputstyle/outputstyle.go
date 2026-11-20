// Package outputstyle checks the frontmatter and instructions of output style
// files.
package outputstyle

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/component"
)

// Name is the name of the check.
const Name = "output-style"

var spec = component.Spec{
	Prefix:   "output-style",
	Noun:     "an output style",
	Known:    []string{"description", "force-for-plugin", "keep-coding-instructions", "name"},
	Required: []string{"name", "description"},
	Body:     "instructions",
	Fields: []component.Field{
		{Key: "keep-coding-instructions", Valid: component.IsBool, Message: "keep-coding-instructions must be true or false"},
		{Key: "force-for-plugin", Valid: component.IsBool, Message: "force-for-plugin must be true or false"},
	},
}

// Check is the output style check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "output-style frontmatter and instruction bodies" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the output style files below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	if err := check.DecodeOptions(ctx.Options, &struct{}{}); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := component.Files(ctx.Tree, ctx.Paths, "output-styles")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		source, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		spec.Check(report, file, source)
	}
	return report.Findings(), nil
}
