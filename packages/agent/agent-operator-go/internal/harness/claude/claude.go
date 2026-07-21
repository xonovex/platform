// Package claude is the harness=claude leaf: command/args for Claude Code.
package claude

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// CommandBuilder builds command/args for Claude Code.
type CommandBuilder struct{}

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
