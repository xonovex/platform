package launch

import (
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// targets are the executors a plan runs through.
type targets struct {
	// terminal is nil for a run without a terminal wrapper.
	terminal termshared.TerminalExecutor
	// vcs is nil for a run without a checkout.
	vcs wsshared.VCS
}

// selectTargets checks that the isolation of the plan is installed and selects
// its terminal and its workspace variant, before the run builds anything.
func selectTargets(deps Dependencies, plan Plan) (targets, error) {
	if err := checkIsolation(deps.Sandboxes, plan); err != nil {
		return targets{}, err
	}
	terminal, err := selectTerminal(deps.Terminals, plan.Terminal)
	if err != nil {
		return targets{}, err
	}
	vcs, err := selectWorkspace(deps.Workspaces, plan.Checkout)
	if err != nil {
		return targets{}, err
	}
	return targets{terminal: terminal, vcs: vcs}, nil
}
