// Package runner runs the steps of a plan and collects their findings. It
// holds no check: the plan names them.
package runner

import (
	"fmt"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/project"
)

// Run runs each step in order. A step that returns an error stops the run: the
// findings of a run that did not finish prove nothing.
func Run(base check.Context, steps []project.Step) ([]findings.Finding, error) {
	var all []findings.Finding
	for _, step := range steps {
		ctx := base
		ctx.Paths = step.Paths
		ctx.Options = step.Options
		got, err := step.Check.Run(ctx)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", step.Check.Name(), err)
		}
		all = append(all, normalize(got, step.Cap)...)
	}
	return all, nil
}

// normalize caps the level of each finding and orders them by place.
func normalize(got []findings.Finding, limit findings.Level) []findings.Finding {
	out := make([]findings.Finding, len(got))
	for i, f := range got {
		if f.Level.Rank() > limit.Rank() {
			f.Level = limit
		}
		out[i] = f
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Line < b.Line
	})
	return out
}
