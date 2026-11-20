package agents

import (
	"fmt"
	"maps"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// leaves returns the constructor of each agent type's declared facts. A new
// agent type adds one file that declares its facts and one entry here; no other
// file names a concrete agent.
func leaves() map[types.AgentType]func() *types.AgentConfig {
	return map[types.AgentType]func() *types.AgentConfig{
		types.AgentClaude:   claudeAgent,
		types.AgentCodex:    codexAgent,
		types.AgentCopilot:  copilotAgent,
		types.AgentKiro:     kiroAgent,
		types.AgentOpencode: opencodeAgent,
		types.AgentPi:       piAgent,
	}
}

// GetAgent retrieves a fresh agent configuration by type.
func GetAgent(agentType types.AgentType) (*types.AgentConfig, error) {
	build, ok := leaves()[agentType]
	if !ok {
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}
	return build(), nil
}

// GetAgentTypes returns all available agent types in sorted order.
func GetAgentTypes() []types.AgentType {
	return slices.Sorted(maps.Keys(leaves()))
}
