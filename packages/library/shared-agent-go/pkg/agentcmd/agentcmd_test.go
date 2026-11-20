package agentcmd

import (
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
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

func TestBuildAgentCommandBinaryPrefix(t *testing.T) {
	cmd, err := BuildAgentCommand(agentOf(t, types.AgentClaude), nil, []string{"--foo"}, "/env/bin")
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
	cmd, err := BuildAgentCommand(agentOf(t, types.AgentOpencode), nil, nil, "")
	if err != nil {
		t.Fatalf("BuildAgentCommand err = %v", err)
	}
	if len(cmd) == 0 || cmd[0] != "opencode" {
		t.Errorf("binary = %v, want opencode first", cmd)
	}
}

func TestBuildAgentCommandFromDeclaredFacts(t *testing.T) {
	codexProvider, err := providers.GetProvider("gpt5-codex", types.AgentCodex)
	if err != nil {
		t.Fatalf("GetProvider err = %v", err)
	}
	tests := []struct {
		name     string
		agent    types.AgentType
		provider *types.ModelProvider
		args     []string
		want     []string
	}{
		{"kiro starts the chat program", types.AgentKiro, nil, []string{"hi"},
			[]string{"/env/bin/kiro-cli-chat", "chat", "--agent-engine", "v3", "hi"}},
		{"copilot adds nothing", types.AgentCopilot, nil, []string{"hi"},
			[]string{"/env/bin/copilot", "hi"}},
		{"pi adds nothing", types.AgentPi, nil, nil,
			[]string{"/env/bin/pi"}},
		{"codex turns its sandbox off before the provider arguments", types.AgentCodex, codexProvider, []string{"hi"},
			append(append([]string{"/env/bin/codex", "--sandbox", "danger-full-access"}, codexProvider.CliArgs...), "hi")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := BuildAgentCommand(agentOf(t, test.agent), test.provider, test.args, "/env/bin")
			if err != nil {
				t.Fatalf("BuildAgentCommand err = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("command = %v, want %v", got, test.want)
			}
		})
	}
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

func TestBuildAgentCommandRejectsUnknownAgent(t *testing.T) {
	agent := &types.AgentConfig{Type: types.AgentType("unknown"), Binary: "x"}

	if _, err := BuildAgentCommand(agent, nil, nil, ""); err == nil {
		t.Fatal("BuildAgentCommand err = nil, want unknown agent error")
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
