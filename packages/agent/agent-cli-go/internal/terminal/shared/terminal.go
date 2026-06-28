// Package shared is the terminal-output axis core: the TerminalExecutor port and
// its config types. It was formerly shared pkg/types/terminal.go (evicted from
// the shared module in the per-axis split, as the operator never implements it).
// The leaves (tmux) implement the port; this core names no concrete leaf.
package shared

// TerminalType represents the terminal wrapper type.
type TerminalType string

const (
	TerminalNone TerminalType = ""
	TerminalTmux TerminalType = "tmux"
)

// TerminalConfig holds terminal wrapper configuration.
type TerminalConfig struct {
	Type           TerminalType
	SessionName    string // Auto-generated if empty
	WindowName     string // Defaults to directory basename
	Detach         bool   // Run in background
	AttachExisting bool   // Attach to existing session
}

// TerminalExecutor is the terminal axis port: a terminal wrapper that runs a
// command inside a managed session.
type TerminalExecutor interface {
	IsAvailable() bool
	IsInside() bool
	Execute(config *TerminalConfig, command []string, env []string, workDir string, verbose bool) (int, error)
}
