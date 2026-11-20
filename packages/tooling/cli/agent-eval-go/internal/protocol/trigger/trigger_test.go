package trigger_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/protocoltest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol/trigger"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

const doc = `{"version":1,"skill":"alpha","queries":[
 {"id":"yes","prompt":"load it","shouldTrigger":true,"split":"validation"},
 {"id":"no","prompt":"leave it","shouldTrigger":false,"split":"validation"},
 {"id":"train","prompt":"train query","shouldTrigger":true,"split":"train"}]}`

func files() map[string]string {
	return map[string]string{
		"/p/core/.claude-plugin/plugin.json":     `{"name":"core","skills":"./skills/"}`,
		"/p/core/skills/alpha/SKILL.md":          "a",
		"/p/core/skills/alpha/eval-trigger.json": doc,
		"/p/core/skills/beta/SKILL.md":           "b",
	}
}

func input(t *testing.T, e harness.Executor, s settings.Settings, target string) protocol.Input {
	return protocol.Input{Target: target, Settings: s, Env: protocoltest.NewEnv(t, files(), []string{"/p"}, e)}
}

func base() settings.Settings {
	return settings.Settings{Harness: "claude", Model: "m", Runs: 3, Threshold: 0.5, Concurrency: 2}
}

func selecting(by map[string][]string) *protocoltest.Exec {
	n := map[string]int{}
	return &protocoltest.Exec{Respond: func(r harness.Request) (harness.Observation, error) {
		seq := by[r.Prompt]
		i := n[r.Prompt]
		n[r.Prompt]++
		return harness.Observation{Selected: seq[i%len(seq)], Usage: model.Usage{CostUSD: 0.01}, DurationMs: 3}, nil
	}}
}

func TestPlanAndRun(t *testing.T) {
	exec := selecting(map[string][]string{
		"load it":  {"target", "target", "none"},
		"leave it": {"none", "none", "competitor:beta"},
	})
	s := base()
	s.Split = model.Validation
	plan, err := trigger.New().Plan(input(t, exec, s, "alpha"))
	if err != nil || len(plan.Units) != 2 || plan.Subject != "alpha" || plan.Units[0].ModelRuns != 3 {
		t.Fatalf("%v %+v", err, plan)
	}
	results := protocoltest.RunAll(t, plan)
	yes, no := results[0].Result, results[1].Result
	if yes.Outcome != model.Passed || *yes.Score < 0.66 || !*yes.SkillTriggered || yes.Runs != 3 || len(yes.Selected) != 3 ||
		yes.Usage.CostUSD < 0.029 || yes.DurationMs != 9 {
		t.Fatalf("%+v", yes)
	}
	if no.Outcome != model.Passed || *no.SkillTriggered {
		t.Fatalf("%+v", no)
	}
	calls := exec.Calls()
	if calls[0].Role != harness.RoleTrigger || calls[0].Target != "alpha" || len(calls[0].Offer) != 1 ||
		len(calls[0].Offer[0].SkillDirs) != 1 || calls[0].Offer[0].SkillDirs[0] != "/p/core/skills/alpha" {
		t.Fatalf("only the target is on offer: %+v", calls[0])
	}
	var cases []model.CaseResult
	for _, r := range results {
		cases = append(cases, r.Result)
	}
	if v := plan.Decide(cases); !v.Pass || !strings.HasPrefix(v.Lines[0], "2 of 2") {
		t.Fatalf("%+v", v)
	}
}

func TestAWrongQueryFailsTheVerdict(t *testing.T) {
	exec := selecting(map[string][]string{"load it": {"none"}, "leave it": {"none"}, "train query": {"target"}})
	plan, err := trigger.New().Plan(input(t, exec, base(), "alpha"))
	if err != nil || len(plan.Units) != 3 {
		t.Fatalf("%v %d", err, len(plan.Units))
	}
	var cases []model.CaseResult
	for _, r := range protocoltest.RunAll(t, plan) {
		cases = append(cases, r.Result)
	}
	v := plan.Decide(cases)
	if v.Pass || len(v.Lines) != 2 || v.Lines[1] != "failed: yes" {
		t.Fatalf("%+v", v)
	}
}

func TestPlanErrors(t *testing.T) {
	exec := selecting(nil)
	if _, err := trigger.New().Plan(input(t, exec, base(), "missing")); err == nil {
		t.Fatal("unknown skill")
	}
	if _, err := trigger.New().Plan(input(t, exec, base(), "beta")); err == nil || !strings.Contains(err.Error(), "eval-trigger.json") {
		t.Fatalf("%v", err)
	}
	s := base()
	s.Split = model.Test
	if _, err := trigger.New().Plan(input(t, exec, s, "alpha")); err == nil || !strings.Contains(err.Error(), "no queries") {
		t.Fatalf("%v", err)
	}
	f := files()
	f["/p/core/skills/alpha/eval-trigger.json"] = `{}`
	in := protocol.Input{Target: "alpha", Settings: base(), Env: protocoltest.NewEnv(t, f, []string{"/p"}, exec)}
	if _, err := trigger.New().Plan(in); err == nil {
		t.Fatal("an invalid document is an error")
	}
	if trigger.New().Name() != model.Trigger {
		t.Fatal("name")
	}
}

func TestExecutorFaultIsReturned(t *testing.T) {
	exec := &protocoltest.Exec{Respond: func(harness.Request) (harness.Observation, error) { return harness.Observation{}, errors.New("down") }}
	plan, _ := trigger.New().Plan(input(t, exec, base(), "alpha"))
	if _, err := plan.Units[0].Run(context.Background()); err == nil {
		t.Fatal("want an error")
	}
}
