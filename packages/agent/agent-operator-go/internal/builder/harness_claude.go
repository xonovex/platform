package builder

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// ClaudeCommandBuilder builds command/args for Claude Code
type ClaudeCommandBuilder struct{}

func (c *ClaudeCommandBuilder) Command(run *agentv1alpha1.AgentRun) ([]string, []string) {
	agent, _ := agents.GetAgent(types.AgentClaude)
	args := agents.BuildClaudeArgs(nil, types.AgentExecOptions{Sandbox: true})
	if run.Spec.Prompt != "" {
		args = append(args, "--print", "--prompt", run.Spec.Prompt)
	}
	return []string{agent.Binary}, args
}
