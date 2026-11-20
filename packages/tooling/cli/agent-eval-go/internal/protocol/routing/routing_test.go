package routing_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/protocoltest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/routing"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

func trig(skill, id, prompt string, should bool) string {
	b := "false"
	if should {
		b = "true"
	}
	return `{"version":1,"skill":"` + skill + `","queries":[{"id":"` + id + `","prompt":"` + prompt + `","shouldTrigger":` + b + `,"split":"validation"}]}`
}

func files() map[string]string {
	return map[string]string{
		"/p/core/.claude-plugin/plugin.json":      `{"name":"core","skills":"./skills/"}`,
		"/p/core/skills/alpha/SKILL.md":           "a",
		"/p/core/skills/alpha/eval-trigger.json":  trig("alpha", "q", "fix the bug", true),
		"/p/core/skills/beta/SKILL.md":            "b",
		"/p/core/skills/beta/eval-trigger.json":   trig("beta", "q", "fix the bug", false),
		"/p/other/.claude-plugin/plugin.json":     `{"name":"other","skills":"./skills/"}`,
		"/p/other/skills/gamma/SKILL.md":          "g",
		"/p/other/skills/gamma/eval-trigger.json": trig("gamma", "q", "unrelated", true),
	}
}

func in(t *testing.T, e harness.Executor, target string) protocol.Input {
	return protocol.Input{
		Target:   target,
		Settings: settings.Settings{Harness: "claude", Model: "m", Runs: 2, Threshold: 0.5},
		Env:      protocoltest.NewEnv(t, files(), []string{"/p"}, e),
	}
}

func answer(selected string) *protocoltest.Exec {
	return &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) {
		return harness.Observation{Selected: selected, Usage: model.Usage{CostUSD: 0.01}}, nil
	}}
}

func TestScenarioOffersAllCandidatesAndWantsThePositive(t *testing.T) {
	exec := answer("target")
	plan, err := routing.New().Plan(in(t, exec, ""))
	if err != nil || len(plan.Units) != 1 || plan.Subject != "all" || !strings.HasPrefix(plan.Units[0].ID, "route-") {
		t.Fatalf("%v %+v", err, plan)
	}
	res := protocoltest.RunAll(t, plan)[0]
	if res.Err != nil || res.Result.Outcome != model.Passed || !*res.Result.SkillTriggered || len(res.Result.Selected) != 2 {
		t.Fatalf("%+v", res)
	}
	call := exec.Calls()[0]
	var dirs []string
	for _, o := range call.Offer {
		dirs = append(dirs, o.SkillDirs...)
	}
	if call.Target != "alpha" || len(dirs) != 2 || dirs[0] != "/p/core/skills/alpha" || dirs[1] != "/p/core/skills/beta" {
		t.Fatalf("%+v", call)
	}
	if v := plan.Decide([]model.CaseResult{res.Result}); !v.Pass {
		t.Fatalf("%+v", v)
	}
}

func TestWrongSkillFailsTheScenario(t *testing.T) {
	plan, err := routing.New().Plan(in(t, answer("competitor:beta"), "core"))
	if err != nil || plan.Subject != "core" {
		t.Fatalf("%v", err)
	}
	res := protocoltest.RunAll(t, plan)[0].Result
	if res.Outcome != model.Failed || *res.SkillTriggered {
		t.Fatalf("%+v", res)
	}
	if v := plan.Decide([]model.CaseResult{res}); v.Pass || !strings.HasPrefix(v.Lines[1], "failed: route-") {
		t.Fatalf("%+v", v)
	}
}

func TestPlanErrors(t *testing.T) {
	if _, err := routing.New().Plan(in(t, answer(""), "nope")); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("%v", err)
	}
	if _, err := routing.New().Plan(in(t, answer(""), "other")); err == nil || !strings.Contains(err.Error(), "no routing scenario") {
		t.Fatalf("%v", err)
	}
	i := in(t, answer(""), "")
	i.Settings.Split = model.Test
	if _, err := routing.New().Plan(i); err == nil {
		t.Fatal("no scenario in the split")
	}
	f := files()
	f["/p/core/skills/beta/eval-trigger.json"] = "{}"
	bad := protocol.Input{Target: "core", Settings: i.Settings, Env: protocoltest.NewEnv(t, f, []string{"/p"}, answer(""))}
	if _, err := routing.New().Plan(bad); err == nil {
		t.Fatal("an invalid trigger document is an error")
	}
	if routing.New().Name() != model.Routing {
		t.Fatal("name")
	}
}

func TestExecutorFaultIsReturned(t *testing.T) {
	exec := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) { return harness.Observation{}, errors.New("down") }}
	plan, _ := routing.New().Plan(in(t, exec, ""))
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("want an error")
	}
}
