// Package terminal is the terminal-output axis composition root: it selects a
// TerminalExecutor leaf for a requested type. It is the only terminal package
// that imports a concrete leaf (tmux); the port lives in terminal/shared.
package terminal

import (
	termshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/terminal/shared"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/terminal/tmux"
)

// GetExecutor returns a terminal executor for the specified type, or nil when no
// wrapper applies (TerminalNone or an unknown type).
func GetExecutor(terminalType termshared.TerminalType) termshared.TerminalExecutor {
	switch terminalType {
	case termshared.TerminalTmux:
		return tmux.NewExecutor()
	default:
		return nil
	}
}

// GetAvailableTypes returns all terminal types that are currently available.
func GetAvailableTypes() []termshared.TerminalType {
	allTypes := []termshared.TerminalType{
		termshared.TerminalTmux,
	}

	available := make([]termshared.TerminalType, 0, len(allTypes))
	for _, t := range allTypes {
		executor := GetExecutor(t)
		if executor != nil && executor.IsAvailable() {
			available = append(available, t)
		}
	}

	return available
}
