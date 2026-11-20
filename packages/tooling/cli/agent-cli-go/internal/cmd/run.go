package cmd

import (
	"errors"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// newRunCommand binds the run flags into launch.Options and executes them
// through deps.
func newRunCommand(deps launch.Dependencies) *cobra.Command {
	options := defaultOptions()
	cmd := &cobra.Command{
		Use:   "run [agent-args...]",
		Short: "Run an AI coding agent",
		Long:  `Run an AI coding agent with specified provider and sandbox options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			run := options
			run.Verbose, _ = cmd.Flags().GetBool("verbose")
			run.Explicit = chosenKeys(cmd)
			run.AgentArgs = args
			err := launch.Execute(deps, run)
			// The agent reported its own failure; the CLI passes on only the code.
			var exit launch.ExitError
			if errors.As(err, &exit) {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
			}
			return err
		},
	}

	bindRunFlags(cmd, &options, deps)
	cmd.Flags().BoolVarP(&options.DryRun, "dry-run", "n", false, "Show configuration without executing")
	return cmd
}

// defaultOptions are the options a run holds before any flag or layer changes
// them.
func defaultOptions() launch.Options {
	return launch.Options{
		Agent:     "claude",
		Isolation: "none",
		Provision: "none",
		Network:   "host",
		NixSource: "packages",
		NixShell:  "default",
		VCS:       "git",
	}
}

// chosenKeys are the configuration keys whose flags the user set. A flag that
// keeps its default must yield to the file layers.
func chosenKeys(cmd *cobra.Command) runconfig.Keys {
	chosen := runconfig.Keys{}
	for flag, key := range map[string]runconfig.Key{
		"agent":     runconfig.KeyAgent,
		"isolation": runconfig.KeyIsolation,
		"provision": runconfig.KeyProvision,
		"network":   runconfig.KeyNetwork,
	} {
		if cmd.Flags().Changed(flag) {
			chosen[key] = struct{}{}
		}
	}
	return chosen
}

// bindRunFlags binds the flags that describe a run, which `run` and `preflight`
// share, into options and registers their completions.
func bindRunFlags(cmd *cobra.Command, options *launch.Options, deps launch.Dependencies) {
	flags := cmd.Flags()
	// Bare axis selectors.
	flags.StringVarP(&options.Agent, "agent", "a", options.Agent, "Agent to run ("+strings.Join(agentNames(), ", ")+")")
	flags.StringVarP(&options.Provider, "provider", "p", "", "Model provider, by the name of a provider file (see agent-cli providers list)")
	flags.StringArrayVar(&options.Models, "model", nil, "Model slot override of the provider, as slot=id (repeatable)")
	flags.StringVar(&options.Isolation, "isolation", options.Isolation, "Isolation axis (none, bwrap, docker, seatbelt, docker-sandboxes, openshell)")
	flags.StringVar(&options.Provision, "provision", options.Provision, "Provision axis (none, nix, command)")
	flags.StringVar(&options.Network, "network", options.Network, "Network egress axis (host, none, managed)")
	flags.StringSliceVar(&options.AllowHosts, "allow-host", nil, "Host a managed network run may reach: host, host:port or *.domain (repeatable; denies all others)")
	flags.StringSliceVar(&options.Credentials, "credential", nil, "Inject a credential as an environment variable: NAME=scheme:locator, scheme one of env, file, keychain, secret-service, pass (repeatable)")

	// Per-type knobs under the --<axis>-<type>-<option> grammar.
	flags.StringVar(&options.IsolationDockerRuntime, "isolation-docker-runtime", "", "Kernel-isolating container runtime, e.g. runsc (docker only)")
	flags.BoolVar(&options.IsolationBwrapPassthrough, "isolation-bwrap-passthrough", false, "Expose host/base-image tools as a fallback (bwrap only; forfeits host-tools-unreachable)")
	flags.StringSliceVar(&options.InitCommands, "init-command", nil, "Init command to run before the agent for --provision command (repeatable)")
	flags.StringVar(&options.NixSource, "nix-source", options.NixSource, "Nix source for --provision nix (packages, flake)")
	flags.StringVar(&options.NixRev, "nix-rev", "", "Pinned nixpkgs rev for --nix-source packages")
	flags.StringSliceVar(&options.NixPackages, "nix-packages", nil, "Additional packages for --nix-source packages; selected agent is automatic (repeatable)")
	flags.StringVar(&options.NixShell, "nix-shell", options.NixShell, "devShell name for --nix-source flake")
	flags.StringVar(&options.Image, "image", "", "Container image (for docker and openshell isolation)")

	// Workspace / terminal / misc.
	flags.StringVarP(&options.WorkDir, "work-dir", "w", "", "Working directory")
	flags.StringVar(&options.Project, "project", "", "Project agentic.json (default: the nearest agentic.json from the work directory up to the repository root)")
	flags.StringVar(&options.ApproveDigest, trust.ApproveFlag, "", "Approve the project file with this digest (sha256:...) without a question")
	flags.StringVar(&options.WorktreeBranch, "worktree-branch", "", "Create worktree with branch name")
	flags.StringVar(&options.WorktreeSourceBranch, "worktree-source-branch", "", "Source branch for worktree")
	flags.StringVar(&options.WorktreeDir, "worktree-dir", "", "Worktree directory path")
	flags.StringSliceVar(&options.BindPaths, "bind", nil, "Read-write bind mount")
	flags.StringSliceVar(&options.RoBindPaths, "ro-bind", nil, "Read-only bind mount")
	flags.StringSliceVar(&options.Plugins, "plugin", nil, "Installed plugin to overlay on the run (repeatable; see agent-cli plugin)")
	flags.StringSliceVar(&options.OverlayDirs, "overlay-dir", nil, "Directory of skills, agents, commands and AGENTS.md to overlay on the run (repeatable)")
	flags.StringVar(&options.Store, "store", "", "Plugin store directory (default: $XDG_DATA_HOME/agentic/plugins)")
	flags.StringSliceVar(&options.CustomEnv, "env", nil, "Environment variables (KEY=VALUE)")
	flags.StringVarP(&options.Terminal, "terminal", "t", "", "Terminal wrapper (tmux, herdr)")
	flags.StringVar(&options.TerminalSession, "terminal-session", "", "Custom tmux session name (tmux only)")
	flags.StringVar(&options.TerminalWindow, "terminal-window", "", "Custom tmux window name (tmux only)")
	flags.BoolVar(&options.TerminalDetach, "terminal-detach", false, "Run in background (detach from terminal; herdr keeps the focus where it is)")
	flags.StringVar(&options.VCS, "vcs", options.VCS, "VCS type for worktree (git, jj)")

	// Independent policy guarantees.
	flags.BoolVar(&options.Policy.RequirePinnedProvision, "require-pinned-provision", false,
		"Require provisioning from a pinned source")
	flags.BoolVar(&options.Policy.RequireHostToolsUnreachable, "require-host-tools-unreachable", false,
		"Require host tools to be unreachable from the sandbox")
	flags.BoolVar(&options.Policy.RequireEgressRestricted, "require-egress-restricted", false,
		"Require network egress to be disabled or enforced by a supported transport")
	flags.BoolVar(&options.Policy.RequireKernelIsolation, "require-kernel-isolation", false,
		"Require a kernel-isolating runtime such as docker with runsc")

	// The flags exist above, so registration cannot fail.
	_ = cmd.RegisterFlagCompletionFunc("agent", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return agentNames(), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("provider", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return definedProviders(deps, *options, options.Agent), cobra.ShellCompDirectiveNoFileComp
	})

	_ = cmd.RegisterFlagCompletionFunc("isolation", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return isolationNames(deps), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("network", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{string(network.NetworkHost), string(network.NetworkNone), string(network.NetworkManaged)}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("terminal", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return terminalNames(deps), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("credential", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return credentialCandidates(deps, toComplete), cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	})
}

// isolationNames lists the isolations the build registers.
func isolationNames(deps launch.Dependencies) []string {
	if deps.Sandboxes == nil {
		return nil
	}
	var names []string
	for _, method := range deps.Sandboxes.IsolationMethods() {
		names = append(names, string(method))
	}
	return names
}

// terminalNames lists the terminal wrappers the build registers.
func terminalNames(deps launch.Dependencies) []string {
	names := make([]string, 0, len(deps.Terminals))
	for name := range deps.Terminals {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return names
}

// credentialCandidates completes NAME=scheme: once the user typed the variable
// name and the equals sign, and offers nothing before it, because the variable
// name is the user's own.
func credentialCandidates(deps launch.Dependencies, toComplete string) []string {
	if deps.Credentials == nil {
		return nil
	}
	name, _, hasEquals := strings.Cut(toComplete, "=")
	if !hasEquals {
		return nil
	}
	var candidates []string
	for _, scheme := range deps.Credentials.Schemes() {
		candidates = append(candidates, name+"="+scheme+":")
	}
	return candidates
}

// agentNames lists the registered agent types in sorted order.
func agentNames() []string {
	names := make([]string, 0, len(agents.GetAgentTypes()))
	for _, agentType := range agents.GetAgentTypes() {
		names = append(names, string(agentType))
	}
	return names
}
