package agents

import (
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
)

var claudeAgent = &types.AgentConfig{
	Type:        types.AgentClaude,
	DisplayName: "Claude Code",
	Binary:      "claude",
	NixPackage:  "claude-code",
}

// BuildClaudeArgs builds arguments for Claude agent
func BuildClaudeArgs(baseArgs []string, options types.AgentExecOptions) []string {
	args := make([]string, 0, len(baseArgs)+4)

	// Add permission bypass for sandbox mode
	if options.Sandbox {
		args = append(args, "--permission-mode", "bypassPermissions")
	}

	// Claude uses environment variables, not CLI args for provider config
	// So we ignore providerCliArgs

	// Add base args
	args = append(args, baseArgs...)

	return args
}

// BuildClaudeEnv builds environment for Claude agent
func BuildClaudeEnv(providerEnv map[string]string) map[string]string {
	env := make(map[string]string)

	// Copy provider environment
	for k, v := range providerEnv {
		env[k] = v
	}

	return env
}
