// Package shared defines the terminal executor interface and its configuration.
package shared

// TerminalType represents the terminal wrapper type. Each leaf declares its own
// name; the shared package knows only the absence of a wrapper and the
// selection from the environment.
type TerminalType string

const (
	TerminalNone TerminalType = ""
	// TerminalAuto selects the one wrapper whose session the CLI runs in.
	TerminalAuto TerminalType = "auto"
)

// TerminalConfig holds terminal wrapper configuration.
type TerminalConfig struct {
	Type           TerminalType
	SessionName    string // Auto-generated if empty
	WindowName     string // Defaults to directory basename
	Detach         bool   // Run in background
	AttachExisting bool   // Attach to existing session
	// RunID is the id of the run record the run expects to get, empty when the
	// CLI keeps no run records. A wrapper that can label its session uses it.
	RunID string
}

// TerminalExecutor is the terminal axis port: a terminal wrapper that runs a
// command inside a managed session.
type TerminalExecutor interface {
	IsAvailable() bool
	IsInside() bool
	Execute(config *TerminalConfig, command []string, env []string, workDir string) (int, error)
}
