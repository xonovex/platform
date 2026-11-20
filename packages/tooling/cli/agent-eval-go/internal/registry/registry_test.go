package registry_test

import (
	"context"
	"strings"
	"testing"

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
	if got := strings.Join(registry.Harnesses().Names(), ","); got != "claude,codex" {
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
