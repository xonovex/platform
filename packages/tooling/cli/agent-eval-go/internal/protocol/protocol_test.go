package protocol_test

import (
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
)

type stub struct{ name model.Protocol }

func (s stub) Name() model.Protocol                     { return s.name }
func (stub) Plan(protocol.Input) (protocol.Plan, error) { return protocol.Plan{}, nil }

func TestRegistry(t *testing.T) {
	r := protocol.NewRegistry()
	r.Register(stub{model.Output})
	r.Register(stub{model.Trigger})
	if names := r.Names(); len(names) != 2 || names[0] != model.Output {
		t.Fatalf("%v", names)
	}
	if p, err := r.Get(model.Trigger); err != nil || p.Name() != model.Trigger {
		t.Fatalf("%v", err)
	}
	if _, err := r.Get("x"); err == nil {
		t.Fatal("unknown protocol")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("want a panic")
		}
	}()
	r.Register(stub{model.Output})
}

func TestOffersGroupByPlugin(t *testing.T) {
	got := protocol.Offers(
		catalog.Skill{Name: "a", Plugin: "p1", Dir: "/p1/a"},
		catalog.Skill{Name: "b", Plugin: "p2", Dir: "/p2/b"},
		catalog.Skill{Name: "c", Plugin: "p1", Dir: "/p1/c"},
	)
	if len(got) != 2 || got[0].Name != "p1" || len(got[0].SkillDirs) != 2 || got[1].SkillDirs[0] != "/p2/b" {
		t.Fatalf("%+v", got)
	}
}

func TestHelpers(t *testing.T) {
	if s := protocol.Sum(model.Usage{CostUSD: 1}, model.Usage{CostUSD: 2, InputTokens: 3}); s.CostUSD != 3 || s.InputTokens != 3 {
		t.Fatalf("%+v", s)
	}
	if protocol.UsagePtr(model.Usage{}) != nil || protocol.UsagePtr(model.Usage{CostUSD: 1}) == nil {
		t.Fatal("empty usage is nil")
	}
	if protocol.RunSeconds(0, 60) != 60 || protocol.RunSeconds(30, 60) != 30 {
		t.Fatal("seconds")
	}
	lines := protocol.Failures([]model.CaseResult{
		{ID: "a", Outcome: model.Passed}, {ID: "b", Outcome: model.Failed},
		{ID: "c", Condition: model.Absent, Outcome: model.Errored, Error: "boom"},
	})
	if len(lines) != 2 || lines[0] != "failed: b" || lines[1] != "failed: c@absent (boom)" {
		t.Fatalf("%v", lines)
	}
}
