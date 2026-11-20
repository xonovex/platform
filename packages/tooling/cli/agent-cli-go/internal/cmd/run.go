package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// newRunCommand binds the run flags into launch.Options and executes them
// through deps.
func newRunCommand(deps launch.Dependencies) *cobra.Command {
	options := launch.Options{
		Agent:     "claude",
		Isolation: "none",
		Provision: "none",
		Network:   "host",
		NixSource: "packages",
		NixShell:  "default",
		VCS:       "git",
	}
	cmd := &cobra.Command{
		Use:   "run [agent-args...]",
		Short: "Run an AI coding agent",
		Long:  `Run an AI coding agent with specified provider and sandbox options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			run := options
			run.Verbose, _ = cmd.Flags().GetBool("verbose")
			run.IsolationSet = cmd.Flags().Changed("isolation")
			run.ProvisionSet = cmd.Flags().Changed("provision")
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

	flags := cmd.Flags()
	// Bare axis selectors.
	flags.StringVarP(&options.Agent, "agent", "a", options.Agent, "Agent to run (claude, opencode)")
	flags.StringVarP(&options.Provider, "provider", "p", "", "Model provider")
	flags.StringVar(&options.Isolation, "isolation", options.Isolation, "Isolation axis (none, bwrap, docker)")
	flags.StringVar(&options.Provision, "provision", options.Provision, "Provision axis (none, nix, command)")
	flags.StringVar(&options.Network, "network", options.Network, "Network egress axis (host, none)")

	// Per-type knobs under the --<axis>-<type>-<option> grammar.
	flags.StringVar(&options.IsolationDockerRuntime, "isolation-docker-runtime", "", "Kernel-isolating container runtime, e.g. runsc (docker only)")
	flags.BoolVar(&options.IsolationBwrapPassthrough, "isolation-bwrap-passthrough", false, "Expose host/base-image tools as a fallback (bwrap only; forfeits host-tools-unreachable)")
	flags.StringSliceVar(&options.InitCommands, "init-command", nil, "Init command to run before the agent for --provision command (repeatable)")
	flags.StringVar(&options.NixSource, "nix-source", options.NixSource, "Nix source for --provision nix (packages, flake)")
	flags.StringVar(&options.NixRev, "nix-rev", "", "Pinned nixpkgs rev for --nix-source packages")
	flags.StringSliceVar(&options.NixPackages, "nix-packages", nil, "Additional packages for --nix-source packages; selected agent is automatic (repeatable)")
	flags.StringVar(&options.NixShell, "nix-shell", options.NixShell, "devShell name for --nix-source flake")
	flags.StringVar(&options.Image, "image", "", "Container image (for docker isolation)")

	// Workspace / terminal / misc.
	flags.StringVarP(&options.WorkDir, "work-dir", "w", "", "Working directory")
	flags.StringVar(&options.WorktreeBranch, "worktree-branch", "", "Create worktree with branch name")
	flags.StringVar(&options.WorktreeSourceBranch, "worktree-source-branch", "", "Source branch for worktree")
	flags.StringVar(&options.WorktreeDir, "worktree-dir", "", "Worktree directory path")
	flags.StringVarP(&options.Config, "config", "c", "", "Configuration file")
	flags.StringSliceVar(&options.BindPaths, "bind", nil, "Read-write bind mount")
	flags.StringSliceVar(&options.RoBindPaths, "ro-bind", nil, "Read-only bind mount")
	flags.StringSliceVar(&options.CustomEnv, "env", nil, "Environment variables (KEY=VALUE)")
	flags.StringVarP(&options.Terminal, "terminal", "t", "", "Terminal wrapper (tmux)")
	flags.StringVar(&options.TerminalSession, "terminal-session", "", "Custom tmux session name")
	flags.StringVar(&options.TerminalWindow, "terminal-window", "", "Custom tmux window name")
	flags.BoolVar(&options.TerminalDetach, "terminal-detach", false, "Run in background (detach from terminal)")
	flags.StringVar(&options.VCS, "vcs", options.VCS, "VCS type for worktree (git, jj)")
	flags.BoolVarP(&options.DryRun, "dry-run", "n", false, "Show configuration without executing")

	// Independent policy guarantees.
	flags.BoolVar(&options.Policy.RequirePinnedProvision, "require-pinned-provision", false,
		"Require provisioning from a pinned source")
	flags.BoolVar(&options.Policy.RequireHostToolsUnreachable, "require-host-tools-unreachable", false,
		"Require host tools to be unreachable from the sandbox")
	flags.BoolVar(&options.Policy.RequireEgressRestricted, "require-egress-restricted", false,
		"Require network egress to be disabled or enforced by a supported transport")
	flags.BoolVar(&options.Policy.RequireKernelIsolation, "require-kernel-isolation", false,
		"Require a kernel-isolating runtime such as docker with runsc")

	return cmd
}
