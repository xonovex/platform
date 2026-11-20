package agents

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// Agent is the behavior of one agent type: its configuration and how it turns
// execution options and provider environment into arguments and environment.
type Agent struct {
	Config    func() *types.AgentConfig
	BuildArgs func(baseArgs []string, options types.AgentExecOptions) []string
	BuildEnv  func(providerEnv map[string]string) map[string]string
}

// Lookup returns the behavior of an agent type. A new agent type adds one file
// that defines its Agent and one case here.
func Lookup(agentType types.AgentType) (Agent, error) {
	switch agentType {
	case types.AgentClaude:
		return claude(), nil
	case types.AgentOpencode:
		return opencode(), nil
	default:
		return Agent{}, fmt.Errorf("unknown agent type: %s", agentType)
	}
}

// GetAgent retrieves a fresh agent configuration by type.
func GetAgent(agentType types.AgentType) (*types.AgentConfig, error) {
	agent, err := Lookup(agentType)
	if err != nil {
		return nil, err
	}
	return agent.Config(), nil
}

// GetAgentTypes returns all available agent types.
func GetAgentTypes() []types.AgentType {
	return []types.AgentType{types.AgentClaude, types.AgentOpencode}
}
