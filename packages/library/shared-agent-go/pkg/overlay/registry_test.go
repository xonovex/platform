package overlay

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestEveryAgentHasADialect(t *testing.T) {
	registered := AgentTypes()
	declared := agents.GetAgentTypes()
	if len(registered) != len(declared) {
		t.Fatalf("dialects %v, agents %v", registered, declared)
	}
	for index, agentType := range declared {
		if registered[index] != agentType {
			t.Errorf("dialect %s does not match agent %s", registered[index], agentType)
		}
		if _, err := Resolve(agentType); err != nil {
			t.Errorf("Resolve(%s): %v", agentType, err)
		}
	}
}

func TestResolveRejectsAnUnknownAgent(t *testing.T) {
	if _, err := Resolve(types.AgentType("nope")); err == nil {
		t.Error("an unknown agent resolved")
	}
}

func TestCheckRefusesALoginAgentWhoseHomeTheOverlayReplaces(t *testing.T) {
	for _, agentType := range AgentTypes() {
		agent, err := agents.GetAgent(agentType)
		if err != nil {
			t.Fatal(err)
		}
		dialect, _ := Resolve(agentType)
		refused := dialect.Facts().ReplacesHome && agent.Access == types.AccessLogin
		if err := Check(agent); (err != nil) != refused {
			t.Errorf("Check(%s) = %v, want refused = %v", agentType, err, refused)
		}
	}
	kiro, _ := agents.GetAgent(types.AgentKiro)
	if err := Check(kiro); err == nil || !strings.Contains(err.Error(), "login") {
		t.Errorf("Check(kiro) = %v", err)
	}
	if err := Check(&types.AgentConfig{Type: "nope"}); err == nil {
		t.Error("an agent without a dialect passed")
	}
}
