package agentcmd

import (
	"fmt"
	"path"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// BuildAgentCommand builds the command array for executing an agent. When
// binaryPrefix is non-empty, the agent binary is resolved under that directory
// (e.g. a provisioned closure's bin), which is a path inside the sandbox and so
// always uses slashes. It takes only the fields it needs (data coupling), not a
// whole config struct.
func BuildAgentCommand(agent *types.AgentConfig, provider *types.ModelProvider, agentArgs []string, binaryPrefix string) ([]string, error) {
	if _, err := agents.GetAgent(agent.Type); err != nil {
		return nil, err
	}

	var providerCliArgs []string
	if provider != nil {
		providerCliArgs = providers.GetProviderCliArgs(provider)
	}

	builtArgs := agents.BuildArgs(agent, agentArgs, types.AgentExecOptions{
		Sandbox:         true,
		ProviderCliArgs: providerCliArgs,
	})

	binary := agent.Binary
	if binaryPrefix != "" {
		binary = path.Join(binaryPrefix, binary)
	}

	cmd := make([]string, 0, 1+len(builtArgs))
	cmd = append(cmd, binary)
	cmd = append(cmd, builtArgs...)

	return cmd, nil
}

// BuildProviderEnv builds the agent environment from the configured provider,
// merging provider environment with agent-specific environment. The provider
// credential is read through getenv. It takes only the fields it needs (data
// coupling), not a whole config struct.
func BuildProviderEnv(agent *types.AgentConfig, provider *types.ModelProvider, getenv providers.Getenv) (map[string]string, error) {
	if provider == nil {
		return map[string]string{}, nil
	}

	if _, err := agents.GetAgent(agent.Type); err != nil {
		return nil, err
	}

	providerEnv, err := providers.BuildProviderEnv(provider, getenv)
	if err != nil {
		return nil, fmt.Errorf("failed to build provider environment: %w", err)
	}

	return agents.BuildEnv(agent, providerEnv), nil
}
