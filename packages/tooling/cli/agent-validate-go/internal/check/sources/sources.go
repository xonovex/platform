// Package sources checks the source catalog of each skill: the schema of every
// catalog file and the rules that hold across its sources.
package sources

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

// Name is the name of the check.
const Name = sources.CheckName

// Options tune the check.
type Options struct {
	// Required fails a skill that has no catalog. A repository that has not
	// moved every skill to a catalog yet leaves it off.
	Required bool `json:"required"`
}

func defaults() Options { return Options{Required: true} }

// Check is the sources check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "skill source catalogs: schema, identity and review rules" }

// DefaultPaths returns the plugin package root.
func (Check) DefaultPaths() []string { return []string{"packages/runtime/plugin"} }

// Run checks the catalog of each skill below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	dirs, err := catalog.SkillDirs(ctx.Tree, ctx.Paths)
	if err != nil {
		return nil, err
	}
	loader := sources.Loader{FS: ctx.Tree.FS, Validator: ctx.Validator}
	report := check.NewReporter(Name)
	for _, dir := range dirs {
		_, problems, err := loader.Load(ctx.Tree.Abs(dir))
		if err != nil {
			return nil, err
		}
		for _, f := range problems {
			if f.Level == findings.Pass || (f.Code == sources.CodeMissing && !options.Required) {
				continue
			}
			f.Path = ctx.Tree.Rel(f.Path)
			report.Merge([]findings.Finding{f})
		}
	}
	return report.Findings(), nil
}
