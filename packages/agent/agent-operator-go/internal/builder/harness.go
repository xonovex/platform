package builder

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// HarnessCommandBuilder builds the command and args for an agent type
type HarnessCommandBuilder interface {
	Command(run *agentv1alpha1.AgentRun) (command []string, args []string)
}

var harnessCommands = map[agentv1alpha1.AgentType]HarnessCommandBuilder{
	agentv1alpha1.AgentTypeClaude:   &ClaudeCommandBuilder{},
	agentv1alpha1.AgentTypeOpencode: &OpencodeCommandBuilder{},
}

// GetHarnessCommand returns the command builder for the given agent type
func GetHarnessCommand(agent agentv1alpha1.AgentType) (HarnessCommandBuilder, error) {
	b, ok := harnessCommands[agent]
	if !ok {
		return nil, fmt.Errorf("unsupported agent type: %s", agent)
	}
	return b, nil
}
