package predicate_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
)

type stub struct {
	kind string
	pass func(model.Predicate) bool
	err  error
	n    int
}

func (s *stub) Kind() string { return s.kind }
func (s *stub) Evaluate(_ context.Context, ps []model.Predicate, _ predicate.Subject) ([]predicate.Verdict, model.Usage, error) {
	if s.err != nil {
		return nil, model.Usage{}, s.err
	}
	out := make([]predicate.Verdict, len(ps)+s.n)
	for i, p := range ps {
		out[i] = predicate.Verdict{Passed: s.pass(p)}
	}
	return out, model.Usage{CostUSD: 0.5}, nil
}

func TestEvaluateKeepsOrderAcrossKindsAndSumsUsage(t *testing.T) {
	r := predicate.NewRegistry()
	r.Register(&stub{kind: "a", pass: func(p model.Predicate) bool { return p.Path == "ok" }})
	r.Register(&stub{kind: "b", pass: func(model.Predicate) bool { return true }})
	ps := []model.Predicate{{Kind: "a", Path: "ok"}, {Kind: "b"}, {Kind: "a", Path: "no"}}
	got, usage, err := r.Evaluate(context.Background(), ps, predicate.Subject{})
	if err != nil || !got[0].Passed || !got[1].Passed || got[2].Passed || usage.CostUSD != 1 {
		t.Fatalf("%v %+v %+v", err, got, usage)
	}
}

func TestEvaluateErrors(t *testing.T) {
	r := predicate.NewRegistry()
	r.Register(&stub{kind: "bad", err: errors.New("down")})
	r.Register(&stub{kind: "long", n: 1, pass: func(model.Predicate) bool { return true }})
	for _, kind := range []string{"missing", "bad", "long"} {
		if _, _, err := r.Evaluate(context.Background(), []model.Predicate{{Kind: kind}}, predicate.Subject{}); err == nil {
			t.Errorf("%s: want an error", kind)
		}
	}
}

func TestRegisterTwicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want a panic")
		}
	}()
	r := predicate.NewRegistry()
	r.Register(&stub{kind: "a"})
	r.Register(&stub{kind: "a"})
}

func ratio(v float64) *float64 { return &v }

func TestAggregateAppliesThresholds(t *testing.T) {
	ps := []model.Predicate{{Kind: "x"}, {Kind: "x", Threshold: ratio(0.5)}, {Kind: "x"}}
	runs := [][]predicate.Verdict{
		{{Passed: true}, {Passed: true}, {Passed: false, Detail: "d"}},
		{{Passed: false, Detail: "first"}, {Passed: false}, {Passed: false}},
	}
	got := predicate.Aggregate(ps, runs)
	if got[0].Passed || !strings.HasPrefix(got[0].Detail, "1/2 runs passed: first") {
		t.Fatalf("%+v", got[0])
	}
	if !got[1].Passed || got[1].Detail != "1/2 runs passed" {
		t.Fatalf("%+v", got[1])
	}
	if got[2].Passed || got[2].Detail != "0/2 runs passed: d" {
		t.Fatalf("%+v", got[2])
	}
	single := predicate.Aggregate(ps[:1], [][]predicate.Verdict{{{Passed: false}}})
	if single[0].Detail != "no run passed" {
		t.Fatalf("%+v", single)
	}
	if none := predicate.Aggregate(ps[:1], nil); none[0].Passed {
		t.Fatal("no run, no pass")
	}
}

func TestScore(t *testing.T) {
	if predicate.Score(nil) != 1 || !predicate.AllPassed(nil) {
		t.Fatal("empty passes")
	}
	rs := []model.PredicateResult{{Passed: true}, {Passed: false}}
	if predicate.Score(rs) != 0.5 || predicate.AllPassed(rs) {
		t.Fatal("half")
	}
}
