package agents

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

var agentRegistry = map[types.AgentType]*types.AgentConfig{
	types.AgentClaude:   claudeAgent,
	types.AgentOpencode: opencodeAgent,
}

// GetAgent retrieves an agent by type
func GetAgent(agentType types.AgentType) (*types.AgentConfig, error) {
	agent, ok := agentRegistry[agentType]
	if !ok {
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
	return agent, nil
}

// GetAgentTypes returns all available agent types
func GetAgentTypes() []types.AgentType {
	agentTypes := make([]types.AgentType, 0, len(agentRegistry))
	for t := range agentRegistry {
		agentTypes = append(agentTypes, t)
	}
	return agentTypes
}
