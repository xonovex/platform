// Package opencode is the harness=opencode leaf: command/args for Opencode.
package opencode

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// CommandBuilder builds command/args for Opencode.
type CommandBuilder struct{}

// Command returns the binary and args for an AgentRun.
func (o *CommandBuilder) Command(_ *agentv1alpha1.AgentRun, providerCliArgs []string) ([]string, []string) {
	agent, _ := agents.GetAgent(types.AgentOpencode)
	args := agents.BuildOpencodeArgs(nil, types.AgentExecOptions{
		Sandbox:         true,
		ProviderCliArgs: providerCliArgs,
	})
	return []string{agent.Binary}, args
}
