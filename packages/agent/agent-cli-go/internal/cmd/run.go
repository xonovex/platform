package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/executor"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/worktree"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/wrapper"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/config"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
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
	flagSandbox              string
	flagWorkDir              string
	flagWorktreeBranch       string
	flagWorktreeSourceBranch string
	flagWorktreeDir          string
	flagConfig               string
	flagBind                 []string
	flagRoBind               []string
	flagEnv                  []string
	flagNetwork              bool
	flagImage                string
	flagComposeFile          string
	flagService              string
	flagTerminal             string
	flagTerminalSession      string
	flagTerminalWindow       string
	flagTerminalDetach       bool
	flagDryRun               bool
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&flagAgent, "agent", "a", "claude", "Agent to run (claude, opencode)")
	runCmd.Flags().StringVarP(&flagProvider, "provider", "p", "", "Model provider")
	runCmd.Flags().StringVarP(&flagSandbox, "sandbox", "s", "none", "Sandbox method (none, bwrap, docker, compose, nix)")
	runCmd.Flags().StringVarP(&flagWorkDir, "work-dir", "w", "", "Working directory")
	runCmd.Flags().StringVar(&flagWorktreeBranch, "worktree-branch", "", "Create worktree with branch name")
	runCmd.Flags().StringVar(&flagWorktreeSourceBranch, "worktree-source-branch", "", "Source branch for worktree")
	runCmd.Flags().StringVar(&flagWorktreeDir, "worktree-dir", "", "Worktree directory path")
	runCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "Configuration file")
	runCmd.Flags().StringSliceVar(&flagBind, "bind", []string{}, "Read-write bind mount")
	runCmd.Flags().StringSliceVar(&flagRoBind, "ro-bind", []string{}, "Read-only bind mount")
	runCmd.Flags().StringSliceVar(&flagEnv, "env", []string{}, "Environment variables (KEY=VALUE)")
	runCmd.Flags().BoolVarP(&flagNetwork, "network", "N", true, "Enable network access")
	runCmd.Flags().StringVar(&flagImage, "image", "", "Container image (for docker/nix sandboxes)")
	runCmd.Flags().StringVar(&flagComposeFile, "compose-file", "", "Docker Compose file (for compose sandbox)")
	runCmd.Flags().StringVar(&flagService, "service", "", "Docker Compose service (for compose sandbox)")
	runCmd.Flags().StringVarP(&flagTerminal, "terminal", "t", "", "Terminal wrapper (tmux)")
	runCmd.Flags().StringVar(&flagTerminalSession, "terminal-session", "", "Custom tmux session name")
	runCmd.Flags().StringVar(&flagTerminalWindow, "terminal-window", "", "Custom tmux window name")
	runCmd.Flags().BoolVar(&flagTerminalDetach, "terminal-detach", false, "Run in background (detach from terminal)")
	runCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "n", false, "Show configuration without executing")
}

