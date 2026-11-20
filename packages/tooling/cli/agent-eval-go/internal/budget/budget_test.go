package budget_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

func TestDefaults(t *testing.T) {
	l := budget.Limits{}.Resolve()
	if l.MaxModelRuns != budget.DefaultMaxModelRuns || l.MaxCostUSD != budget.DefaultMaxCostUSD {
		t.Fatalf("%+v", l)
	}
	kept := budget.Limits{MaxModelRuns: 3, MaxCostUSD: 0.5}.Resolve()
	if kept.MaxModelRuns != 3 || kept.MaxCostUSD != 0.5 {
		t.Fatalf("%+v", kept)
	}
}

func TestAdmitStopsAtTheRunLimitWithoutReserving(t *testing.T) {
	g := budget.New(budget.Limits{MaxModelRuns: 5})
	if err := g.Admit(3); err != nil {
		t.Fatal(err)
	}
	if err := g.Admit(3); !errors.Is(err, budget.ErrExceeded) {
		t.Fatalf("%v", err)
	}
	if err := g.Admit(2); err != nil {
		t.Fatal("a refused ask reserves nothing")
	}
	if runs, _ := g.Spent(); runs != 5 {
		t.Fatal(runs)
	}
}

func TestAdmitStopsOnceTheCostIsSpent(t *testing.T) {
	g := budget.New(budget.Limits{MaxCostUSD: 0.10})
	g.Record(model.Usage{CostUSD: 0.06, InputTokens: 1})
	if err := g.Admit(1); err != nil {
		t.Fatal(err)
	}
	g.Record(model.Usage{CostUSD: 0.05, OutputTokens: 2})
	if err := g.Admit(1); !errors.Is(err, budget.ErrExceeded) {
		t.Fatalf("%v", err)
	}
	if _, usage := g.Spent(); usage.InputTokens != 1 || usage.OutputTokens != 2 {
		t.Fatalf("%+v", usage)
	}
}
