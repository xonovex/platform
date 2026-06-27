// Package shared is the harness axis core: the command-builder port plus the
// agent-type registry. Real leaves exist per agent (claude, opencode); the
// registry is package-level (lazy DI is the deferred follow-up).
package shared

import (
	"fmt"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/harness/claude"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/harness/opencode"
)

// CommandBuilder builds the command and args for an agent type.
type CommandBuilder interface {
	Command(run *agentv1alpha1.AgentRun) (command []string, args []string)
}

var harnessCommands = map[agentv1alpha1.AgentType]CommandBuilder{
	agentv1alpha1.AgentTypeClaude:   &claude.CommandBuilder{},
	agentv1alpha1.AgentTypeOpencode: &opencode.CommandBuilder{},
}

// GetHarnessCommand returns the command builder for the given agent type.
func GetHarnessCommand(agent agentv1alpha1.AgentType) (CommandBuilder, error) {
	b, ok := harnessCommands[agent]
	if !ok {
		return nil, fmt.Errorf("unsupported agent type: %s", agent)
	}
	return b, nil
}