func runAgent(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get agent
	agent, err := agents.GetAgent(types.AgentType(flagAgent))
	if err != nil {
		scriptlib.LogError(err.Error())
		scriptlib.LogInfo("Available agents: " + fmt.Sprint(agents.GetAgentTypes()))
		return err
	}

	// Load config file
	fileConfig, err := config.LoadConfigFile(flagConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get provider if specified
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

	// Resolve work directory
	workDir := flagWorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	workDir, _ = filepath.Abs(workDir)

	var sourceRepoDir string

	// Setup worktree if requested
	if flagWorktreeBranch != "" {
		repoName := filepath.Base(workDir)
		wtDir := flagWorktreeDir
		if wtDir == "" {
			wtDir = sharedworktree.GetDefaultDir(flagWorktreeBranch, repoName)
		}

		wtConfig := worktree.Config{
			SourceBranch: flagWorktreeSourceBranch,
			Branch:       flagWorktreeBranch,
			Dir:          wtDir,
		}

		newWorkDir, err := worktree.Setup(wtConfig, workDir, verbose)
		if err != nil {
			return err
		}

		sourceRepoDir = workDir
		workDir = newWorkDir
	}

	// Build bind paths with worktree
	bindPaths := worktree.BuildBindPaths(flagBind, sourceRepoDir)

	// Get sandbox executor
	sandboxMethod := types.SandboxMethod(flagSandbox)
	sandboxExecutor, err := sandbox.GetExecutor(sandboxMethod)
	if err != nil {
		return err
	}

	// Check availability
	available, err := sandboxExecutor.IsAvailable()
	if err != nil || !available {
		availableMethods := sandbox.GetAvailableMethods()
		scriptlib.LogError(fmt.Sprintf("Sandbox method %s is not available", flagSandbox))
		scriptlib.LogInfo(fmt.Sprintf("Available methods: %v", availableMethods))
		return fmt.Errorf("sandbox method %s is not available", flagSandbox)
	}

	// Build sandbox config (used by both sandbox and terminal wrapper)
	sandboxConfig := &types.SandboxConfig{
		Method:      sandboxMethod,
		Agent:       agent,
		Provider:    provider,
		WorkDir:     workDir,
		RepoDir:     sourceRepoDir,
		Network:     flagNetwork,
		BindPaths:   bindPaths,
		RoBindPaths: flagRoBind,
		CustomEnv:   flagEnv,
		AgentArgs:   args,
		Verbose:     verbose,
		Image:       flagImage,
		ComposeFile: flagComposeFile,
		Service:     flagService,
	}

	// Handle dry-run mode
	if flagDryRun {
		return printDryRun(sandboxMethod, sandboxExecutor, sandboxConfig, agent, provider, args, workDir, verbose)
	}

	// Check if terminal wrapper is requested
	if flagTerminal != "" {
		return executeWithTerminal(sandboxMethod, sandboxExecutor, sandboxConfig, agent, provider, args, workDir, verbose)
	}

	// For "none" sandbox, use direct execution for simplicity
	if sandboxMethod == types.SandboxNone {
		execOpts := executor.Options{
			Agent:     agent,
			Provider:  provider,
			Args:      args,
			Cwd:       workDir,
			Verbose:   verbose,
			Sandbox:   false,
			CustomEnv: flagEnv,
		}

		exitCode, err := executor.Execute(execOpts)
		if err != nil {
			return err
		}

		if exitCode != 0 {
			os.Exit(exitCode)
		}

		return nil
	}

	// Execute via sandbox
	exitCode, err := sandboxExecutor.Execute(sandboxConfig)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// executeWithTerminal handles execution when a terminal wrapper is requested
func executeWithTerminal(sandboxMethod types.SandboxMethod, sandboxExecutor types.SandboxExecutor, sandboxConfig *types.SandboxConfig, agent *types.AgentConfig, provider *types.ModelProvider, args []string, workDir string, verbose bool) error {
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

	// Build terminal config
	terminalConfig := &types.TerminalConfig{
		Type:        terminalType,
		SessionName: flagTerminalSession,
		WindowName:  flagTerminalWindow,
		Detach:      flagTerminalDetach,
	}

	// Build the full command and environment
	var fullCommand []string
	var env []string

	if sandboxMethod == types.SandboxNone {
		// Build direct agent command
		fullCommand = buildDirectCommand(agent, provider, args, sandboxConfig.CustomEnv)
		env = buildDirectEnv(agent, provider, sandboxConfig.CustomEnv)
	} else {
		// Use sandbox's GetCommand to get the full wrapped command
		fullCommand = sandboxExecutor.GetCommand(sandboxConfig)
		// For sandbox execution, environment is handled by the sandbox
		env = os.Environ()
		env = append(env, sandboxConfig.CustomEnv...)
	}

	if verbose {
		scriptlib.LogInfo("Using terminal wrapper: " + flagTerminal)
	}

	// Execute via terminal wrapper
	exitCode, err := terminalExecutor.Execute(terminalConfig, fullCommand, env, workDir, verbose)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// buildDirectCommand builds the command for direct (non-sandbox) execution
func buildDirectCommand(agent *types.AgentConfig, provider *types.ModelProvider, args []string, customEnv []string) []string {
	var providerCliArgs []string
	if provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(provider)
	}

	execOpts := types.AgentExecOptions{
		Sandbox:         false,
		ProviderCliArgs: providerCliArgs,
	}

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

// buildDirectEnv builds the environment for direct (non-sandbox) execution
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

// printDryRun displays what would be executed without actually running it
func printDryRun(sandboxMethod types.SandboxMethod, sandboxExecutor types.SandboxExecutor, sandboxConfig *types.SandboxConfig, agent *types.AgentConfig, provider *types.ModelProvider, args []string, workDir string, verbose bool) error {
	scriptlib.LogInfo("Dry run - would execute:")

	if sandboxMethod == types.SandboxNone {
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
		return nil
	}

	// Get the full command from the sandbox executor
	command := sandboxExecutor.GetCommand(sandboxConfig)
	if len(command) > 0 {
		fmt.Println(strings.Join(command, " "))
	}

	return nil
}
