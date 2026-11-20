package agents

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// opencodeAgent declares OpenCode, which takes the provider through arguments
// and reads no provider environment.
func opencodeAgent() *types.AgentConfig {
	return &types.AgentConfig{
		Type:         types.AgentOpencode,
		DisplayName:  "OpenCode",
		Binary:       "opencode",
		NixPackage:   "opencode",
		Access:       types.AccessProvider,
		ProviderArgs: true,
		VersionArgs:  []string{"--version"},
	}
}

// BuildOpencodeArgs builds the arguments of OpenCode for the agent operator.
func BuildOpencodeArgs(baseArgs []string, options types.AgentExecOptions) []string {
	return BuildArgs(opencodeAgent(), baseArgs, options)
}
