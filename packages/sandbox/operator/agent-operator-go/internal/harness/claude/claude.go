// Package claude is the harness=claude leaf: command/args for Claude Code.
package claude

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	harnessshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/harness/shared"
)

// CommandBuilder builds command/args for Claude Code.
type CommandBuilder struct{}

var _ harnessshared.CommandBuilder = CommandBuilder{}

// Command returns the binary and args for an AgentRun.
func (CommandBuilder) Command(run *agentv1alpha1.AgentRun, _ []string) ([]string, []string, error) {
	agent, err := agents.GetAgent(types.AgentClaude)
	if err != nil {
		return nil, nil, err
	}
	args := agents.BuildClaudeArgs(nil, types.AgentExecOptions{Sandbox: true})
	if run.Spec.Prompt != "" {
		args = append(args, "--print", "--prompt", run.Spec.Prompt)
	}
	return []string{agent.Binary}, args, nil
}
