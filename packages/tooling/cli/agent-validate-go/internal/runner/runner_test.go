package runner_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/project"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/runner"
)

type fake struct {
	name string
	out  []findings.Finding
	err  error
	seen *check.Context
}

func (f fake) Name() string           { return f.name }
func (f fake) Summary() string        { return "" }
func (f fake) DefaultPaths() []string { return nil }
func (f fake) Run(ctx check.Context) ([]findings.Finding, error) {
	if f.seen != nil {
		*f.seen = ctx
	}
	return f.out, f.err
}

func TestRunCapsOrdersAndHandsOverTheStepSettings(t *testing.T) {
	var seen check.Context
	steps := []project.Step{{
		Check: fake{name: "a", seen: &seen, out: []findings.Finding{
			{Check: "a", Level: findings.Fail, Path: "z", Line: 1},
			{Check: "a", Level: findings.Warn, Path: "b", Line: 9},
			{Check: "a", Level: findings.Drift, Path: "b", Line: 2},
		}},
		Paths: []string{"p"}, Options: []byte(`{"x":1}`), Cap: findings.Drift,
	}}
	got, err := runner.Run(check.Context{}, steps)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Line != 2 || got[1].Line != 9 || got[2].Path != "z" {
		t.Fatalf("order = %+v", got)
	}
	if got[2].Level != findings.Drift || got[1].Level != findings.Warn {
		t.Fatalf("levels = %+v", got)
	}
	if seen.Paths[0] != "p" || string(seen.Options) != `{"x":1}` {
		t.Fatalf("context = %+v", seen)
	}
}

func TestRunStopsAtTheFirstFault(t *testing.T) {
	steps := []project.Step{
		{Check: fake{name: "a", err: errors.New("boom")}, Cap: findings.Fail},
		{Check: fake{name: "b"}, Cap: findings.Fail},
	}
	if _, err := runner.Run(check.Context{}, steps); err == nil || err.Error() != "a: boom" {
		t.Fatalf("err = %v", err)
	}
}
