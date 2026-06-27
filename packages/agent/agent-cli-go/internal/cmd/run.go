package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cfgpkg "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/config"
	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/proxy"
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
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
	wsp "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

var runCmd = &cobra.Command{
	Use:   "run [agent-args...]",
	Short: "Run an AI coding agent",
	Long:  `Run an AI coding agent with specified provider and sandbox options.`,
	RunE:  runAgent,
}

var (
	flagAgent                     string
	flagProvider                  string
	flagIsolation                 string
	flagProvision                 string
	flagNetwork                   string
	flagNetworkProxyEgressAllow   []string
	flagIsolationBwrapPassthrough bool
	flagIsolationDockerRuntime    string
	flagInitCommand               []string
	flagNixSource                 string
	flagNixRev                    string
	flagNixPackages               []string
	flagNixShell                  string
	flagWorkDir                   string
	flagWorktreeBranch            string
	flagWorktreeSourceBranch      string
	flagWorktreeDir               string
	flagConfig                    string
	flagBind                      []string
	flagRoBind                    []string
	flagEnv                       []string
	flagImage                     string
	flagTerminal                  string
	flagTerminalSession           string
	flagTerminalWindow            string
	flagTerminalDetach            bool
	flagVCS                       string
	flagDryRun                    bool
	flagRequirePinned             bool
)

func init() {
	rootCmd.AddCommand(runCmd)

	// Bare axis selectors.
	runCmd.Flags().StringVarP(&flagAgent, "agent", "a", "claude", "Agent to run (claude, opencode)")
	runCmd.Flags().StringVarP(&flagProvider, "provider", "p", "", "Model provider")
	runCmd.Flags().StringVar(&flagIsolation, "isolation", "none", "Isolation axis (none, bwrap, docker)")
	runCmd.Flags().StringVar(&flagProvision, "provision", "none", "Provision axis (none, nix, command)")
	runCmd.Flags().StringVar(&flagNetwork, "network", "host", "Network egress axis (host, none, proxy)")

	// Per-type knobs under the --<axis>-<type>-<option> grammar.
	runCmd.Flags().StringVar(&flagIsolationDockerRuntime, "isolation-docker-runtime", "", "Kernel-isolating container runtime, e.g. runsc (docker only)")
	runCmd.Flags().BoolVar(&flagIsolationBwrapPassthrough, "isolation-bwrap-passthrough", false, "Expose host/base-image tools as a fallback (bwrap only; forfeits host-tools-unreachable)")
	runCmd.Flags().StringSliceVar(&flagNetworkProxyEgressAllow, "network-proxy-egress-allow", []string{}, "Extra egress allowlist host for --network proxy (proxy only, repeatable)")
	runCmd.Flags().StringSliceVar(&flagInitCommand, "init-command", []string{}, "Init command to run before the agent for --provision command (repeatable)")
	runCmd.Flags().StringVar(&flagNixSource, "nix-source", "packages", "Nix source for --provision nix (packages, flake)")
	runCmd.Flags().StringVar(&flagNixRev, "nix-rev", "", "Pinned nixpkgs rev for --nix-source packages")
	runCmd.Flags().StringSliceVar(&flagNixPackages, "nix-packages", []string{}, "Packages for --nix-source packages (repeatable)")
	runCmd.Flags().StringVar(&flagNixShell, "nix-shell", "default", "devShell name for --nix-source flake")
	runCmd.Flags().StringVar(&flagImage, "image", "", "Container image (for docker isolation)")

	// Workspace / terminal / misc.
	runCmd.Flags().StringVarP(&flagWorkDir, "work-dir", "w", "", "Working directory")
	runCmd.Flags().StringVar(&flagWorktreeBranch, "worktree-branch", "", "Create worktree with branch name")
	runCmd.Flags().StringVar(&flagWorktreeSourceBranch, "worktree-source-branch", "", "Source branch for worktree")
	runCmd.Flags().StringVar(&flagWorktreeDir, "worktree-dir", "", "Worktree directory path")
	runCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "Configuration file")
	runCmd.Flags().StringSliceVar(&flagBind, "bind", []string{}, "Read-write bind mount")
	runCmd.Flags().StringSliceVar(&flagRoBind, "ro-bind", []string{}, "Read-only bind mount")
	runCmd.Flags().StringSliceVar(&flagEnv, "env", []string{}, "Environment variables (KEY=VALUE)")
	runCmd.Flags().StringVarP(&flagTerminal, "terminal", "t", "", "Terminal wrapper (tmux)")
	runCmd.Flags().StringVar(&flagTerminalSession, "terminal-session", "", "Custom tmux session name")
	runCmd.Flags().StringVar(&flagTerminalWindow, "terminal-window", "", "Custom tmux window name")
	runCmd.Flags().BoolVar(&flagTerminalDetach, "terminal-detach", false, "Run in background (detach from terminal)")
	runCmd.Flags().StringVar(&flagVCS, "vcs", "git", "VCS type for worktree (git, jj)")
	runCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "n", false, "Show configuration without executing")

	// Bare policy flag.
	runCmd.Flags().BoolVar(&flagRequirePinned, "require-pinned-toolchain", false,
		"Mandate pinned provisioning + host-tools-unreachable; reject leaky/host-exposed cells")
}

