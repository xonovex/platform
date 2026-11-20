package instruction_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/instruction"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/protocoltest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

const rules = "# Rules\n\n## Tests\n\nAlways run the tests before you finish.\n\n## Style\n\nUse tabs.\n"

const evalDoc = `{"version":1,"instruction":"AGENTS.md","cases":[
 {"id":"tests","prompt":"change the code","anchor":{"heading":"Tests"},"ablate":"before you finish",
  "expect":{
   "present":[{"kind":"session","event":"command","match":"test"}],
   "absent":[{"kind":"session","event":"command","match":"test","max":0}],
   "ablated":[{"kind":"session","event":"command","match":"test","max":0}]}}]}`

func files() map[string]string {
	return map[string]string{"/e/eval.json": evalDoc, "/e/AGENTS.md": rules}
}

// follows makes an agent run commands only when the instruction text keeps the
// phrase that asks for them.
func follows(phrase string) *protocoltest.Exec {
	return &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		var events []harness.Event
		if strings.Contains(r.Instructions, phrase) {
			events = []harness.Event{{Kind: harness.EventCommand, Text: "go test"}}
		}
		return harness.Observation{Events: events}, nil
	}}
}

func input(t *testing.T, e harness.Executor, f map[string]string) protocol.Input {
	return protocol.Input{
		Target: "/e/eval.json", Env: protocoltest.NewEnv(t, f, nil, e),
		Settings: settings.Settings{Harness: "claude", Model: "m", Runs: 1, MinimumLift: 0.5},
	}
}

func runAll(t *testing.T, plan protocol.Plan) []model.CaseResult {
	var cases []model.CaseResult
	for _, r := range protocoltest.RunAll(t, plan) {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		cases = append(cases, r.Result)
	}
	return cases
}

func TestARuleThatChangesBehaviorPasses(t *testing.T) {
	exec := follows("Always run the tests before you finish.")
	plan, err := instruction.New().Plan(input(t, exec, files()))
	if err != nil || len(plan.Units) != 3 || plan.Subject != "AGENTS.md" {
		t.Fatalf("%v %+v", err, plan)
	}
	keys := plan.Units[0].Key() + "," + plan.Units[1].Key() + "," + plan.Units[2].Key()
	if keys != "tests@present,tests@absent,tests@ablated" {
		t.Fatal(keys)
	}
	cases := runAll(t, plan)
	v := plan.Decide(cases)
	if !v.Pass || !strings.HasPrefix(v.Lines[0], "mean lift: +1.00 over 2 comparisons") {
		t.Fatalf("%+v", v)
	}
	absent := cases[1]
	if absent.ControlScore == nil || *absent.ControlScore != 0 || *absent.Score != 1 || len(absent.Predicates) != 1 {
		t.Fatalf("the present predicates score the control: %+v", absent)
	}
	var present, removed, ablated string
	for _, c := range exec.Calls() {
		switch {
		case !strings.Contains(c.Instructions, "Tests"):
			removed = c.Instructions
		case !strings.Contains(c.Instructions, "before you finish"):
			ablated = c.Instructions
		default:
			present = c.Instructions
		}
	}
	if present != rules || strings.Contains(removed, "Always run") || !strings.Contains(ablated, "Always run the tests .") {
		t.Fatalf("present=%q removed=%q ablated=%q", present, removed, ablated)
	}
	if calls := exec.Calls(); calls[0].Role != harness.RoleAgent || calls[0].Prompt != "change the code" {
		t.Fatalf("%+v", calls[0])
	}
}

func TestARuleThatChangesNothingFails(t *testing.T) {
	plan, _ := instruction.New().Plan(input(t, follows(""), files()))
	v := plan.Decide(runAll(t, plan))
	if v.Pass || !strings.Contains(strings.Join(v.Lines, "\n"), "mean lift: +0.00") {
		t.Fatalf("%+v", v)
	}
}

func TestAnExpectationThatFailsFailsTheVerdict(t *testing.T) {
	plan, _ := instruction.New().Plan(input(t, follows("zzz"), files()))
	v := plan.Decide(runAll(t, plan))
	if v.Pass || !strings.Contains(strings.Join(v.Lines, "\n"), "failed: tests@present") {
		t.Fatalf("%+v", v)
	}
}

func TestNoComparisonMeansNoLiftFigure(t *testing.T) {
	f := files()
	f["/e/eval.json"] = `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"t","prompt":"p","anchor":{"heading":"Tests"},
	  "expect":{"present":[{"kind":"session","event":"command","match":"test"}]}}]}`
	plan, err := instruction.New().Plan(input(t, follows("Always"), f))
	if err != nil || len(plan.Units) != 1 {
		t.Fatalf("%v", err)
	}
	if v := plan.Decide(runAll(t, plan)); !v.Pass || !strings.HasPrefix(v.Lines[0], "no case compares") {
		t.Fatalf("%+v", v)
	}
}

func TestPlanErrors(t *testing.T) {
	f := files()
	delete(f, "/e/AGENTS.md")
	if _, err := instruction.New().Plan(input(t, follows(""), f)); err == nil {
		t.Fatal("missing instruction file")
	}
	f = files()
	f["/e/AGENTS.md"] = "# Nothing here\n"
	if _, err := instruction.New().Plan(input(t, follows(""), f)); err == nil || !strings.Contains(err.Error(), "tests") {
		t.Fatalf("an anchor that matches nothing is found while planning: %v", err)
	}
	f = files()
	f["/e/eval.json"] = "{}"
	if _, err := instruction.New().Plan(input(t, follows(""), f)); err == nil {
		t.Fatal("invalid document")
	}
	down := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) { return harness.Observation{}, errors.New("down") }}
	plan, _ := instruction.New().Plan(input(t, down, files()))
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("executor fault")
	}
	if instruction.New().Name() != model.Instruction {
		t.Fatal("name")
	}
}

func TestModelPredicatesDoubleTheBudgetEstimate(t *testing.T) {
	f := files()
	f["/e/eval.json"] = `{"version":1,"instruction":"AGENTS.md","cases":[{"id":"t","prompt":"p","anchor":{"heading":"Tests"},
	  "expect":{"present":[{"kind":"llm","criterion":"c"}]}}]}`
	in := input(t, follows(""), f)
	in.Settings.Runs = 2
	plan, _ := instruction.New().Plan(in)
	if plan.Units[0].ModelRuns != 4 {
		t.Fatalf("%d", plan.Units[0].ModelRuns)
	}
}
