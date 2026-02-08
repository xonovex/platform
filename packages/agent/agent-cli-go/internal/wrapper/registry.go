package wrapper

import (
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/wrapper/tmux"
)

// GetExecutor returns a terminal executor for the specified type
func GetExecutor(terminalType types.TerminalType) types.TerminalExecutor {
	switch terminalType {
	case types.TerminalTmux:
		return tmux.NewExecutor()
	default:
		return nil
	}
}

// GetAvailableTypes returns all terminal types that are currently available
func GetAvailableTypes() []types.TerminalType {
	allTypes := []types.TerminalType{
		types.TerminalTmux,
	}

	available := make([]types.TerminalType, 0, len(allTypes))
	for _, t := range allTypes {
		executor := GetExecutor(t)
		if executor != nil && executor.IsAvailable() {
			available = append(available, t)
		}
	}

	return available
}
