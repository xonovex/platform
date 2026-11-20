// Package overlay selects the overlay dialect of an agent. This file is the only
// one that imports the leaves, and the only one that names an agent type: a new
// harness adds one leaf package and one entry here.
package overlay

import (
	"fmt"
	"maps"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/claude"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/codex"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/copilot"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/kiro"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/opencode"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/pi"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// leaves returns the constructor of the dialect of each agent type.
func leaves() map[types.AgentType]func() shared.Dialect {
	return map[types.AgentType]func() shared.Dialect{
		types.AgentClaude:   claude.New,
		types.AgentCodex:    codex.New,
		types.AgentCopilot:  copilot.New,
		types.AgentKiro:     kiro.New,
		types.AgentOpencode: opencode.New,
		types.AgentPi:       pi.New,
	}
}

// Resolve returns the dialect of an agent type.
func Resolve(agentType types.AgentType) (shared.Dialect, error) {
	build, ok := leaves()[agentType]
	if !ok {
		return nil, fmt.Errorf("agent %s has no overlay dialect", agentType)
	}
	return build(), nil
}

// AgentTypes returns the agent types that have a dialect, in sorted order.
func AgentTypes() []types.AgentType {
	return slices.Sorted(maps.Keys(leaves()))
}

// Check fails closed when the overlay of agent would cut it off from its model
// access: a dialect that replaces the home directory of the agent also replaces
// the login stored there, so an agent that uses its own login cannot take it.
func Check(agent *types.AgentConfig) error {
	dialect, err := Resolve(agent.Type)
	if err != nil {
		return err
	}
	if dialect.Facts().ReplacesHome && agent.Access == types.AccessLogin {
		return fmt.Errorf("agent %s keeps its login in the directory the overlay replaces; run it without plugins and overlay directories", agent.Type)
	}
	return nil
}
