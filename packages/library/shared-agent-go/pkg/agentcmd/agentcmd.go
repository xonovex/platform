package agentcmd

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

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
