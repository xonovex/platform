package project_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/project"
)

type stub struct{ name string }

func (s stub) Name() string           { return s.name }
func (s stub) Summary() string        { return "" }
func (s stub) DefaultPaths() []string { return []string{"default"} }
func (s stub) Run(check.Context) ([]findings.Finding, error) {
	return nil, nil
}

var registered = []check.Check{stub{"b"}, stub{"a"}}

func TestDecode(t *testing.T) {
	got, err := project.Decode([]byte(`{"skip":["v"],"checks":{"a":{"enabled":false,"paths":["x"],"level":"warn","options":{"k":1}}}}`))
	if err != nil || got.Skip[0] != "v" || got.Checks["a"].Level != "warn" || *got.Checks["a"].Enabled || string(got.Checks["a"].Options) != `{"k":1}` {
		t.Fatalf("project = %+v, %v", got, err)
	}
	if _, err := project.Decode([]byte(`{"checks":[]}`)); err == nil || !strings.Contains(err.Error(), "validate section") {
		t.Fatalf("err = %v", err)
	}
}

func TestPlanDefaultsToEveryEnabledCheck(t *testing.T) {
	off := false
	steps, err := project.Plan(registered, project.Project{Checks: map[string]project.Settings{
		"a": {Enabled: &off},
		"b": {Paths: []string{"x"}, Level: "drift"},
	}}, nil, nil)
	if err != nil || len(steps) != 1 || steps[0].Check.Name() != "b" || steps[0].Paths[0] != "x" || steps[0].Cap != findings.Drift {
		t.Fatalf("steps = %+v, %v", steps, err)
	}
}

func TestPlanNamedChecksRunEvenWhenOff(t *testing.T) {
	off := false
	steps, err := project.Plan(registered, project.Project{Checks: map[string]project.Settings{"a": {Enabled: &off}}}, []string{"a"}, []string{"cli/path"})
	if err != nil || len(steps) != 1 || steps[0].Paths[0] != "cli/path" || steps[0].Cap != findings.Fail {
		t.Fatalf("steps = %+v, %v", steps, err)
	}
	plain, _ := project.Plan(registered, project.Project{}, []string{"b"}, nil)
	if plain[0].Paths[0] != "default" {
		t.Fatal("a check without settings keeps its default paths")
	}
}

func TestPlanErrors(t *testing.T) {
	if _, err := project.Plan(registered, project.Project{}, []string{"nope"}, nil); !errors.Is(err, project.ErrUnknownCheck) || !strings.Contains(err.Error(), "a, b") {
		t.Fatalf("err = %v", err)
	}
	if _, err := project.Plan(registered, project.Project{Checks: map[string]project.Settings{"nope": {}}}, nil, nil); !errors.Is(err, project.ErrUnknownCheck) {
		t.Fatalf("err = %v", err)
	}
	if _, err := project.Plan(registered, project.Project{Checks: map[string]project.Settings{"a": {Level: "loud"}}}, nil, nil); !errors.Is(err, findings.ErrUnknownLevel) {
		t.Fatalf("err = %v", err)
	}
}
