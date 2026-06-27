package none

import (
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/executor"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// Executor implements direct host execution (no sandbox).
//
// Isolation: host tools leaked. Execute runs the agent on the host with the host
// PATH, so every host tool is reachable. Rejected under RequireHostToolsUnreachable.
type Executor struct{}

// NewExecutor creates a new none executor
func NewExecutor() *Executor {
	return &Executor{}
}

// IsAvailable always returns true (direct execution always available)
func (e *Executor) IsAvailable() (bool, error) {
	return true, nil
}

// Execute runs the agent directly without sandboxing
func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
	opts := executor.Options{
		Agent:     config.Agent,
		Provider:  config.Provider,
		Args:      config.AgentArgs,
		Cwd:       config.WorkDir,
		Verbose:   config.Verbose,
		Sandbox:   false,
		CustomEnv: config.CustomEnv,
	}

	return executor.Execute(opts)
}

// GetCommand returns the command that would be executed
func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
	return []string{config.Agent.Binary}
}
