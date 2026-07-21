package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cfgpkg "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/config"
	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	provnix "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/nix"
	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/plugins"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/terminal"
	termshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/terminal/shared"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/git"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/jj"
	wsshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
	wsp "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

type runOptions struct {
	agent                       string
	provider                    string
	isolation                   string
	provision                   string
	network                     string
	isolationBwrapPassthrough   bool
	isolationDockerRuntime      string
	initCommands                []string
	nixSource                   string
	nixRev                      string
	nixPackages                 []string
	nixShell                    string
	workDir                     string
	worktreeBranch              string
	worktreeSourceBranch        string
	worktreeDir                 string
	config                      string
	bindPaths                   []string
	roBindPaths                 []string
	customEnv                   []string
	image                       string
	terminal                    string
	terminalSession             string
	terminalWindow              string
	terminalDetach              bool
	vcs                         string
	dryRun                      bool
	requirePinnedProvision      bool
	requireHostToolsUnreachable bool
	requireEgressRestricted     bool
	requireKernelIsolation      bool
}

func newRunCommand() *cobra.Command {
	options := runOptions{
		agent:     "claude",
		isolation: "none",
		provision: "none",
		network:   "host",
		nixSource: "packages",
		nixShell:  "default",
		vcs:       "git",
	}
	cmd := &cobra.Command{
		Use:   "run [agent-args...]",
		Short: "Run an AI coding agent",
		Long:  `Run an AI coding agent with specified provider and sandbox options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgent(cmd, args, options)
		},
	}

	// Bare axis selectors.
	cmd.Flags().StringVarP(&options.agent, "agent", "a", options.agent, "Agent to run (claude, opencode)")
	cmd.Flags().StringVarP(&options.provider, "provider", "p", "", "Model provider")
	cmd.Flags().StringVar(&options.isolation, "isolation", options.isolation, "Isolation axis (none, bwrap, docker)")
	cmd.Flags().StringVar(&options.provision, "provision", options.provision, "Provision axis (none, nix, command)")
	cmd.Flags().StringVar(&options.network, "network", options.network, "Network egress axis (host, none)")

	// Per-type knobs under the --<axis>-<type>-<option> grammar.
	cmd.Flags().StringVar(&options.isolationDockerRuntime, "isolation-docker-runtime", "", "Kernel-isolating container runtime, e.g. runsc (docker only)")
	cmd.Flags().BoolVar(&options.isolationBwrapPassthrough, "isolation-bwrap-passthrough", false, "Expose host/base-image tools as a fallback (bwrap only; forfeits host-tools-unreachable)")
	cmd.Flags().StringSliceVar(&options.initCommands, "init-command", nil, "Init command to run before the agent for --provision command (repeatable)")
	cmd.Flags().StringVar(&options.nixSource, "nix-source", options.nixSource, "Nix source for --provision nix (packages, flake)")
	cmd.Flags().StringVar(&options.nixRev, "nix-rev", "", "Pinned nixpkgs rev for --nix-source packages")
	cmd.Flags().StringSliceVar(&options.nixPackages, "nix-packages", nil, "Packages for --nix-source packages (repeatable)")
	cmd.Flags().StringVar(&options.nixShell, "nix-shell", options.nixShell, "devShell name for --nix-source flake")
	cmd.Flags().StringVar(&options.image, "image", "", "Container image (for docker isolation)")

	// Workspace / terminal / misc.
	cmd.Flags().StringVarP(&options.workDir, "work-dir", "w", "", "Working directory")
	cmd.Flags().StringVar(&options.worktreeBranch, "worktree-branch", "", "Create worktree with branch name")
	cmd.Flags().StringVar(&options.worktreeSourceBranch, "worktree-source-branch", "", "Source branch for worktree")
	cmd.Flags().StringVar(&options.worktreeDir, "worktree-dir", "", "Worktree directory path")
	cmd.Flags().StringVarP(&options.config, "config", "c", "", "Configuration file")
	cmd.Flags().StringSliceVar(&options.bindPaths, "bind", nil, "Read-write bind mount")
	cmd.Flags().StringSliceVar(&options.roBindPaths, "ro-bind", nil, "Read-only bind mount")
	cmd.Flags().StringSliceVar(&options.customEnv, "env", nil, "Environment variables (KEY=VALUE)")
	cmd.Flags().StringVarP(&options.terminal, "terminal", "t", "", "Terminal wrapper (tmux)")
	cmd.Flags().StringVar(&options.terminalSession, "terminal-session", "", "Custom tmux session name")
	cmd.Flags().StringVar(&options.terminalWindow, "terminal-window", "", "Custom tmux window name")
	cmd.Flags().BoolVar(&options.terminalDetach, "terminal-detach", false, "Run in background (detach from terminal)")
	cmd.Flags().StringVar(&options.vcs, "vcs", options.vcs, "VCS type for worktree (git, jj)")
	cmd.Flags().BoolVarP(&options.dryRun, "dry-run", "n", false, "Show configuration without executing")

	// Independent policy guarantees.
	cmd.Flags().BoolVar(&options.requirePinnedProvision, "require-pinned-provision", false,
		"Require provisioning from a pinned source")
	cmd.Flags().BoolVar(&options.requireHostToolsUnreachable, "require-host-tools-unreachable", false,
		"Require host tools to be unreachable from the sandbox")
	cmd.Flags().BoolVar(&options.requireEgressRestricted, "require-egress-restricted", false,
		"Require network egress to be disabled or enforced by a supported transport")
	cmd.Flags().BoolVar(&options.requireKernelIsolation, "require-kernel-isolation", false,
		"Require a kernel-isolating runtime such as docker with runsc")

	return cmd
}

var runCmd = newRunCommand()

func init() {
	rootCmd.AddCommand(runCmd)
}

// flags is the axis-resolution view of the run flags. It is the input to
// resolveAxes, keeping the resolution testable without a live cobra command.
type flags struct {
	isolation                   string
	provision                   string
	network                     string
	image                       string
	isolationDockerRuntime      string
	isolationBwrapPassthrough   bool
	requirePinnedProvision      bool
	requireHostToolsUnreachable bool
	requireEgressRestricted     bool
	requireKernelIsolation      bool
	isolationChanged            bool
	provisionChanged            bool
}

// policy derives the demanded guarantees from the bare policy flags.
func (f flags) policy() policy.SandboxPolicy {
	return policy.SandboxPolicy{
		RequirePinnedProvision:      f.requirePinnedProvision,
		RequireHostToolsUnreachable: f.requireHostToolsUnreachable,
		RequireEgressRestricted:     f.requireEgressRestricted,
		RequireKernelIsolation:      f.requireKernelIsolation,
	}
}

// resolvedAxes carries the resolved plugin instances and RunConfig values.
type resolvedAxes struct {
	Isolation     isoshared.Isolator
	Provision     provshared.Provisioner
	IsolationName isolation.IsolationMethod
	ProvisionName provision.ProvisionMethod
	Network       netshared.Mode
	Passthrough   bool
	Runtime       string
	Image         string
}

// resolveAxes determines and resolves the confinement axes from the flags. The
// registry is the source of truth for valid isolation/provision methods (Select
// rejects unregistered ones); network is a closed enum validated here. When a
// pinned toolchain is required but no cell was chosen explicitly, the pinned combo
// (bwrap × nix) is selected.
func resolveAxes(f flags) (resolvedAxes, error) {
	isoStr, provStr, netStr := f.isolation, f.provision, f.network
	if isoStr == "" {
		isoStr = "none"
	}
	if provStr == "" {
		provStr = "none"
	}
	if netStr == "" {
		netStr = "host"
	}

	net, err := netshared.ParseMode(netStr)
	if err != nil {
		return resolvedAxes{}, err
	}

	pol := f.policy()
	isoName := isolation.IsolationMethod(isoStr)
	provName := provision.ProvisionMethod(provStr)
	if pol.RequirePinnedProvision && !f.isolationChanged && !f.provisionChanged {
		isoName, provName = isolation.IsolationBwrap, provision.ProvisionNix
	}

	reg := plugins.DefaultRegistry()
	req := sandbox.Request{
		Isolation:   isoName,
		Provision:   provName,
		Network:     net,
		Passthrough: f.isolationBwrapPassthrough,
		Runtime:     f.isolationDockerRuntime,
		Image:       f.image,
	}
	iso, prov, err := sandbox.Select(reg, req, pol)
	if err != nil {
		return resolvedAxes{}, err
	}

	return resolvedAxes{
		Isolation:     iso,
		Provision:     prov,
		IsolationName: isoName,
		ProvisionName: provName,
		Network:       net,
		Passthrough:   f.isolationBwrapPassthrough,
		Runtime:       f.isolationDockerRuntime,
		Image:         f.image,
	}, nil
}

func runAgent(cmd *cobra.Command, args []string, options runOptions) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	agent, err := agents.GetAgent(types.AgentType(options.agent))
	if err != nil {
		scriptlib.LogError(err.Error())
		scriptlib.LogInfo("Available agents: " + fmt.Sprint(agents.GetAgentTypes()))
		return err
	}

	fileConfig, err := cfgpkg.LoadConfigFile(options.config)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, err := resolveProvider(agent.Type, options.provider, fileConfig.Provider)
	if err != nil {
		return err
	}

	workDir := options.workDir
	if workDir == "" {
		workDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("resolve current working directory: %w", err)
		}
	}
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve work directory %q: %w", workDir, err)
	}

	workspace, err := prepareWorkspace(options, workDir, verbose)
	if err != nil {
		return err
	}

	bindPaths := wsshared.BuildBindPaths(append(append([]string{}, fileConfig.BindPaths...), options.bindPaths...), workspace.sourceRepoDir)

	axes, err := resolveAxes(flags{
		isolation:                   options.isolation,
		provision:                   options.provision,
		network:                     options.network,
		image:                       options.image,
		isolationDockerRuntime:      options.isolationDockerRuntime,
		isolationBwrapPassthrough:   options.isolationBwrapPassthrough,
		requirePinnedProvision:      options.requirePinnedProvision,
		requireHostToolsUnreachable: options.requireHostToolsUnreachable,
		requireEgressRestricted:     options.requireEgressRestricted,
		requireKernelIsolation:      options.requireKernelIsolation,
		isolationChanged:            cmd.Flags().Changed("isolation"),
		provisionChanged:            cmd.Flags().Changed("provision"),
	})
	if err != nil {
		return err
	}

	available, err := axes.Isolation.Available()
	if err != nil || !available {
		scriptlib.LogError(fmt.Sprintf("Isolation method %s is not available", axes.IsolationName))
		scriptlib.LogInfo(fmt.Sprintf("Available isolations: %v", sandbox.AvailableIsolations(plugins.DefaultRegistry())))
		return fmt.Errorf("isolation method %s is not available", axes.IsolationName)
	}

	// isolation=none has no namespace, so it cannot restrict egress; reject a
	// restricted network up front so every dispatch path (run, terminal, dry-run)
	// fails closed, not just isolator.Run.
	if axes.IsolationName == isolation.IsolationNone && axes.Network != netshared.ModeHost {
		return fmt.Errorf("isolation=none cannot restrict network egress (network=%q); use bwrap or docker", axes.Network)
	}

	input, err := provisionInput(axes.ProvisionName, options, workspace.sourceRepoDir, workspace.executionDir)
	if err != nil {
		return err
	}
	contribution, err := axes.Provision.Contribute(input)
	if err != nil {
		return err
	}

	runCfg := isoshared.RunConfig{
		HomeDir:         fileConfig.HomeDir,
		WorkDir:         workspace.executionDir,
		RepoDir:         workspace.sourceRepoDir,
		Network:         axes.Network,
		HostPassthrough: axes.Passthrough,
		Image:           axes.Image,
		Runtime:         axes.Runtime,
		BindPaths:       bindPaths,
		RoBindPaths:     append(append([]string{}, fileConfig.RoBindPaths...), options.roBindPaths...),
		CustomEnv:       append(append([]string{}, fileConfig.CustomEnv...), options.customEnv...),
		Agent:           agent,
		Provider:        provider,
		AgentArgs:       args,
		Verbose:         verbose,
	}

	// Construct the command before choosing a dispatch path. This shared
	// preflight makes dry-run, terminal, and direct execution reject the same
	// invalid paths and missing provider credentials.
	command, err := axes.Isolation.Command(runCfg, contribution)
	if err != nil {
		return fmt.Errorf("prepare sandbox command: %w", err)
	}

	if options.dryRun {
		return printDryRun(axes, command, agent, provider, args, workspace.displayDir)
	}

	if options.terminal != "" {
		return executeWithTerminal(axes, runCfg, contribution, workspace.executionDir, verbose, options)
	}

	exitCode, err := axes.Isolation.Run(runCfg, contribution)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

// resolveProvider resolves the model provider from the flag, falling back to the
// file config's provider.
func resolveProvider(agentType types.AgentType, optionProvider, fileProvider string) (*types.ModelProvider, error) {
	name := optionProvider
	if name == "" {
		name = fileProvider
	}
	if name == "" {
		return nil, nil
	}
	provider, err := providers.GetProvider(name, agentType)
	if err != nil {
		scriptlib.LogError(err.Error())
		scriptlib.LogInfo("Available providers: " + fmt.Sprint(providers.GetProviderNames(agentType)))
		return nil, err
	}
	return provider, nil
}

type preparedWorkspace struct {
	sourceRepoDir string
	executionDir  string
	displayDir    string
}

// prepareWorkspace validates and optionally creates the requested workspace.
func prepareWorkspace(options runOptions, workDir string, verbose bool) (preparedWorkspace, error) {
	if options.worktreeBranch == "" {
		return preparedWorkspace{
			sourceRepoDir: workDir,
			executionDir:  workDir,
			displayDir:    workDir,
		}, nil
	}

	if err := validation.ValidateBranch(options.worktreeBranch); err != nil {
		return preparedWorkspace{}, fmt.Errorf("invalid --worktree-branch: %w", err)
	}
	if options.worktreeSourceBranch != "" {
		if err := validation.ValidateBranch(options.worktreeSourceBranch); err != nil {
			return preparedWorkspace{}, fmt.Errorf("invalid --worktree-source-branch: %w", err)
		}
	}

	vcsType := wsp.VCSType(options.vcs)
	if !vcsType.IsValid() {
		return preparedWorkspace{}, fmt.Errorf("unknown --vcs %q; valid values: git, jj", options.vcs)
	}

	wtDir := options.worktreeDir
	if wtDir == "" {
		wtDir = wsshared.GetDefaultDir(options.worktreeBranch, filepath.Base(workDir))
	}
	displayDir := wtDir
	if !filepath.IsAbs(displayDir) {
		displayDir = filepath.Join(workDir, displayDir)
	}
	if options.dryRun {
		return preparedWorkspace{sourceRepoDir: workDir, executionDir: workDir, displayDir: displayDir}, nil
	}

	var vcs wsshared.VCS = git.New()
	if vcsType == wsp.VCSJujutsu {
		vcs = jj.New()
	}

	created, err := vcs.Setup(wsshared.Config{
		SourceBranch: options.worktreeSourceBranch,
		Branch:       options.worktreeBranch,
		Dir:          wtDir,
	}, workDir, verbose)
	if err != nil {
		return preparedWorkspace{}, err
	}
	return preparedWorkspace{sourceRepoDir: workDir, executionDir: created, displayDir: created}, nil
}

// provisionInput assembles the neutral provisioner Input. The nix source is built
// only for the nix provisioner; the others ignore it. An invalid --nix-source is
// returned as an error here (naming the bad value) rather than degrading to a
// zero source that fails later with a misleading diagnostic.
func provisionInput(provName provision.ProvisionMethod, options runOptions, repoDir, workDir string) (provshared.Input, error) {
	in := provshared.Input{InitCommands: options.initCommands}
	if provName == provision.ProvisionNix {
		src, err := provnix.SourceFromFlags(options.nixSource, options.nixRev, options.nixPackages, options.nixShell, "", repoDir, workDir)
		if err != nil {
			return provshared.Input{}, err
		}
		in.NixSource = src
	}
	return in, nil
}

// executeWithTerminal runs the resolved cell inside a terminal wrapper.
func executeWithTerminal(axes resolvedAxes, runCfg isoshared.RunConfig, contribution provision.Contribution, workDir string, verbose bool, options runOptions) error {
	terminalType := termshared.TerminalType(options.terminal)
	executor := terminal.GetExecutor(terminalType)
	if executor == nil {
		scriptlib.LogError(fmt.Sprintf("Unknown terminal type: %s", options.terminal))
		scriptlib.LogInfo(fmt.Sprintf("Available types: %v", terminal.GetAvailableTypes()))
		return fmt.Errorf("unknown terminal type: %s", options.terminal)
	}
	if !executor.IsAvailable() {
		scriptlib.LogError(fmt.Sprintf("Terminal type %s is not available (not installed)", options.terminal))
		return fmt.Errorf("terminal type %s is not available", options.terminal)
	}

	terminalConfig := &termshared.TerminalConfig{
		Type:        terminalType,
		SessionName: options.terminalSession,
		WindowName:  options.terminalWindow,
		Detach:      options.terminalDetach,
	}

	// Each isolator owns its terminal launch: bwrap/docker bake their env into the
	// command and return the host env; none returns the agent command with its full
	// resolved environment (provider + custom + the provisioner's PATH/env).
	fullCommand, env, err := axes.Isolation.TerminalCommand(runCfg, contribution)
	if err != nil {
		return fmt.Errorf("prepare terminal command: %w", err)
	}

	if verbose {
		scriptlib.LogInfo("Using terminal wrapper: " + options.terminal)
	}

	exitCode, err := executor.Execute(terminalConfig, fullCommand, env, workDir, verbose)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}

// printDryRun displays what would be executed without running it.
func printDryRun(axes resolvedAxes, command []string, agent *types.AgentConfig, provider *types.ModelProvider, args []string, workDir string) error {
	scriptlib.LogInfo("Dry run - would execute:")

	if axes.IsolationName == isolation.IsolationNone {
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

	if len(command) > 0 {
		fmt.Println(strings.Join(command, " "))
	}
	return nil
}
