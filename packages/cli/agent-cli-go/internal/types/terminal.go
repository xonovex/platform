package types

// TerminalType represents the terminal wrapper type
type TerminalType string

const (
	TerminalNone TerminalType = ""
	TerminalTmux TerminalType = "tmux"
)

// TerminalConfig holds terminal wrapper configuration
type TerminalConfig struct {
	Type           TerminalType
	SessionName    string // Auto-generated if empty
	WindowName     string // Defaults to directory basename
	Detach         bool   // Run in background
	AttachExisting bool   // Attach to existing session
}

// TerminalExecutor defines the interface for terminal wrapper implementations
type TerminalExecutor interface {
	IsAvailable() bool
	IsInside() bool
	Execute(config *TerminalConfig, command []string, env []string, workDir string, verbose bool) (int, error)
}
