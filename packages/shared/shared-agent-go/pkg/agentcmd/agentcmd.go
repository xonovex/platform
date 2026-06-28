package agentcmd

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// BuildAgentCommand builds the command array for executing an agent. When
// binaryPrefix is non-empty, the agent binary is resolved under that directory
// (e.g. a provisioned closure's bin). It takes only the fields it needs (data
// coupling), not a whole config struct.
func BuildAgentCommand(agent *types.AgentConfig, provider *types.ModelProvider, agentArgs []string, binaryPrefix string) []string {
	var providerCliArgs []string
	if provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(provider)
	}

	execOpts := types.AgentExecOptions{
		Sandbox:         true,
		ProviderCliArgs: providerCliArgs,
	}

	var builtArgs []string
	switch agent.Type {
	case types.AgentClaude:
		builtArgs = agents.BuildClaudeArgs(agentArgs, execOpts)
	case types.AgentOpencode:
		builtArgs = agents.BuildOpencodeArgs(agentArgs, execOpts)
	}

	binary := agent.Binary
	if binaryPrefix != "" {
		binary = binaryPrefix + "/" + binary
	}

	cmd := make([]string, 0, 1+len(builtArgs))
	cmd = append(cmd, binary)
	cmd = append(cmd, builtArgs...)

	return cmd
}

// BuildProviderEnv builds the agent environment from the configured provider,
// merging provider environment with agent-specific environment. It takes only
// the fields it needs (data coupling), not a whole config struct.
func BuildProviderEnv(agent *types.AgentConfig, provider *types.ModelProvider) (map[string]string, error) {
	if provider == nil {
		return map[string]string{}, nil
	}

	providerEnv, err := providers.BuildProviderEnv(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to build provider environment: %w", err)
	}

	var agentEnv map[string]string
	switch agent.Type {
	case types.AgentClaude:
		agentEnv = agents.BuildClaudeEnv(providerEnv)
	case types.AgentOpencode:
		agentEnv = agents.BuildOpencodeEnv(providerEnv)
	}

	return agentEnv, nil
}
