package agentcmd

import (
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestBuildAgentCommandBinaryPrefix(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"}

	cmd := BuildAgentCommand(agent, nil, []string{"--foo"}, "/env/bin")
	if len(cmd) == 0 {
		t.Fatal("BuildAgentCommand returned empty command")
	}
	if cmd[0] != "/env/bin/claude" {
		t.Errorf("binary = %q, want /env/bin/claude", cmd[0])
	}
}

func TestBuildAgentCommandNoPrefix(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentOpencode, Binary: "opencode"}

	cmd := BuildAgentCommand(agent, nil, nil, "")
	if len(cmd) == 0 || cmd[0] != "opencode" {
		t.Errorf("binary = %v, want opencode first", cmd)
	}
}

func TestBuildProviderEnvNilProvider(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"}

	env, err := BuildProviderEnv(agent, nil)
	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v, want nil", err)
	}
	if len(env) != 0 {
		t.Errorf("env = %v, want empty", env)
	}
}
