package settings_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

func f(v float64) *float64 { return &v }

func TestDefaultsPerProtocol(t *testing.T) {
	for protocol, want := range map[model.Protocol]struct {
		runs  int
		batch int
	}{
		model.Trigger: {3, 8}, model.Routing: {3, 0}, model.Output: {1, 0}, model.Instruction: {1, 0}, model.Orchestration: {1, 0},
	} {
		s, err := settings.Resolve(model.Project{}, settings.Flags{}, nil, protocol)
		if err != nil || s.Runs != want.runs || s.BatchSize != want.batch || s.Harness != "claude" ||
			s.Model != "claude-haiku-5-5" || s.JudgeModel != "claude-sonnet-5-5" || s.Concurrency != 2 ||
			s.ProfileName != "default" || s.Threshold != 0.5 || s.MinimumLift != 0.5 {
			t.Errorf("%s: %v %+v", protocol, err, s)
		}
	}
}

func TestPrecedenceFlagEnvProjectDefault(t *testing.T) {
	project := model.Project{
		DefaultProfile: "ci",
		Profiles:       map[string]model.Profile{"ci": {Harness: "codex", Model: "from-project", Effort: "low", TimeoutSeconds: 90}},
		Protocols:      model.ProtocolSettings{Trigger: &model.TriggerSettings{Runs: 2, BatchSize: 5, Threshold: f(0.7)}},
		Budgets:        &model.Budgets{MaxModelRuns: 10, MaxCostUSD: 1},
	}
	s, err := settings.Resolve(project, settings.Flags{}, nil, model.Trigger)
	if err != nil || s.Harness != "codex" || s.Model != "from-project" || s.JudgeModel != "gpt-5.3-codex" ||
		s.Runs != 2 || s.BatchSize != 5 || s.Threshold != 0.7 || s.Effort != "low" || s.TimeoutSeconds != 90 ||
		s.Limits.MaxModelRuns != 10 || s.Limits.MaxCostUSD != 1 {
		t.Fatalf("%v %+v", err, s)
	}
	env := map[string]string{"CODEX_MODEL": "from-env", "CODEX_JUDGE_MODEL": "judge-env"}
	s, _ = settings.Resolve(project, settings.Flags{}, env, model.Trigger)
	if s.Model != "from-env" || s.JudgeModel != "judge-env" {
		t.Fatalf("%+v", s)
	}
	s, err = settings.Resolve(project, settings.Flags{
		Harness: "claude", Model: "flag", JudgeModel: "flag-judge", Runs: 1, Threshold: f(0.9), BatchSize: 4,
		Concurrency: 1, MaxCostUSD: 0.5, MaxModelRuns: 6, Split: "validation",
	}, env, model.Trigger)
	if err != nil || s.Harness != "claude" || s.Model != "flag" || s.JudgeModel != "flag-judge" || s.Runs != 1 ||
		s.Threshold != 0.9 || s.BatchSize != 4 || s.Concurrency != 1 || s.Limits.MaxCostUSD != 0.5 ||
		s.Limits.MaxModelRuns != 6 || s.Split != model.Validation {
		t.Fatalf("%v %+v", err, s)
	}
	viaEnv, _ := settings.Resolve(model.Project{}, settings.Flags{}, map[string]string{"SKILL_EVAL_HARNESS": "codex"}, model.Output)
	if viaEnv.Harness != "codex" {
		t.Fatal("the harness variable selects the harness")
	}
}

func TestProtocolSettingsOfTheOtherProtocols(t *testing.T) {
	p := model.Project{Protocols: model.ProtocolSettings{
		Routing:       &model.RoutingSettings{Runs: 2, Threshold: f(0.6)},
		Output:        &model.OutputSettings{Runs: 2, MinimumWinRate: f(0.8)},
		Instruction:   &model.InstructionSettings{Runs: 2, MinimumLift: f(0.25)},
		Orchestration: &model.OrchestrationSettings{Runs: 2},
	}}
	if s, _ := settings.Resolve(p, settings.Flags{}, nil, model.Routing); s.Runs != 2 || s.Threshold != 0.6 {
		t.Fatalf("%+v", s)
	}
	if s, _ := settings.Resolve(p, settings.Flags{}, nil, model.Output); s.Runs != 2 || s.MinimumWinRate != 0.8 {
		t.Fatalf("%+v", s)
	}
	if s, _ := settings.Resolve(p, settings.Flags{}, nil, model.Instruction); s.Runs != 2 || s.MinimumLift != 0.25 {
		t.Fatalf("%+v", s)
	}
	if s, _ := settings.Resolve(p, settings.Flags{}, nil, model.Orchestration); s.Runs != 2 {
		t.Fatalf("%+v", s)
	}
}

func TestRejections(t *testing.T) {
	for name, tc := range map[string]struct {
		project model.Project
		flags   settings.Flags
		want    string
	}{
		"unknown profile flag":    {model.Project{}, settings.Flags{Profile: "x"}, "not defined"},
		"unknown default profile": {model.Project{DefaultProfile: "x"}, settings.Flags{}, "not defined"},
		"unknown harness":         {model.Project{}, settings.Flags{Harness: "vim"}, "no eval support"},
		"too many runs":           {model.Project{}, settings.Flags{Runs: 4}, "runs must be"},
		"too much concurrency":    {model.Project{}, settings.Flags{Concurrency: 3}, "concurrency must be"},
		"bad threshold":           {model.Project{}, settings.Flags{Threshold: f(1.5)}, "threshold must be"},
		"bad split":               {model.Project{}, settings.Flags{Split: "dev"}, "split must be"},
	} {
		_, err := settings.Resolve(tc.project, tc.flags, nil, model.Trigger)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: %v", name, err)
		}
	}
}
