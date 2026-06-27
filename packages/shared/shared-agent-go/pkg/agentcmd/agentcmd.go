package agentcmd

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// BuildAgentCommand builds the command array for executing an agent. When
// binaryPrefix is non-empty, the agent binary is resolved under that directory
// (e.g. a provisioned closure's bin).
func BuildAgentCommand(config *types.SandboxConfig, binaryPrefix string) []string {
	var providerCliArgs []string
	if config.Provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(config.Provider)
	}

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

	binary := config.Agent.Binary
	if binaryPrefix != "" {
		binary = binaryPrefix + "/" + binary
	}

	cmd := make([]string, 0, 1+len(agentArgs))
	cmd = append(cmd, binary)
	cmd = append(cmd, agentArgs...)

	return cmd
}

// BuildProviderEnv builds the agent environment from the configured provider,
// merging provider environment with agent-specific environment.
func BuildProviderEnv(config *types.SandboxConfig) (map[string]string, error) {
	if config.Provider == nil {
		return map[string]string{}, nil
	}

	providerEnv, err := providers.BuildProviderEnv(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to build provider environment: %w", err)
	}

	var agentEnv map[string]string
	switch config.Agent.Type {
	case types.AgentClaude:
		agentEnv = agents.BuildClaudeEnv(providerEnv)
	case types.AgentOpencode:
		agentEnv = agents.BuildOpencodeEnv(providerEnv)
	}

	return agentEnv, nil
}
