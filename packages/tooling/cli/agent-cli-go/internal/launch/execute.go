package launch

import (
	"fmt"
)

// ExitError reports that the agent ran and exited with a non-zero code, which
// the CLI passes on as its own exit code.
type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("agent exited with code %d", e.Code) }

// launched is what the start of the agent did: whether the agent, or its
// terminal wrapper, ran to an exit code, that code, and whether the program
// runs on in a terminal session after the start returns.
type launched struct {
	started bool
	code    int
	runsOn  bool
}

// outcome is what a run did: the start of the agent and the sources of the
// overlay it ran over.
type outcome struct {
	launched
	overlay []OverlaySource
}

// Execute resolves the configuration layers and plans the run from options, then
// carries it out. Every check that can fail runs before the checkout is created,
// and the sandbox command is built before the dispatch, so a dry run, a terminal
// run and a direct run reject the same invalid inputs. A run that is not a dry
// run leaves a run record.
func Execute(deps Dependencies, options Options) error {
	res, err := resolve(deps, options, false)
	if err != nil {
		return err
	}
	if res.planErr != nil {
		return res.planErr
	}
	if res.plan.Terminal, err = resolveAutoTerminal(deps, res.plan.Terminal); err != nil {
		return err
	}
	if res.plan.DryRun {
		_, err := carryOut(deps, res)
		return err
	}
	started := deps.Host.Now()
	runID := nextRunID(deps, started)
	res.plan.Terminal = withRunID(res.plan.Terminal, runID)
	result, err := carryOut(deps, res)
	writeRecord(deps, res, runID, started, result, err)
	return err
}

// carryOut runs the plan: it selects the executors, then builds the overlay and
// starts the agent over it. Its outcome is meaningful even when it returns an
// error.
func carryOut(deps Dependencies, res resolution) (outcome, error) {
	chosen, err := selectTargets(deps, res.plan)
	if err != nil {
		return outcome{}, err
	}
	return withOverlay(deps.Overlays, deps.Log, res.plan.Overlay, func(overlay OverlayResult) (launched, error) {
		return launchOverlaid(deps, res, chosen, overlay)
	})
}

// launchOverlaid completes the run config of the plan over the overlay, step by
// step, and starts the agent, or reports the command of a dry run.
func launchOverlaid(deps Dependencies, res resolution, chosen targets, overlay OverlayResult) (launched, error) {
	plan := res.plan
	run := applyOverlay(plan.Run, plan.AgentArgs, overlay)
	secrets, err := credentialValues(deps.Credentials, plan)
	if err != nil {
		return launched{}, err
	}
	run.Secrets = secrets
	route, stop, err := openEgress(deps.Egress, plan)
	if err != nil {
		return launched{}, err
	}
	defer stop()
	run.Egress = route
	if run.Process.Dir, err = setupCheckout(chosen.vcs, plan, run.Process.Dir); err != nil {
		return launched{}, err
	}
	sandboxed, err := confine(deps.Host.Getenv, plan, run)
	if err != nil {
		return launched{}, err
	}
	if plan.DryRun {
		return launched{}, printDryRun(deps.Log, deps.Out, plan, overlay, res.config.Provenance(), sandboxed.command)
	}
	if chosen.terminal != nil {
		return startInTerminal(deps.Log, plan, chosen.terminal, sandboxed)
	}
	return startDirect(plan, sandboxed)
}
