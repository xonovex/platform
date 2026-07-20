package opencode

import (
	"slices"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestCommandBuilderIncludesProviderArguments(t *testing.T) {
	builder := &CommandBuilder{}
	run := &agentv1alpha1.AgentRun{Spec: agentv1alpha1.AgentRunSpec{
		Provider: &agentv1alpha1.ProviderSpec{CliArgs: []string{"--model", "custom/model"}},
	}}

	command, args := builder.Command(run)

	if len(command) != 1 || command[0] != "opencode" {
		t.Errorf("command = %v, want [opencode]", command)
	}
	if !slices.Contains(args, "custom/model") {
		t.Errorf("args = %v, want provider model", args)
	}
}

func TestCommandBuilderSupportsRunWithoutProvider(t *testing.T) {
	builder := &CommandBuilder{}

	command, args := builder.Command(&agentv1alpha1.AgentRun{})

	if len(command) != 1 || command[0] != "opencode" || len(args) != 0 {
		t.Errorf("Command() = %v %v, want executable without arguments", command, args)
	}
}
