package launch

import (
	"fmt"

	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// selectWorkspace returns the variant for the planned checkout, or nil when the
// run creates no checkout.
func selectWorkspace(workspaces map[wsp.VCSType]wsshared.VCS, checkout *Checkout) (wsshared.VCS, error) {
	if checkout == nil {
		return nil, nil
	}
	vcs, ok := workspaces[checkout.VCS]
	if !ok {
		return nil, fmt.Errorf("unknown --vcs %q", checkout.VCS)
	}
	return vcs, nil
}

// setupCheckout creates the planned checkout through vcs and returns the
// directory the agent starts in: the checkout, or dir for a run without one and
// for a dry run, which creates nothing.
func setupCheckout(vcs wsshared.VCS, plan Plan, dir string) (string, error) {
	if vcs == nil || plan.DryRun {
		return dir, nil
	}
	return vcs.Setup(plan.Checkout.Config, plan.SourceDir, plan.Verbose)
}
