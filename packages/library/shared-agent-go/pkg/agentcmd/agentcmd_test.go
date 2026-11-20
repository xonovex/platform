package agentcmd

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func agentOf(t *testing.T, agentType types.AgentType) *types.AgentConfig {
	t.Helper()
	agent, err := agents.GetAgent(agentType)
	if err != nil {
		t.Fatalf("GetAgent(%s) err = %v", agentType, err)
	}
	return agent
}

func TestBuildProviderEnvNilProvider(t *testing.T) {
	env, err := BuildProviderEnv(agentOf(t, types.AgentClaude), nil, func(string) string { return "" })
	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v, want nil", err)
	}
	if len(env) != 0 {
		t.Errorf("env = %v, want empty", env)
	}
}

func TestBuildProviderEnvRejectsUnknownAgent(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentType("unknown"), Binary: "x"}

	if _, err := BuildProviderEnv(agent, &types.ModelProvider{Name: "p"}, func(string) string { return "" }); err == nil {
		t.Fatal("BuildProviderEnv err = nil, want unknown agent error")
	}
}

func TestBuildProviderEnvAppliesAgentEnvironment(t *testing.T) {
	provider := &types.ModelProvider{
		Name:                "custom",
		CredentialSourceEnv: "SOURCE",
		CredentialTargetEnv: "TARGET",
	}
	getenv := func(key string) string {
		if key == "SOURCE" {
			return "secret"
		}
		return ""
	}

	env, err := BuildProviderEnv(agentOf(t, types.AgentClaude), provider, getenv)
	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v", err)
	}
	if env["TARGET"] != "secret" {
		t.Errorf("TARGET = %q, want secret", env["TARGET"])
	}

	env, err = BuildProviderEnv(agentOf(t, types.AgentOpencode), provider, getenv)
	if err != nil {
		t.Fatalf("BuildProviderEnv err = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("env = %v, want empty for an agent that reads no provider environment", env)
	}
}

func TestBuildProviderEnvReportsAMissingCredential(t *testing.T) {
	provider := &types.ModelProvider{Name: "custom", CredentialSourceEnv: "SOURCE", CredentialTargetEnv: "TARGET"}

	_, err := BuildProviderEnv(agentOf(t, types.AgentClaude), provider, func(string) string { return "" })

	if err == nil || !strings.Contains(err.Error(), "failed to build provider environment") {
		t.Fatalf("BuildProviderEnv err = %v, want the missing credential reported", err)
	}
}
