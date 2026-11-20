package launch

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/config"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
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

// EgressOpener starts the enforcement of a managed network run: the allowlist
// proxy for a transport that needs one. It returns the route the isolator
// applies, and a stop function that ends the enforcement.
type EgressOpener interface {
	Open(transport netshared.Transport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error)
}

// Dependencies are the ports and plugin sets one run executes through.
type Dependencies struct {
	Host Host
	Log  logging.Logger
	// Out receives the command line of a dry run.
	Out       io.Writer
	Sandboxes *sandbox.Registry
	// Egress starts managed network enforcement; it may be nil for a build that
	// offers no managed egress, and a managed run then fails.
	Egress EgressOpener
	// Credentials maps a credential scheme to its store; it may be nil for a build
	// without credential stores, and a run that binds a credential then fails.
	Credentials *credshared.Registry
	// Overlays builds the artifact overlay; it may be nil for a build without one.
	Overlays   Overlays
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

	overlayResult, err := prepareOverlay(deps.Overlays, plan)
	if err != nil {
		return err
	}
	// A terminal session outlives this process, so its overlay stays on disk.
	keepOverlay := false
	defer func() {
		if !keepOverlay {
			overlayResult.Cleanup()
		}
	}()

	run := applyOverlay(plan.Run, overlayResult)
	run.Secrets, err = credentialValues(deps.Credentials, plan)
	if err != nil {
		return err
	}
	if plan.Run.Network == netshared.ModeManaged {
		route, stop, err := openEgress(deps.Egress, plan)
		if err != nil {
			return err
		}
		defer stop()
		run.Egress = route
	}
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
		return printDryRun(deps.Log, deps.Out, plan, overlayResult, command)
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
		if overlayResult.Root != "" {
			keepOverlay = true
			deps.Log.Info("The overlay stays at " + overlayResult.Root + " while the terminal session runs; remove it afterwards")
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

// credentialValues resolves the credential bindings of the run. A dry run
// resolves nothing, so it prompts no keychain and holds no secret: it carries the
// redacted placeholder of each credential, which is enough to build the command.
func credentialValues(registry *credshared.Registry, plan Plan) (map[string]string, error) {
	if len(plan.Credentials) == 0 {
		return nil, nil
	}
	if plan.DryRun {
		placeholders := make(map[string]string, len(plan.Credentials))
		for _, binding := range plan.Credentials {
			placeholders[binding.Name] = credshared.Redacted
		}
		return placeholders, nil
	}
	if registry == nil {
		return nil, fmt.Errorf("this build has no credential stores")
	}
	secrets, err := credshared.Resolve(registry, plan.Credentials)
	if err != nil {
		return nil, err
	}
	return credshared.Reveal(secrets), nil
}

// openEgress starts the enforcement of a managed run for the transport the
// isolator declared.
func openEgress(opener EgressOpener, plan Plan) (isoshared.EgressRoute, func(), error) {
	if opener == nil {
		return isoshared.EgressRoute{}, nil, fmt.Errorf("this build has no managed egress")
	}
	route, stop, err := opener.Open(plan.Run.Egress.Transport, plan.Allow)
	if err != nil {
		return isoshared.EgressRoute{}, nil, fmt.Errorf("start managed egress: %w", err)
	}
	return route, stop, nil
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
func printDryRun(log logging.Logger, out io.Writer, plan Plan, overlay OverlayResult, command []string) error {
	log.Info("Dry run - would execute:")
	log.Info("  Agent: " + plan.Agent.DisplayName)
	log.Info("  Protocol: " + describeProtocol(plan.Agent))
	log.Info("  Model access: " + describeAccess(plan.Agent))
	log.Info("  Version probe: " + describeVersionProbe(plan.Agent))
	if plan.Provider != nil {
		log.Info("  Provider: " + plan.Provider.DisplayName)
	}
	log.Info("  Work directory: " + plan.DisplayDir())
	if len(plan.Run.AgentArgs) > 0 {
		log.Info(fmt.Sprintf("  Arguments: %v", plan.Run.AgentArgs))
	} else {
		log.Info("  Arguments: (none)")
	}
	logOverlay(log, overlay)
	logNetwork(log, plan)
	if len(command) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, strings.Join(command, " ")); err != nil {
		return fmt.Errorf("write dry run: %w", err)
	}
	return nil
}

// logNetwork reports the network mode, the allowed hosts and the credential
// bindings of a dry run. A binding shows where the credential lives and never what
// it holds.
func logNetwork(log logging.Logger, plan Plan) {
	log.Info("  Network: " + string(plan.Run.Network))
	if plan.Run.Network == netshared.ModeManaged {
		var hosts []string
		for _, rule := range plan.Allow.Rules() {
			hosts = append(hosts, rule.String())
		}
		log.Info("  Allowed hosts: " + strings.Join(hosts, ", "))
		log.Info("  Egress transport: " + string(plan.Run.Egress.Transport))
	}
	for _, binding := range plan.Credentials {
		log.Info("  Credential: " + binding.Name + " from " + binding.Scheme + ":" + binding.Locator + " (" + credshared.Redacted + ")")
	}
}
