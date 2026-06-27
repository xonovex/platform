package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/plugins"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/worktree"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/wrapper"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/config"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
	sharedworktree "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/worktree"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

var runCmd = &cobra.Command{
	Use:   "run [agent-args...]",
	Short: "Run an AI coding agent",
	Long:  `Run an AI coding agent with specified provider and sandbox options.`,
	RunE:  runAgent,
}

var (
	flagAgent                string
	flagProvider             string
	flagIsolation            string
	flagProvision            string
	flagNetwork              string
	flagEgressAllow          []string
	flagHostPassthrough      bool
	flagInitCommand          []string
	flagNixSource            string
	flagNixRev               string
	flagNixPackages          []string
	flagNixShell             string
	flagWorkDir              string
	flagWorktreeBranch       string
	flagWorktreeSourceBranch string
	flagWorktreeDir          string
	flagConfig               string
	flagBind                 []string
	flagRoBind               []string
	flagEnv                  []string
	flagImage                string
	flagTerminal             string
	flagTerminalSession      string
	flagTerminalWindow       string
	flagTerminalDetach       bool
	flagVCS                  string
	flagDryRun               bool
	flagRequirePinned        bool
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&flagAgent, "agent", "a", "claude", "Agent to run (claude, opencode)")
	runCmd.Flags().StringVarP(&flagProvider, "provider", "p", "", "Model provider")
	runCmd.Flags().StringVar(&flagIsolation, "isolation", "none", "Isolation axis (none, bwrap, docker)")
	runCmd.Flags().StringVar(&flagProvision, "provision", "none", "Provision axis (none, nix, command)")
	runCmd.Flags().StringVar(&flagNetwork, "network", "host", "Network egress axis (host, none, proxy)")
	runCmd.Flags().StringSliceVar(&flagEgressAllow, "egress-allow", []string{}, "Extra egress allowlist hosts for --network proxy (repeatable)")
	runCmd.Flags().BoolVar(&flagHostPassthrough, "host-passthrough", false, "Expose host/base-image tools as a fallback (forfeits host-tools-unreachable)")
	runCmd.Flags().StringSliceVar(&flagInitCommand, "init-command", []string{}, "Init command to run before the agent for --provision command (repeatable)")
	runCmd.Flags().StringVar(&flagNixSource, "nix-source", "packages", "Nix source for --provision nix (packages, flake)")
	runCmd.Flags().StringVar(&flagNixRev, "nix-rev", "", "Pinned nixpkgs rev for --nix-source packages")
	runCmd.Flags().StringSliceVar(&flagNixPackages, "nix-packages", []string{}, "Packages for --nix-source packages (repeatable)")
	runCmd.Flags().StringVar(&flagNixShell, "nix-shell", "default", "devShell name for --nix-source flake")
	runCmd.Flags().StringVarP(&flagWorkDir, "work-dir", "w", "", "Working directory")
	runCmd.Flags().StringVar(&flagWorktreeBranch, "worktree-branch", "", "Create worktree with branch name")
	runCmd.Flags().StringVar(&flagWorktreeSourceBranch, "worktree-source-branch", "", "Source branch for worktree")
	runCmd.Flags().StringVar(&flagWorktreeDir, "worktree-dir", "", "Worktree directory path")
	runCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "Configuration file")
	runCmd.Flags().StringSliceVar(&flagBind, "bind", []string{}, "Read-write bind mount")
	runCmd.Flags().StringSliceVar(&flagRoBind, "ro-bind", []string{}, "Read-only bind mount")
	runCmd.Flags().StringSliceVar(&flagEnv, "env", []string{}, "Environment variables (KEY=VALUE)")
	runCmd.Flags().StringVar(&flagImage, "image", "", "Container image (for docker isolation)")
	runCmd.Flags().StringVarP(&flagTerminal, "terminal", "t", "", "Terminal wrapper (tmux)")
	runCmd.Flags().StringVar(&flagTerminalSession, "terminal-session", "", "Custom tmux session name")
	runCmd.Flags().StringVar(&flagTerminalWindow, "terminal-window", "", "Custom tmux window name")
	runCmd.Flags().BoolVar(&flagTerminalDetach, "terminal-detach", false, "Run in background (detach from terminal)")
	runCmd.Flags().StringVar(&flagVCS, "vcs", "git", "VCS type for worktree (git, jj)")
	runCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "n", false, "Show configuration without executing")
	runCmd.Flags().BoolVar(&flagRequirePinned, "require-pinned-toolchain", false,
		"Mandate pinned provisioning + host-tools-unreachable; reject leaky/host-exposed cells")
}

