package agentcmd

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestBuildAgentCommandBinaryPrefix(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"}

	cmd, err := BuildAgentCommand(agent, nil, []string{"--foo"}, "/env/bin")
	if err != nil {
		t.Fatalf("BuildAgentCommand err = %v", err)
	}
	if len(cmd) == 0 {
		t.Fatal("BuildAgentCommand returned empty command")
	}
	if cmd[0] != "/env/bin/claude" {
		t.Errorf("binary = %q, want /env/bin/claude", cmd[0])
	}
}

func TestBuildAgentCommandNoPrefix(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentOpencode, Binary: "opencode"}

	cmd, err := BuildAgentCommand(agent, nil, nil, "")
	if err != nil {
		t.Fatalf("BuildAgentCommand err = %v", err)
	}
	if len(cmd) == 0 || cmd[0] != "opencode" {
		t.Errorf("binary = %v, want opencode first", cmd)
	}
}

func TestBuildProviderEnvNilProvider(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"}

	env, err := BuildProviderEnv(agent, nil, func(string) string { return "" })
	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v, want nil", err)
	}
	if len(env) != 0 {
		t.Errorf("env = %v, want empty", env)
	}
}

func TestBuildAgentCommandRejectsUnknownAgent(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentType("unknown"), Binary: "x"}

	if _, err := BuildAgentCommand(agent, nil, nil, ""); err == nil {
		t.Fatal("BuildAgentCommand err = nil, want unknown agent error")
	}
}

func TestBuildProviderEnvAppliesAgentEnvironment(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentClaude, Binary: "claude"}
	provider := &types.ModelProvider{
		Name:                "custom",
		CredentialSourceEnv: "SOURCE",
		CredentialTargetEnv: "TARGET",
	}

	env, err := BuildProviderEnv(agent, provider, func(key string) string {
		if key == "SOURCE" {
			return "secret"
		}
		return ""
	})

	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v", err)
	}
	if env["TARGET"] != "secret" {
		t.Errorf("TARGET = %q, want secret", env["TARGET"])
	}
}
