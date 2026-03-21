package builder

import agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"

// ClaudeCommandBuilder builds command/args for Claude Code
type ClaudeCommandBuilder struct{}

func (c *ClaudeCommandBuilder) Command(run *agentv1alpha1.AgentRun) ([]string, []string) {
	args := []string{"--permission-mode", "bypassPermissions"}
	if run.Spec.Prompt != "" {
		args = append(args, "--print", "--prompt", run.Spec.Prompt)
	}
	return []string{"claude"}, args
}
