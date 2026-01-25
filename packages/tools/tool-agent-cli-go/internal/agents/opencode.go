package agents

import (
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/types"
)

var opencodeAgent = &types.AgentConfig{
	Type:        types.AgentOpencode,
	DisplayName: "OpenCode",
	Binary:      "opencode",
	NixPackage:  "opencode",
}

// BuildOpencodeArgs builds arguments for OpenCode agent
func BuildOpencodeArgs(baseArgs []string, options types.AgentExecOptions) []string {
	args := make([]string, 0, len(baseArgs)+len(options.ProviderCliArgs))

	// Add provider CLI args first
	args = append(args, options.ProviderCliArgs...)

	// Add base args
	args = append(args, baseArgs...)

	return args
}

// BuildOpencodeEnv builds environment for OpenCode agent
func BuildOpencodeEnv(providerEnv map[string]string) map[string]string {
	// OpenCode doesn't use provider environment variables
	// It uses CLI args for model selection instead
	return map[string]string{}
}
