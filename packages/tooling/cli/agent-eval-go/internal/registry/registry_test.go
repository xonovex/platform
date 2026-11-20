package registry_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/registry"
)

type exec struct{}

func (exec) Execute(context.Context, harness.Request) (harness.Observation, error) {
	return harness.Observation{Structured: `{"assertion_results":[{"text":"c","passed":true,"evidence":"e"}]}`}, nil
}

func TestEveryLeafIsRegistered(t *testing.T) {
	harnesses, err := registry.Harnesses(agents.Embedded().Get)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(harnesses.Names(), ","); got != "claude,codex" {
		t.Fatal(got)
	}
	names := registry.Protocols().Names()
	if len(names) != len(model.Protocols()) {
		t.Fatalf("every protocol has a leaf: %v", names)
	}
	for _, p := range model.Protocols() {
		if _, err := registry.Protocols().Get(p); err != nil {
			t.Error(err)
		}
	}
}

func TestHarnessesFailOnFactsALeafCannotUse(t *testing.T) {
	missing := func(types.AgentType) (*types.AgentConfig, error) { return nil, errors.New("no facts") }
	if _, err := registry.Harnesses(missing); err == nil || err.Error() != "no facts" {
		t.Errorf("Harnesses() error = %v, want the lookup fault", err)
	}
	bare := func(agentType types.AgentType) (*types.AgentConfig, error) {
		return &types.AgentConfig{Type: agentType}, nil
	}
	if _, err := registry.Harnesses(bare); err == nil {
		t.Error("Harnesses() over facts without a layout must fail")
	}
}

func TestPredicateRegistryEvaluatesEveryKind(t *testing.T) {
	preds := []model.Predicate{
		{Kind: model.KindFilesystem, Path: "a", Check: "absent"},
		{Kind: model.KindSession, Event: "command", Match: "x", Max: new(int)},
		{Kind: model.KindLLM, Criterion: "c"},
	}
	got, _, err := registry.Predicates(exec{}).Evaluate(context.Background(), preds, predicate.Subject{})
	if err != nil || !got[0].Passed || !got[1].Passed || !got[2].Passed {
		t.Fatalf("%v %+v", err, got)
	}
}
