package agents

import (
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// GetAgent retrieves a fresh agent configuration by type.
func GetAgent(agentType types.AgentType) (*types.AgentConfig, error) {
	switch agentType {
	case types.AgentClaude:
		return claudeAgent(), nil
	case types.AgentOpencode:
		return opencodeAgent(), nil
	default:
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
}

// GetAgentTypes returns all available agent types.
func GetAgentTypes() []types.AgentType {
	return []types.AgentType{types.AgentClaude, types.AgentOpencode}
}
