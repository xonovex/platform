package agentcmd

import (
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestBuildAgentCommandBinaryPrefix(t *testing.T) {
	config := &types.SandboxConfig{
		Agent:     &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
		AgentArgs: []string{"--foo"},
	}

	cmd := BuildAgentCommand(config, "/env/bin")
	if len(cmd) == 0 {
		t.Fatal("BuildAgentCommand returned empty command")
	}
	if cmd[0] != "/env/bin/claude" {
		t.Errorf("binary = %q, want /env/bin/claude", cmd[0])
	}
}

func TestBuildAgentCommandNoPrefix(t *testing.T) {
	config := &types.SandboxConfig{
		Agent: &types.AgentConfig{Type: types.AgentOpencode, Binary: "opencode"},
	}

	cmd := BuildAgentCommand(config, "")
	if len(cmd) == 0 || cmd[0] != "opencode" {
		t.Errorf("binary = %v, want opencode first", cmd)
	}
}

func TestBuildProviderEnvNilProvider(t *testing.T) {
	config := &types.SandboxConfig{
		Agent: &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"},
	}

	env, err := BuildProviderEnv(config)
	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v, want nil", err)
	}
	if len(env) != 0 {
		t.Errorf("env = %v, want empty", env)
	}
}
