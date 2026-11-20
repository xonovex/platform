package judge_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
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

var two = []string{"first", "second"}

const good = `{"assertion_results":[{"text":"first","passed":true,"evidence":"e1"},{"text":"second","passed":false,"evidence":"e2"}]}`

func TestGradeReadsStructuredAnswer(t *testing.T) {
	e := &exec{obs: harness.Observation{Structured: good, Usage: model.Usage{CostUSD: 0.01}}}
	verdicts, usage, err := judge.Grader{Executor: e}.Grade(context.Background(),
		judge.Spec{Harness: "claude", Model: "m", BudgetUSD: 0.1}, "task", "response", two)
	if err != nil || usage.CostUSD != 0.01 {
		t.Fatalf("%v %+v", err, usage)
	}
	if !verdicts[0].Passed || verdicts[1].Passed || verdicts[1].Evidence != "e2" {
		t.Fatalf("%+v", verdicts)
	}
	if e.got.Role != harness.RoleJudge || e.got.Assertions != 2 || e.got.Model != "m" || e.got.Caps.BudgetUSD != 0.1 ||
		!strings.Contains(e.got.Prompt, "1. first") || !strings.Contains(e.got.Prompt, "response") {
		t.Fatalf("%+v", e.got)
	}
}

func TestParseFallsBackToTextAndFences(t *testing.T) {
	for name, obs := range map[string]harness.Observation{
		"text":   {Text: "Here you go: " + good + " done"},
		"fenced": {Text: "```json\n" + good + "\n```"},
		"bare":   {Text: "```\n" + good + "\n```"},
	} {
		v := judge.Parse(obs, two)
		if !v[0].Passed || v[1].Passed {
			t.Errorf("%s: %+v", name, v)
		}
	}
}

func TestUnreadableAnswersFailEveryAssertion(t *testing.T) {
	for name, obs := range map[string]harness.Observation{
		"empty":      {},
		"prose":      {Text: "all good"},
		"bad json":   {Text: "{not json}"},
		"short":      {Structured: `{"assertion_results":[{"text":"a","passed":true,"evidence":"x"}]}`},
		"no passed":  {Structured: `{"assertion_results":[{"text":"a"},{"text":"b","passed":true}]}`},
		"fence none": {Text: "```json\nnot an object\n```"},
	} {
		for _, v := range judge.Parse(obs, two) {
			if v.Passed || v.Evidence != judge.UnparseableEvidence {
				t.Errorf("%s: %+v", name, v)
			}
		}
	}
}

func TestPromptUsesCustomRubric(t *testing.T) {
	p := judge.Prompt("Be strict.", "t", "r", two)
	if !strings.HasPrefix(p, "Be strict.") || strings.Contains(p, "partial credit") {
		t.Fatal(p)
	}
	if !strings.Contains(judge.Prompt("  ", "t", "r", two), "partial credit") {
		t.Fatal("blank rubric uses the default")
	}
}

func TestGradeEdgeCases(t *testing.T) {
	e := &exec{}
	if v, _, err := (judge.Grader{Executor: e}).Grade(context.Background(), judge.Spec{}, "t", "r", nil); err != nil || v != nil {
		t.Fatal("no assertions, no run")
	}
	e.err = errors.New("down")
	if _, _, err := (judge.Grader{Executor: e}).Grade(context.Background(), judge.Spec{}, "t", "r", two); err == nil {
		t.Fatal("infrastructure fault is an error")
	}
}
