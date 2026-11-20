package launch

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	agentprovision "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
)

// checkIsolation fails when the isolation of the plan is not installed on this
// machine, and names the ones that are.
func checkIsolation(sandboxes *sandbox.Registry[isoshared.Isolator], plan Plan) error {
	if available, err := plan.Isolator.Available(); err != nil || !available {
		return fmt.Errorf("isolation method %s is not available; available isolations: %v", plan.IsolationName, sandbox.AvailableIsolations(sandboxes))
	}
	return nil
}

// confined is a run the isolation of its plan accepts: the complete run config,
// the contribution of the provisioner and the sandbox command line.
type confined struct {
	run          isoshared.RunConfig
	contribution agentprovision.Contribution
	command      []string
}

// confine applies the provisioner of the plan to run, checks that the agent is
// reachable where the provisioner leaves it, resolves the provider environment
// from getenv and builds the sandbox command. It starts nothing.
func confine(getenv providers.Getenv, plan Plan, run isoshared.RunConfig) (confined, error) {
	contribution, err := plan.Provisioner.Contribute(plan.Provision)
	if err != nil {
		return confined{}, err
	}
	handled, err := plan.Provisioner.CheckAgent(contribution, plan.Agent)
	if err != nil {
		return confined{}, err
	}
	if !handled {
		if err := plan.Isolator.CheckAgent(run); err != nil {
			return confined{}, err
		}
	}
	if run.ProviderEnv, err = agentcmd.PlanProviderEnv(plan.Agent, plan.Provider).Resolve(getenv); err != nil {
		return confined{}, fmt.Errorf("prepare sandbox command: %w", err)
	}
	command, err := plan.Isolator.Command(run, contribution)
	if err != nil {
		return confined{}, fmt.Errorf("prepare sandbox command: %w", err)
	}
	return confined{run: run, contribution: contribution, command: command}, nil
}

// startDirect runs the confined agent in the process of the CLI and waits for
// it.
func startDirect(plan Plan, sandboxed confined) (launched, error) {
	code, err := plan.Isolator.Run(sandboxed.run, sandboxed.contribution)
	if err != nil {
		return launched{}, err
	}
	return exited(launched{code: code})
}

// exited marks a start that ran to its exit code, and turns a non-zero code
// into an ExitError.
func exited(result launched) (launched, error) {
	result.started = true
	if result.code != 0 {
		return result, ExitError{Code: result.code}
	}
	return result, nil
}
