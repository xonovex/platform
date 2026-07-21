package claude

import (
	"slices"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestCommandBuilderIncludesPrompt(t *testing.T) {
	builder := &CommandBuilder{}
	run := &agentv1alpha1.AgentRun{Spec: agentv1alpha1.AgentRunSpec{Prompt: "Review the patch"}}

	command, args := builder.Command(run, nil)

	if len(command) != 1 || command[0] != "claude" {
		t.Errorf("command = %v, want [claude]", command)
	}
	if !slices.Contains(args, "--print") || !slices.Contains(args, "Review the patch") {
		t.Errorf("args = %v, want print prompt", args)
	}
}

func TestCommandBuilderOmitsEmptyPrompt(t *testing.T) {
	builder := &CommandBuilder{}

	_, args := builder.Command(&agentv1alpha1.AgentRun{}, nil)

	if slices.Contains(args, "--prompt") {
		t.Errorf("args = %v, want no prompt", args)
	}
}