func runAgent(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	agent, err := agents.GetAgent(types.AgentType(flagAgent))
	if err != nil {
		scriptlib.LogError(err.Error())
		scriptlib.LogInfo("Available agents: " + fmt.Sprint(agents.GetAgentTypes()))
		return err
	}

	fileConfig, err := config.LoadConfigFile(flagConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var provider *types.ModelProvider
	providerName := flagProvider
	if providerName == "" && fileConfig.Provider != "" {
		providerName = fileConfig.Provider
	}
	if providerName != "" {
		provider, err = providers.GetProvider(providerName, agent.Type)
		if err != nil {
			scriptlib.LogError(err.Error())
			scriptlib.LogInfo("Available providers: " + fmt.Sprint(providers.GetProviderNames(agent.Type)))
			return err
		}
	}

	workDir := flagWorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	workDir, _ = filepath.Abs(workDir)

	var sourceRepoDir string
	if flagWorktreeBranch != "" {
		if err := validation.ValidateBranch(flagWorktreeBranch); err != nil {
			return fmt.Errorf("invalid --worktree-branch: %w", err)
		}
		if flagWorktreeSourceBranch != "" {
			if err := validation.ValidateBranch(flagWorktreeSourceBranch); err != nil {
				return fmt.Errorf("invalid --worktree-source-branch: %w", err)
			}
		}
		vcsType := sharedworktree.VCSType(flagVCS)
		if !vcsType.IsValid() {
			return fmt.Errorf("unknown --vcs %q; valid values: git, jj", flagVCS)
		}

		repoName := filepath.Base(workDir)
		wtDir := flagWorktreeDir
		if wtDir == "" {
			wtDir = sharedworktree.GetDefaultDir(flagWorktreeBranch, repoName)
		}

		newWorkDir, err := worktree.Setup(worktree.Config{
			SourceBranch: flagWorktreeSourceBranch,
			Branch:       flagWorktreeBranch,
			Dir:          wtDir,
			VCS:          vcsType,
		}, workDir, verbose)
		if err != nil {
			return err
		}
		sourceRepoDir = workDir
		workDir = newWorkDir
	}

	bindPaths := worktree.BuildBindPaths(flagBind, sourceRepoDir)

	// --require-pinned-toolchain mandates pinned provisioning + host-tools-unreachable.
	policy := types.SandboxPolicy{
		RequirePinnedProvision:      flagRequirePinned,
		RequireHostToolsUnreachable: flagRequirePinned,
	}

	iso, prov, net, passthrough, err := resolveAxes(cmd, policy)
	if err != nil {
		return err
	}

	// runtime is the container runtime (e.g. runsc for gVisor); default runc gives
	// no kernel boundary. No flag wires a sandboxed runtime in this build.
	const runtime = ""

	reg := plugins.DefaultRegistry()
	req := sandbox.Request{
		Isolation:   iso,
		Provision:   prov,
		Network:     net,
		Passthrough: passthrough,
		Runtime:     runtime,
		Image:       flagImage,
	}
	isolator, provisioner, err := sandbox.Select(reg, req, policy)
	if err != nil {
		return err
	}

	available, err := isolator.Available()
	if err != nil || !available {
		scriptlib.LogError(fmt.Sprintf("Isolation method %s is not available", iso))
		scriptlib.LogInfo(fmt.Sprintf("Available isolations: %v", sandbox.AvailableIsolations(reg)))
		return fmt.Errorf("isolation method %s is not available", iso)
	}

	sandboxConfig := &types.SandboxConfig{
		Isolation:           iso,
		Provision:           prov,
		Network:             net,
		HostPassthrough:     passthrough,
		EgressAllowlist:     append(append([]string{}, types.DefaultEgressAllowlist...), flagEgressAllow...),
		Policy:              policy,
		Agent:               agent,
		Provider:            provider,
		WorkDir:             workDir,
		RepoDir:             sourceRepoDir,
		BindPaths:           bindPaths,
		RoBindPaths:         flagRoBind,
		CustomEnv:           flagEnv,
		AgentArgs:           args,
		SandboxInitCommands: flagInitCommand,
		Verbose:             verbose,
		Image:               flagImage,
		NixSourceKind:       flagNixSource,
		NixRev:              flagNixRev,
		NixPackages:         flagNixPackages,
		NixShell:            flagNixShell,
	}

	contribution, err := provisioner.Contribute(sandboxConfig)
	if err != nil {
		return err
	}

	if flagDryRun {
		return printDryRun(iso, isolator, sandboxConfig, contribution, agent, provider, args, workDir)
	}

	if flagTerminal != "" {
		return executeWithTerminal(iso, isolator, sandboxConfig, contribution, agent, provider, args, workDir, verbose)
	}

	exitCode, err := isolator.Run(sandboxConfig, contribution)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

// resolveAxes determines the three axes from the flags. Isolation and provision
// are passed through verbatim — the registry is the source of truth for valid
// methods (Select rejects unregistered ones), keeping the method set pluggable.
// Network is a closed enum, validated here.
func resolveAxes(cmd *cobra.Command, policy types.SandboxPolicy) (types.IsolationMethod, types.ProvisionMethod, types.NetworkMethod, bool, error) {
	net, err := parseNetwork(flagNetwork)
	if err != nil {
		return "", "", "", false, err
	}

	// Mandate the pinned combo when a pinned toolchain is required but no cell was
	// chosen explicitly.
	if policy.RequirePinnedProvision && !cmd.Flags().Changed("isolation") && !cmd.Flags().Changed("provision") {
		return types.IsolationBwrap, types.ProvisionNix, net, flagHostPassthrough, nil
	}

	return types.IsolationMethod(flagIsolation), types.ProvisionMethod(flagProvision), net, flagHostPassthrough, nil
}

func parseNetwork(s string) (types.NetworkMethod, error) {
	switch types.NetworkMethod(s) {
	case types.NetworkHost, types.NetworkNone, types.NetworkProxy:
		return types.NetworkMethod(s), nil
	default:
		return "", fmt.Errorf("unknown --network %q; valid: host, none, proxy", s)
	}
}

// executeWithTerminal handles execution when a terminal wrapper is requested.
func executeWithTerminal(iso types.IsolationMethod, isolator sandbox.Isolator, sandboxConfig *types.SandboxConfig, contribution types.Contribution, agent *types.AgentConfig, provider *types.ModelProvider, args []string, workDir string, verbose bool) error {
	terminalType := types.TerminalType(flagTerminal)
	terminalExecutor := wrapper.GetExecutor(terminalType)
	if terminalExecutor == nil {
		availableTypes := wrapper.GetAvailableTypes()
		scriptlib.LogError(fmt.Sprintf("Unknown terminal type: %s", flagTerminal))
		scriptlib.LogInfo(fmt.Sprintf("Available types: %v", availableTypes))
		return fmt.Errorf("unknown terminal type: %s", flagTerminal)
	}
	if !terminalExecutor.IsAvailable() {
		scriptlib.LogError(fmt.Sprintf("Terminal type %s is not available (not installed)", flagTerminal))
		return fmt.Errorf("terminal type %s is not available", flagTerminal)
	}

	terminalConfig := &types.TerminalConfig{
		Type:        terminalType,
		SessionName: flagTerminalSession,
		WindowName:  flagTerminalWindow,
		Detach:      flagTerminalDetach,
	}

	var fullCommand []string
	var env []string
	if iso == types.IsolationNone {
		fullCommand = buildDirectCommand(agent, provider, args)
		env = buildDirectEnv(agent, provider, sandboxConfig.CustomEnv)
	} else {
		fullCommand = isolator.Command(sandboxConfig, contribution)
		env = os.Environ()
		env = append(env, sandboxConfig.CustomEnv...)
	}

	if verbose {
		scriptlib.LogInfo("Using terminal wrapper: " + flagTerminal)
	}

	exitCode, err := terminalExecutor.Execute(terminalConfig, fullCommand, env, workDir, verbose)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

// buildDirectCommand builds the command for direct (no-isolation) execution.
func buildDirectCommand(agent *types.AgentConfig, provider *types.ModelProvider, args []string) []string {
	var providerCliArgs []string
	if provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(provider)
	}
	execOpts := types.AgentExecOptions{Sandbox: false, ProviderCliArgs: providerCliArgs}

	var agentArgs []string
	switch agent.Type {
	case types.AgentClaude:
		agentArgs = agents.BuildClaudeArgs(args, execOpts)
	case types.AgentOpencode:
		agentArgs = agents.BuildOpencodeArgs(args, execOpts)
	}

	cmd := make([]string, 0, 1+len(agentArgs))
	cmd = append(cmd, agent.Binary)
	cmd = append(cmd, agentArgs...)
	return cmd
}

// buildDirectEnv builds the environment for direct (no-isolation) execution.
func buildDirectEnv(agent *types.AgentConfig, provider *types.ModelProvider, customEnv []string) []string {
	var providerEnv map[string]string
	if provider != nil {
		providerEnv, _ = providers.BuildProviderEnv(provider)
	}

	var agentEnv map[string]string
	switch agent.Type {
	case types.AgentClaude:
		agentEnv = agents.BuildClaudeEnv(providerEnv)
	case types.AgentOpencode:
		agentEnv = agents.BuildOpencodeEnv(providerEnv)
	}

	env := os.Environ()
	for k, v := range agentEnv {
		env = append(env, k+"="+v)
	}
	env = append(env, customEnv...)
	return env
}

// printDryRun displays what would be executed without running it.
func printDryRun(iso types.IsolationMethod, isolator sandbox.Isolator, sandboxConfig *types.SandboxConfig, contribution types.Contribution, agent *types.AgentConfig, provider *types.ModelProvider, args []string, workDir string) error {
	scriptlib.LogInfo("Dry run - would execute:")

	if iso == types.IsolationNone {
		scriptlib.LogInfo("  Agent: " + agent.DisplayName)
		if provider != nil {
			scriptlib.LogInfo("  Provider: " + provider.DisplayName)
		}
		scriptlib.LogInfo("  Work directory: " + workDir)
		if len(args) > 0 {
			scriptlib.LogInfo("  Arguments: " + fmt.Sprintf("%v", args))
		} else {
			scriptlib.LogInfo("  Arguments: (none)")
		}
	}

	command := isolator.Command(sandboxConfig, contribution)
	if len(command) > 0 {
		fmt.Println(strings.Join(command, " "))
	}
	return nil
}
