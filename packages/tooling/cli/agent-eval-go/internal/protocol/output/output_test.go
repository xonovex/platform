package output_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/llm"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/output"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/protocoltest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

func evalDoc(tier string) string {
	return `{"version":1,"skill":"alpha",` + tier + `"judge":{"harness":"claude","model":"judge-m","rubric":"Be strict."},"cases":[
 {"id":"one","prompt":"do one","gate":"required","fixtures":[{"path":"in.txt","content":"fx"}],
  "expect":[{"kind":"session","event":"command","match":"ls"},{"kind":"llm","criterion":"is good"}]},
 {"id":"two","prompt":"do two","gate":"advisory","expect":[{"kind":"llm","criterion":"is fine"}]}]}`
}

func files(tier string) map[string]string {
	return map[string]string{
		"/p/core/.claude-plugin/plugin.json":    `{"name":"core","skills":"./skills/"}`,
		"/p/core/skills/alpha/SKILL.md":         "a",
		"/p/core/skills/alpha/eval-output.json": evalDoc(tier),
		"/p/core/skills/beta/SKILL.md":          "b",
	}
}

// answer builds an executor: a judge run passes or fails every assertion; a
// generate run reports whether the skill was loaded.
func answer(judgePass, loaded bool) *protocoltest.Exec {
	return &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		if r.Role == harness.RoleJudge {
			p := "false"
			if judgePass {
				p = "true"
			}
			items := make([]string, r.Assertions)
			for i := range items {
				items[i] = `{"text":"x","passed":` + p + `,"evidence":"e"}`
			}
			return harness.Observation{Structured: `{"assertion_results":[` + strings.Join(items, ",") + `]}`, Usage: model.Usage{CostUSD: 0.01}}, nil
		}
		return harness.Observation{
			Text: "answer", TargetInvoked: loaded && r.WithSkill, Usage: model.Usage{CostUSD: 0.02},
			Events: []harness.Event{{Kind: harness.EventCommand, Text: "ls"}},
		}, nil
	}}
}

func input(t *testing.T, e harness.Executor, tier string) protocol.Input {
	env := protocoltest.NewEnv(t, files(tier), []string{"/p"}, e, llm.New(judge.Grader{Executor: e}))
	return protocol.Input{
		Target: "alpha", Env: env,
		Settings: settings.Settings{Harness: "claude", Model: "m", JudgeModel: "jm", Runs: 1, Concurrency: 2},
	}
}

func run(t *testing.T, plan protocol.Plan) []model.CaseResult {
	var cases []model.CaseResult
	for _, r := range protocoltest.RunAll(t, plan) {
		if r.Err != nil {
			t.Fatal(r.Err)
		}
		cases = append(cases, r.Result)
	}
	return cases
}

func TestPlanMakesAUnitPerCaseAndArm(t *testing.T) {
	plan, err := output.New().Plan(input(t, answer(true, true), ""))
	if err != nil || len(plan.Units) != 4 || plan.Subject != "alpha" {
		t.Fatalf("%v %+v", err, plan)
	}
	keys := []string{plan.Units[0].Key(), plan.Units[1].Key(), plan.Units[2].Key(), plan.Units[3].Key()}
	if strings.Join(keys, ",") != "one@present,one@absent,two@present,two@absent" || plan.Units[0].ModelRuns != 2 {
		t.Fatalf("%v", keys)
	}
}

func TestSkillArmOffersTheSkillAndStagesFixtures(t *testing.T) {
	exec := answer(true, true)
	plan, _ := output.New().Plan(input(t, exec, ""))
	_ = run(t, plan)
	var present, absent, judged harness.Request
	for _, c := range exec.Calls() {
		switch {
		case c.Role == harness.RoleJudge && judged.Prompt == "":
			judged = c
		case c.Role == harness.RoleGenerate && c.WithSkill && present.Prompt == "":
			present = c
		case c.Role == harness.RoleGenerate && !c.WithSkill && absent.Prompt == "":
			absent = c
		}
	}
	if present.Target != "alpha" || len(present.Offer) != 1 || len(present.Fixtures) != 1 || present.Fixtures[0].Content != "fx" ||
		present.Prompt != "do one\n\nRelevant input files (read them as needed):\n- in.txt" || absent.Prompt != present.Prompt || judged.Model != "judge-m" || !strings.HasPrefix(judged.Prompt, "Be strict.") {
		t.Fatalf("%+v\n%+v\n%+v", present, absent, judged)
	}
}

