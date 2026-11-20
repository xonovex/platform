package launch

import (
	"fmt"
	"io"
	"sort"
	"strings"

	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// Host is the port the executor consumes to locate the run: the working
// directory, the home directory and the configuration file.
type Host interface {
	config.Files
	Getwd() (string, error)
}

// Dependencies are the ports and plugin sets one run executes through.
type Dependencies struct {
	Host Host
	Log  logging.Logger
	// Out receives the command line of a dry run.
	Out        io.Writer
	Sandboxes  *sandbox.Registry
	Workspaces map[wsp.VCSType]wsshared.VCS
	Terminals  map[termshared.TerminalType]termshared.TerminalExecutor
}

// ExitError reports that the agent ran and exited with a non-zero code, which
// the CLI passes on as its own exit code.
type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("agent exited with code %d", e.Code) }

// Execute plans the run from options and carries it out. Every check that can
// fail runs before the checkout is created, and the sandbox command is built
// before the dispatch, so a dry run, a terminal run and a direct run reject the
// same invalid inputs.
func Execute(deps Dependencies, options Options) error {
	file, err := config.Load(deps.Host, options.Config)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cwd, err := deps.Host.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current working directory: %w", err)
	}
	home := file.HomeDir
	if home == "" {
		if home, err = deps.Host.UserHomeDir(); err != nil {
			return fmt.Errorf("resolve user home directory: %w", err)
		}
	}

	plan, err := NewPlan(options, file, cwd, home, deps.Sandboxes)
	if err != nil {
		return err
	}
	if available, err := plan.Isolator.Available(); err != nil || !available {
		return fmt.Errorf("isolation method %s is not available; available isolations: %v", plan.IsolationName, sandbox.AvailableIsolations(deps.Sandboxes))
	}
	terminal, err := selectTerminal(deps.Terminals, plan.Terminal)
	if err != nil {
		return err
	}
	vcs, err := selectWorkspace(deps.Workspaces, plan.Checkout)
	if err != nil {
		return err
	}

	run := plan.Run
	if vcs != nil && !plan.DryRun {
		dir, err := vcs.Setup(plan.Checkout.Config, plan.SourceDir, plan.Verbose)
		if err != nil {
			return err
		}
		run.WorkDir = dir
	}

	contribution, err := plan.Provisioner.Contribute(plan.Provision)
	if err != nil {
		return err
	}
	handled, err := plan.Provisioner.CheckAgent(contribution, plan.Agent)
	if err != nil {
		return err
	}
	if !handled {
		if err := plan.Isolator.CheckAgent(run); err != nil {
			return err
		}
	}
	command, err := plan.Isolator.Command(run, contribution)
	if err != nil {
		return fmt.Errorf("prepare sandbox command: %w", err)
	}

	if plan.DryRun {
		return printDryRun(deps.Log, deps.Out, plan, command)
	}

	var code int
	if terminal != nil {
		fullCommand, env, err := plan.Isolator.TerminalCommand(run, contribution)
		if err != nil {
			return fmt.Errorf("prepare terminal command: %w", err)
		}
		if plan.Verbose {
			deps.Log.Info("Using terminal wrapper: " + string(plan.Terminal.Type))
		}
		code, err = terminal.Execute(plan.Terminal, fullCommand, env, run.WorkDir)
		if err != nil {
			return err
		}
	} else {
		code, err = plan.Isolator.Run(run, contribution)
		if err != nil {
			return err
		}
	}
	if code != 0 {
		return ExitError{Code: code}
	}
	return nil
}

// selectTerminal returns the installed executor of the requested terminal, or
// nil when the run uses no terminal wrapper.
func selectTerminal(terminals map[termshared.TerminalType]termshared.TerminalExecutor, cfg *termshared.TerminalConfig) (termshared.TerminalExecutor, error) {
	if cfg == nil {
		return nil, nil
	}
	executor, ok := terminals[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unknown terminal type: %s; available types: %v", cfg.Type, availableTerminals(terminals))
	}
	if !executor.IsAvailable() {
		return nil, fmt.Errorf("terminal type %s is not available (not installed)", cfg.Type)
	}
	return executor, nil
}

func availableTerminals(terminals map[termshared.TerminalType]termshared.TerminalExecutor) []string {
	var names []string
	for name, executor := range terminals {
		if executor.IsAvailable() {
			names = append(names, string(name))
		}
	}
	sort.Strings(names)
	return names
}

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

// printDryRun reports the run and writes the sandbox command line to out.
func printDryRun(log logging.Logger, out io.Writer, plan Plan, command []string) error {
	log.Info("Dry run - would execute:")
	log.Info("  Agent: " + plan.Agent.DisplayName)
	if plan.Provider != nil {
		log.Info("  Provider: " + plan.Provider.DisplayName)
	}
	log.Info("  Work directory: " + plan.DisplayDir())
	if len(plan.Run.AgentArgs) > 0 {
		log.Info(fmt.Sprintf("  Arguments: %v", plan.Run.AgentArgs))
	} else {
		log.Info("  Arguments: (none)")
	}
	if len(command) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, strings.Join(command, " ")); err != nil {
		return fmt.Errorf("write dry run: %w", err)
	}
	return nil
}
