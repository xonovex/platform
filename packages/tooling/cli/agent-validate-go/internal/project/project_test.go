package project_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
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

func TestLoad(t *testing.T) {
	ctx := checktest.Context(t, map[string]string{
		"ok/agentic.validate.json":      `{"version":1,"checks":{"a":{"enabled":false,"paths":["x"],"level":"warn","options":{"k":1}}}}`,
		"broken/agentic.validate.json":  `{"version":2,"checks":{"a":{"level":"pass"}}}`,
		"garbage/agentic.validate.json": `{`,
	})
	got, err := project.Load(ctx.Tree, "ok/agentic.validate.json", ctx.Validator, true)
	if err != nil || got.Checks["a"].Level != "warn" || *got.Checks["a"].Enabled || string(got.Checks["a"].Options) != `{"k":1}` {
		t.Fatalf("project = %+v, %v", got, err)
	}
	if _, err := project.Load(ctx.Tree, "missing.json", ctx.Validator, false); err != nil {
		t.Fatalf("an optional file may be missing: %v", err)
	}
	if _, err := project.Load(ctx.Tree, "missing.json", ctx.Validator, true); err == nil {
		t.Fatal("a required file must exist")
	}
	if _, err := project.Load(ctx.Tree, "broken/agentic.validate.json", ctx.Validator, false); err == nil || !strings.Contains(err.Error(), "/version") {
		t.Fatalf("err = %v", err)
	}
	if _, err := project.Load(ctx.Tree, "garbage/agentic.validate.json", ctx.Validator, false); err == nil {
		t.Fatal("malformed JSON is an error")
	}
	if _, err := project.Load(ctx.Tree, "ok", ctx.Validator, false); err == nil {
		t.Fatal("a directory cannot be read as a file")
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