func TestGatePassesWhenTheSkillHelps(t *testing.T) {
	// The judge passes the skill arm's answers and fails the control's through the
	// session predicate, which only the control lacks.
	exec := &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		if r.Role == harness.RoleJudge {
			return answer(true, true).Respond(r)
		}
		obs := harness.Observation{Text: "answer", TargetInvoked: r.WithSkill}
		if r.WithSkill {
			obs.Events = []harness.Event{{Kind: harness.EventCommand, Text: "ls"}}
		}
		return obs, nil
	}}
	plan, _ := output.New().Plan(input(t, exec, ""))
	cases := run(t, plan)
	v := plan.Decide(cases)
	if !v.Pass || !strings.HasPrefix(v.Lines[0], "with the skill: 100%") {
		t.Fatalf("%+v", v)
	}
	var present model.CaseResult
	for _, c := range cases {
		if c.Key() == "one@present" {
			present = c
		}
	}
	if present.SkillTriggered == nil || !*present.SkillTriggered || len(present.Predicates) != 2 || present.Usage == nil {
		t.Fatalf("%+v", present)
	}
}

func TestGateFailsWithoutLiftAndWhenTheSkillIsNotLoaded(t *testing.T) {
	plan, _ := output.New().Plan(input(t, answer(true, true), ""))
	if v := plan.Decide(run(t, plan)); v.Pass {
		t.Fatalf("no lift over the control must fail: %+v", v)
	}
	notLoaded, _ := output.New().Plan(input(t, answer(true, false), ""))
	cases := run(t, notLoaded)
	v := notLoaded.Decide(cases)
	if v.Pass || !contains(v.Lines, "skill not loaded: one") || !contains(v.Lines, "failed: one") {
		t.Fatalf("%+v", v)
	}
	for _, c := range cases {
		if c.Key() == "one@present" && (c.Outcome != model.Failed || c.Error == "") {
			t.Fatalf("%+v", c)
		}
	}
}

func contains(lines []string, want string) bool {
	for _, l := range lines {
		if l == want {
			return true
		}
	}
	return false
}

func TestMinimumWinRate(t *testing.T) {
	exec := &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		if r.Role == harness.RoleJudge {
			return answer(true, true).Respond(r)
		}
		obs := harness.Observation{TargetInvoked: r.WithSkill}
		if r.WithSkill {
			obs.Events = []harness.Event{{Kind: harness.EventCommand, Text: "ls"}}
		}
		return obs, nil
	}}
	in := input(t, exec, "")
	in.Settings.MinimumWinRate = 1
	plan, _ := output.New().Plan(in)
	if v := plan.Decide(run(t, plan)); !v.Pass || !contains(v.Lines, "win rate: 100% (minimum 100%)") {
		t.Fatalf("%+v", v)
	}
	flat := input(t, answer(true, true), "")
	flat.Settings.MinimumWinRate = 0.5
	flatPlan, _ := output.New().Plan(flat)
	if v := flatPlan.Decide(run(t, flatPlan)); v.Pass {
		t.Fatalf("%+v", v)
	}
}

func TestTierSelectsThePolicy(t *testing.T) {
	if output.Policies[model.Conservative].MinimumWithSkill != 0.90 || output.Policies[model.Aggressive].MinimumLift != 0.05 {
		t.Fatal("tier table")
	}
	if _, err := output.New().Plan(input(t, answer(true, true), `"tier":"conservative",`)); err != nil {
		t.Fatal(err)
	}
}

func TestJudgeDefaultsToTheSettings(t *testing.T) {
	f := files("")
	f["/p/core/skills/alpha/eval-output.json"] = `{"version":1,"skill":"alpha","cases":[{"id":"x","prompt":"p","gate":"required","expect":[{"kind":"llm","criterion":"c"}]}]}`
	exec := answer(true, true)
	in := protocol.Input{Target: "alpha", Env: protocoltest.NewEnv(t, f, []string{"/p"}, exec, llm.New(judge.Grader{Executor: exec})),
		Settings: settings.Settings{Harness: "codex", JudgeModel: "jm", Runs: 1}}
	plan, err := output.New().Plan(in)
	if err != nil {
		t.Fatal(err)
	}
	_ = run(t, plan)
	for _, c := range exec.Calls() {
		if c.Role == harness.RoleJudge && (c.Harness != "codex" || c.Model != "jm") {
			t.Fatalf("%+v", c)
		}
	}
}

func TestPlanErrorsAndFaults(t *testing.T) {
	if _, err := output.New().Plan(input(t, answer(true, true), "")); err != nil {
		t.Fatal(err)
	}
	for name, target := range map[string]string{"unknown": "zeta", "no document": "beta"} {
		in := input(t, answer(true, true), "")
		in.Target = target
		if _, err := output.New().Plan(in); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	f := files("")
	f["/p/core/skills/alpha/eval-output.json"] = "{}"
	bad := protocol.Input{Target: "alpha", Env: protocoltest.NewEnv(t, f, []string{"/p"}, answer(true, true))}
	if _, err := output.New().Plan(bad); err == nil {
		t.Fatal("invalid document")
	}
	down := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) { return harness.Observation{}, errors.New("down") }}
	plan, _ := output.New().Plan(input(t, down, ""))
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("executor fault")
	}
	if output.New().Name() != model.Output {
		t.Fatal("name")
	}
}
