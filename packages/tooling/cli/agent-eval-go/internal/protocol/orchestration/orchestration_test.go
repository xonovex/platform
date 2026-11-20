package orchestration_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/orchestration"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/protocoltest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

func flow(budget string) string {
	return `{"version":1,"id":"review-flow",` + budget + `
 "agents":[{"id":"writer","harness":"claude","model":"writer-m"},{"id":"reviewer","harness":"codex"}],
 "steps":[
  {"id":"review","prompt":"review it","agent":"reviewer","after":["write"],
   "expect":[{"kind":"session","event":"message","match":"approved"}]},
  {"id":"write","prompt":"write it","agent":"writer",
   "expect":[{"kind":"filesystem","path":"out.md","check":"exists"}]}]}`
}

func input(t *testing.T, e harness.Executor, doc string) protocol.Input {
	return protocol.Input{
		Target: "/f/flow.json", Env: protocoltest.NewEnv(t, map[string]string{"/f/flow.json": doc}, nil, e),
		Settings: settings.Settings{Harness: "claude", Model: "default-m", JudgeModel: "jm", Runs: 1},
	}
}

func agents() *protocoltest.Exec {
	return &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		if strings.HasPrefix(r.Prompt, "write it") {
			return harness.Observation{Text: "wrote", Files: map[string]string{"out.md": "doc", "b.md": "b"}, Usage: model.Usage{CostUSD: 0.01}}, nil
		}
		return harness.Observation{
			Text: "done", Usage: model.Usage{CostUSD: 0.01},
			Events: []harness.Event{{Kind: harness.EventMessage, Text: "approved"}},
		}, nil
	}}
}

func TestStepsRunInDependencyOrderWithPriorResults(t *testing.T) {
	exec := agents()
	plan, err := orchestration.New().Plan(input(t, exec, flow("")))
	if err != nil || len(plan.Units) != 1 || plan.Subject != "review-flow" || plan.Units[0].ID != "review-flow" || plan.Units[0].ModelRuns != 2 {
		t.Fatalf("%v %+v", err, plan)
	}
	res := protocoltest.RunAll(t, plan)[0]
	if res.Err != nil || res.Result.Outcome != model.Passed || len(res.Result.Predicates) != 2 || *res.Result.Score != 1 || res.Result.Usage.CostUSD != 0.02 {
		t.Fatalf("%+v", res)
	}
	calls := exec.Calls()
	if len(calls) != 2 || calls[0].Harness != "claude" || calls[0].Model != "writer-m" || calls[1].Harness != "codex" || calls[1].Model != "" {
		t.Fatalf("%+v", calls)
	}
	review := calls[1]
	if !strings.Contains(review.Prompt, "## Result of step write\n\nwrote") || len(review.Fixtures) != 2 ||
		review.Fixtures[0].Path != "b.md" || review.Fixtures[1].Content != "doc" || review.Role != harness.RoleAgent {
		t.Fatalf("%+v", review)
	}
	if v := plan.Decide([]model.CaseResult{res.Result}); !v.Pass || !strings.HasPrefix(v.Lines[0], "1 of 1") {
		t.Fatalf("%+v", v)
	}
}

func TestAnAgentWithTheSettingsHarnessUsesTheSettingsModel(t *testing.T) {
	exec := agents()
	plan, _ := orchestration.New().Plan(input(t, exec, strings.Replace(flow(""), `"model":"writer-m"`, `"role":"w"`, 1)))
	_ = protocoltest.RunAll(t, plan)
	if exec.Calls()[0].Model != "default-m" {
		t.Fatalf("%+v", exec.Calls()[0])
	}
}

func TestFailingExpectationNamesTheStep(t *testing.T) {
	exec := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) { return harness.Observation{Text: "x"}, nil }}
	plan, _ := orchestration.New().Plan(input(t, exec, flow("")))
	res := protocoltest.RunAll(t, plan)[0].Result
	if res.Outcome != model.Failed || !strings.HasPrefix(res.Predicates[0].Detail, "step write:") {
		t.Fatalf("%+v", res)
	}
	if v := plan.Decide([]model.CaseResult{res}); v.Pass {
		t.Fatalf("%+v", v)
	}
}

func TestSeveralRunsMakeSeveralUnits(t *testing.T) {
	in := input(t, agents(), flow(""))
	in.Settings.Runs = 2
	plan, _ := orchestration.New().Plan(in)
	if len(plan.Units) != 2 || plan.Units[0].ID != "review-flow-run1" || plan.Units[1].ID != "review-flow-run2" {
		t.Fatalf("%+v", plan.Units)
	}
}

func TestFlowBudgetStopsTheFlow(t *testing.T) {
	plan, _ := orchestration.New().Plan(input(t, agents(), flow(`"budget":{"maxModelRuns":1},`)))
	res := protocoltest.RunAll(t, plan)[0].Result
	if res.Outcome != model.Failed || !strings.Contains(res.Error, "2 model runs of 1") {
		t.Fatalf("%+v", res)
	}
	plan, _ = orchestration.New().Plan(input(t, agents(), flow(`"budget":{"maxCostUsd":0.005},`)))
	res = protocoltest.RunAll(t, plan)[0].Result
	if res.Outcome != model.Failed || !strings.Contains(res.Error, "$0.0100 of $0.01") {
		t.Fatalf("%+v", res)
	}
}

func TestFlowTimeoutFailsTheCaseButCancellationIsAnError(t *testing.T) {
	slow := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) {
		return harness.Observation{}, context.DeadlineExceeded
	}}
	plan, _ := orchestration.New().Plan(input(t, slow, flow(`"budget":{"timeoutSeconds":1},`)))
	res := protocoltest.RunAll(t, plan)[0]
	if res.Err != nil || res.Result.Error != "flow timed out" || res.Result.Outcome != model.Failed {
		t.Fatalf("%+v", res)
	}
	down := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) { return harness.Observation{}, errors.New("down") }}
	plan, _ = orchestration.New().Plan(input(t, down, flow("")))
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("an executor fault is an error")
	}
	plan, _ = orchestration.New().Plan(input(t, slow, flow("")))
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("a deadline that is not the flow's own is an error")
	}
}

func TestPredicateFaultIsAnError(t *testing.T) {
	doc := strings.Replace(flow(""), `"kind":"session","event":"message","match":"approved"`, `"kind":"llm","criterion":"c"`, 1)
	plan, err := orchestration.New().Plan(input(t, agents(), doc))
	if err != nil || plan.Units[0].ModelRuns != 3 {
		t.Fatalf("%v %+v", err, plan.Units)
	}
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("no llm evaluator is registered")
	}
}

func TestPlanErrors(t *testing.T) {
	in := input(t, agents(), "{}")
	if _, err := orchestration.New().Plan(in); err == nil {
		t.Fatal("invalid document")
	}
	cyclic := `{"version":1,"id":"x","agents":[{"id":"a","harness":"claude"}],"steps":[
	 {"id":"one","prompt":"p","agent":"a","after":["two"]},{"id":"two","prompt":"p","agent":"a","after":["one"]}]}`
	if _, err := orchestration.New().Plan(input(t, agents(), cyclic)); err == nil {
		t.Fatal("cycle")
	}
	if orchestration.New().Name() != model.Orchestration {
		t.Fatal("name")
	}
}
