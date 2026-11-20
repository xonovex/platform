package agents

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// claudeAgent declares Claude Code, which reads the provider environment, takes
// no provider arguments, and needs its permission prompts off inside an
// isolation.
func claudeAgent() *types.AgentConfig {
	return &types.AgentConfig{
		Type:        types.AgentClaude,
		DisplayName: "Claude Code",
		Binary:      "claude",
		NixPackage:  "claude-code",
		Protocol:    types.ProtocolMessages,
		Access:      types.AccessProvider,
		ProviderEnv: true,
		SandboxArgs: []string{"--permission-mode", "bypassPermissions"},
		VersionArgs: []string{"--version"},
	}
}

// BuildClaudeArgs builds the arguments of Claude Code for the agent operator.
func BuildClaudeArgs(baseArgs []string, options types.AgentExecOptions) []string {
	return BuildArgs(claudeAgent(), baseArgs, options)
}
