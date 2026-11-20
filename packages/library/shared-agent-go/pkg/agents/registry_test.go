package agents

import (
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestGetAgent(t *testing.T) {
	agent, err := GetAgent(types.AgentClaude)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	if agent.Type != types.AgentClaude || agent.Binary != "claude" {
		t.Fatalf("GetAgent() = %+v, want Claude agent", agent)
	}
}

func TestGetAgentReturnsIndependentConfigurations(t *testing.T) {
	first, err := GetAgent(types.AgentClaude)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	first.Binary = "changed"

	second, err := GetAgent(types.AgentClaude)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}
	if second.Binary == "changed" {
		t.Fatal("GetAgent() returned shared mutable configuration")
	}
}

func TestGetAgentRejectsUnknownType(t *testing.T) {
	agent, err := GetAgent(types.AgentType("unknown"))
	if err == nil || agent != nil {
		t.Fatalf("GetAgent() = (%+v, %v), want nil agent and error", agent, err)
	}
}

func TestGetAgentTypesReturnsSortedTypes(t *testing.T) {
	got := GetAgentTypes()
	want := []types.AgentType{types.AgentClaude, types.AgentOpencode}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAgentTypes() = %v, want %v", got, want)
	}
}

func TestBuildClaudeArgs(t *testing.T) {
	got := BuildClaudeArgs([]string{"review"}, types.AgentExecOptions{Sandbox: true, ProviderCliArgs: []string{"ignored"}})
	want := []string{"--permission-mode", "bypassPermissions", "review"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildClaudeArgs() = %v, want %v", got, want)
	}
}

func TestBuildClaudeEnvReturnsCopy(t *testing.T) {
	providerEnv := map[string]string{"TOKEN": "secret"}
	env := BuildClaudeEnv(providerEnv)
	env["TOKEN"] = "changed"
	if providerEnv["TOKEN"] != "secret" {
		t.Fatal("BuildClaudeEnv() mutated provider environment")
	}
}

func TestBuildOpencodeArgs(t *testing.T) {
	got := BuildOpencodeArgs([]string{"review"}, types.AgentExecOptions{ProviderCliArgs: []string{"--model", "gemini"}})
	want := []string{"--model", "gemini", "review"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildOpencodeArgs() = %v, want %v", got, want)
	}
}

func TestBuildOpencodeEnvIsEmpty(t *testing.T) {
	if got := BuildOpencodeEnv(map[string]string{"TOKEN": "secret"}); len(got) != 0 {
		t.Fatalf("BuildOpencodeEnv() = %v, want empty environment", got)
	}
}

func TestLookupReturnsBehaviorOfEachAgentType(t *testing.T) {
	for _, agentType := range GetAgentTypes() {
		agent, err := Lookup(agentType)
		if err != nil {
			t.Fatalf("Lookup(%s) error = %v", agentType, err)
		}
		if agent.Config().Type != agentType {
			t.Errorf("Lookup(%s).Config().Type = %s", agentType, agent.Config().Type)
		}
		if agent.BuildArgs == nil || agent.BuildEnv == nil {
			t.Errorf("Lookup(%s) has a nil builder", agentType)
		}
	}
}

func TestLookupRejectsUnknownType(t *testing.T) {
	if _, err := Lookup(types.AgentType("unknown")); err == nil {
		t.Fatal("Lookup() error = nil, want unknown agent type error")
	}
}
