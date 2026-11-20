package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/llm"
)

type exec struct {
	obs harness.Observation
	err error
	got harness.Request
}

func (e *exec) Execute(_ context.Context, r harness.Request) (harness.Observation, error) {
	e.got = r
	return e.obs, e.err
}

func TestGradesAllCriteriaInOneCall(t *testing.T) {
	e := &exec{obs: harness.Observation{
		Structured: `{"assertion_results":[{"text":"a","passed":true,"evidence":"ok"},{"text":"b","passed":false,"evidence":"no"}]}`,
		Usage:      model.Usage{CostUSD: 0.02},
	}}
	ev := llm.New(judge.Grader{Executor: e})
	if ev.Kind() != "llm" {
		t.Fatal(ev.Kind())
	}
	got, usage, err := ev.Evaluate(context.Background(),
		[]model.Predicate{{Criterion: "a"}, {Criterion: "b"}},
		predicate.Subject{Task: "t", Obs: harness.Observation{Text: "answer"}, Judge: judge.Spec{Harness: "claude"}})
	if err != nil || !got[0].Passed || got[1].Passed || got[1].Detail != "no" || usage.CostUSD != 0.02 {
		t.Fatalf("%v %+v %+v", err, got, usage)
	}
	if e.got.Assertions != 2 || e.got.Harness != "claude" {
		t.Fatalf("%+v", e.got)
	}
}

func TestJudgeFaultIsAnError(t *testing.T) {
	ev := llm.New(judge.Grader{Executor: &exec{err: errors.New("down")}})
	if _, _, err := ev.Evaluate(context.Background(), []model.Predicate{{Criterion: "a"}}, predicate.Subject{}); err == nil {
		t.Fatal("want an error")
	}
}
