package sandboxutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/agents"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/providers"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
	"github.com/xonovex/platform/packages/lib/core-go/pkg/scriptlib"
)

// BuildAgentCommand builds the command array for executing an agent
func BuildAgentCommand(config *types.SandboxConfig, binaryPrefix string) []string {
	// Get provider CLI args
	var providerCliArgs []string
	if config.Provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(config.Provider)
	}

	// Build agent args based on agent type
	execOpts := types.AgentExecOptions{
		Sandbox:         true,
		ProviderCliArgs: providerCliArgs,
	}

	var agentArgs []string
	switch config.Agent.Type {
	case types.AgentClaude:
		agentArgs = agents.BuildClaudeArgs(config.AgentArgs, execOpts)
	case types.AgentOpencode:
		agentArgs = agents.BuildOpencodeArgs(config.AgentArgs, execOpts)
	}

	// Build binary path
	binary := config.Agent.Binary
	if binaryPrefix != "" {
		binary = binaryPrefix + "/" + binary
	}

	cmd := make([]string, 0, 1+len(agentArgs))
	cmd = append(cmd, binary)
	cmd = append(cmd, agentArgs...)

	return cmd
}

// BuildProviderEnv builds provider environment safely
func BuildProviderEnv(config *types.SandboxConfig) (map[string]string, error) {
	if config.Provider == nil {
		return map[string]string{}, nil
	}

	providerEnv, err := providers.BuildProviderEnv(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to build provider environment: %w", err)
	}

	// Merge with agent env
	var agentEnv map[string]string
	switch config.Agent.Type {
	case types.AgentClaude:
		agentEnv = agents.BuildClaudeEnv(providerEnv)
	case types.AgentOpencode:
		agentEnv = agents.BuildOpencodeEnv(providerEnv)
	}

	return agentEnv, nil
}

// SpawnSandbox spawns a sandbox process and waits for completion
func SpawnSandbox(command string, args []string, env []string, errorPrefix string, verbose bool) (int, error) {
	if verbose {
		scriptlib.LogDebug(verbose, fmt.Sprintf("Executing: %s %s", command, strings.Join(args, " ")))
	}

	cmd := exec.Command(command, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return 0, nil
}

// ParseCustomEnv parses KEY=VALUE environment variable strings
func ParseCustomEnv(customEnv []string) map[string]string {
	env := make(map[string]string)
	for _, e := range customEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return env
}

// MergeEnvMaps merges multiple environment maps, later maps override earlier
func MergeEnvMaps(envMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, envMap := range envMaps {
		for k, v := range envMap {
			result[k] = v
		}
	}
	return result
}

// EnvMapToSlice converts a map of environment variables to a slice
func EnvMapToSlice(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, k+"="+v)
	}
	return env
}

// shellQuote quotes a string for safe use in a shell command
func shellQuote(s string) string {
	// If the string contains no special characters, return as-is
	safe := true
	for _, c := range s {
		isLower := c >= 'a' && c <= 'z'
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		isSpecial := c == '.' || c == '/' || c == ':' || c == '=' || c == '-' || c == '_'
		if !isLower && !isUpper && !isDigit && !isSpecial {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	// Wrap in single quotes and escape any single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// buildShellCommand builds a shell command string from an array of arguments
func buildShellCommand(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return strings.Join(quoted, " ")
}

// WrapWithInitCommands wraps a command with init commands that run before it.
// If no init commands are provided, returns the original command unchanged.
// Otherwise, returns a shell command that runs all init commands in sequence,
// stopping on first failure, then runs the main command.
func WrapWithInitCommands(command []string, initCommands []string) []string {
	if len(initCommands) == 0 {
		return command
	}

	// Build a shell command that chains init commands with &&, then runs the agent
	initChain := strings.Join(initCommands, " && ")
	mainCommand := buildShellCommand(command)
	fullCommand := initChain + " && exec " + mainCommand

	return []string{"sh", "-c", fullCommand}
}