// flags is the axis-resolution view of the run flags. It is the input to
// resolveAxes, keeping the resolution testable without a live cobra command.
type flags struct {
	isolation                 string
	provision                 string
	network                   string
	image                     string
	isolationDockerRuntime    string
	isolationBwrapPassthrough bool
	requirePinned             bool
	isolationChanged          bool
	provisionChanged          bool
}

// policy derives the demanded guarantees from the bare policy flags.
func (f flags) policy() policy.SandboxPolicy {
	return policy.SandboxPolicy{
		RequirePinnedProvision:      f.requirePinned,
		RequireHostToolsUnreachable: f.requirePinned,
	}
}

// resolvedAxes is the named result of axis resolution (weakening the positional
// connascence of the former multi-value tuple). It carries the resolved plugin
// instances plus the knobs the RunConfig needs.
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

func runAgent(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	agent, err := agents.GetAgent(types.AgentType(flagAgent))
	if err != nil {
		scriptlib.LogError(err.Error())
		scriptlib.LogInfo("Available agents: " + fmt.Sprint(agents.GetAgentTypes()))
		return err
	}

	fileConfig, err := cfgpkg.LoadConfigFile(flagConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, err := resolveProvider(agent.Type, fileConfig.Provider)
	if err != nil {
		return err
	}

	workDir := flagWorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	workDir, _ = filepath.Abs(workDir)

	sourceRepoDir, workDir, err := setupWorktree(workDir, verbose)
	if err != nil {
		return err
	}

	bindPaths := wsshared.BuildBindPaths(flagBind, sourceRepoDir)

	axes, err := resolveAxes(flags{
		isolation:                 flagIsolation,
		provision:                 flagProvision,
		network:                   flagNetwork,
		image:                     flagImage,
		isolationDockerRuntime:    flagIsolationDockerRuntime,
		isolationBwrapPassthrough: flagIsolationBwrapPassthrough,
		requirePinned:             flagRequirePinned,
		isolationChanged:          cmd.Flags().Changed("isolation"),
		provisionChanged:          cmd.Flags().Changed("provision"),
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

	contribution, err := axes.Provision.Contribute(provisionInput(axes.ProvisionName, sourceRepoDir, workDir))
	if err != nil {
		return err
	}

	runCfg := isoshared.RunConfig{
		WorkDir:         workDir,
		RepoDir:         sourceRepoDir,
		Network:         axes.Network,
		ProxyEnv:        proxyEnv(),
		HostPassthrough: axes.Passthrough,
		Image:           axes.Image,
		Runtime:         axes.Runtime,
		BindPaths:       bindPaths,
		RoBindPaths:     flagRoBind,
		CustomEnv:       flagEnv,
		Agent:           agent,
		Provider:        provider,
		AgentArgs:       args,
		Verbose:         verbose,
	}

	if flagDryRun {
		return printDryRun(axes, runCfg, contribution, agent, provider, args, workDir)
	}

	if flagTerminal != "" {
		return executeWithTerminal(axes, runCfg, contribution, agent, provider, workDir, verbose)
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
func resolveProvider(agentType types.AgentType, fileProvider string) (*types.ModelProvider, error) {
	name := flagProvider
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

// setupWorktree creates a worktree/workspace when --worktree-branch is set,
// returning the source repo dir and the (possibly redirected) work dir.
func setupWorktree(workDir string, verbose bool) (sourceRepoDir, newWorkDir string, err error) {
	if flagWorktreeBranch == "" {
		return "", workDir, nil
	}

	if err := validation.ValidateBranch(flagWorktreeBranch); err != nil {
		return "", "", fmt.Errorf("invalid --worktree-branch: %w", err)
	}
	if flagWorktreeSourceBranch != "" {
		if err := validation.ValidateBranch(flagWorktreeSourceBranch); err != nil {
			return "", "", fmt.Errorf("invalid --worktree-source-branch: %w", err)
		}
	}

	vcsType := wsp.VCSType(flagVCS)
	if !vcsType.IsValid() {
		return "", "", fmt.Errorf("unknown --vcs %q; valid values: git, jj", flagVCS)
	}

	wtDir := flagWorktreeDir
	if wtDir == "" {
		wtDir = wsshared.GetDefaultDir(flagWorktreeBranch, filepath.Base(workDir))
	}

	var vcs wsshared.VCS = git.New()
	if vcsType == wsp.VCSJujutsu {
		vcs = jj.New()
	}

	created, err := vcs.Setup(wsshared.Config{
		SourceBranch: flagWorktreeSourceBranch,
		Branch:       flagWorktreeBranch,
		Dir:          wtDir,
	}, workDir, verbose)
	if err != nil {
		return "", "", err
	}
	return workDir, created, nil
}

// provisionInput assembles the neutral provisioner Input. The nix source is built
// only for the nix provisioner; the others ignore it.
func provisionInput(provName provision.ProvisionMethod, repoDir, workDir string) provshared.Input {
	in := provshared.Input{InitCommands: flagInitCommand}
	if provName == provision.ProvisionNix {
		// Errors here surface at Contribute via ValidateSource (fail closed).
		src, _ := provnix.SourceFromFlags(flagNixSource, flagNixRev, flagNixPackages, flagNixShell, "", repoDir, workDir)
		in.NixSource = src
	}
	return in
}

// proxyEnv builds the resolved proxy egress environment (allowlist folded in), or
// nil when the network mode is not proxy or no proxy URL is configured.
func proxyEnv() map[string]string {
	return proxy.Options{
		EgressAllowlist: append(append([]string{}, network.DefaultEgressAllowlist...), flagNetworkProxyEgressAllow...),
		URL:             netshared.ProxyURL(),
	}.Env()
}

// executeWithTerminal runs the resolved cell inside a terminal wrapper.
func executeWithTerminal(axes resolvedAxes, runCfg isoshared.RunConfig, contribution provision.Contribution, agent *types.AgentConfig, provider *types.ModelProvider, workDir string, verbose bool) error {
	terminalType := termshared.TerminalType(flagTerminal)
	executor := terminal.GetExecutor(terminalType)
	if executor == nil {
		scriptlib.LogError(fmt.Sprintf("Unknown terminal type: %s", flagTerminal))
		scriptlib.LogInfo(fmt.Sprintf("Available types: %v", terminal.GetAvailableTypes()))
		return fmt.Errorf("unknown terminal type: %s", flagTerminal)
	}
	if !executor.IsAvailable() {
		scriptlib.LogError(fmt.Sprintf("Terminal type %s is not available (not installed)", flagTerminal))
		return fmt.Errorf("terminal type %s is not available", flagTerminal)
	}

	terminalConfig := &termshared.TerminalConfig{
		Type:        terminalType,
		SessionName: flagTerminalSession,
		WindowName:  flagTerminalWindow,
		Detach:      flagTerminalDetach,
	}

	// The isolator's Command is self-contained; bwrap/docker bake their env into
	// the command, so the wrapper only needs the host env. Host (no-isolation)
	// execution has no wrapper baking, so its provider env is added here.
	fullCommand := axes.Isolation.Command(runCfg, contribution)
	env := append(os.Environ(), runCfg.CustomEnv...)
	if axes.IsolationName == isolation.IsolationNone {
		if providerEnv, err := agentcmd.BuildProviderEnv(agent, provider); err == nil {
			for k, v := range providerEnv {
				env = append(env, k+"="+v)
			}
		}
	}

	if verbose {
		scriptlib.LogInfo("Using terminal wrapper: " + flagTerminal)
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
func printDryRun(axes resolvedAxes, runCfg isoshared.RunConfig, contribution provision.Contribution, agent *types.AgentConfig, provider *types.ModelProvider, args []string, workDir string) error {
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

	command := axes.Isolation.Command(runCfg, contribution)
	if len(command) > 0 {
		fmt.Println(strings.Join(command, " "))
	}
	return nil
}
