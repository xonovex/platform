package protocol_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/judge"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/session"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/settings"
)

type exec struct {
	err  error
	runs int
}

func (e *exec) Execute(context.Context, harness.Request) (harness.Observation, error) {
	e.runs++
	if e.err != nil {
		return harness.Observation{}, e.err
	}
	return harness.Observation{
		Events: []harness.Event{{Kind: harness.EventCommand, Text: "ls"}},
		Usage:  model.Usage{CostUSD: 0.01}, DurationMs: 5,
	}, nil
}

func env(e harness.Executor) protocol.Env {
	r := predicate.NewRegistry()
	r.Register(session.New())
	return protocol.Env{Exec: e, Predicates: r}
}

func TestTriggerRequestUsesTheTriggerCaps(t *testing.T) {
	req := protocol.TriggerRequest(settings.Settings{Harness: "claude", Model: "m", TimeoutSeconds: 30}, "q", "skill", nil)
	if req.Role != harness.RoleTrigger || req.Prompt != "q" || req.Target != "skill" || req.Caps.Timeout.Seconds() != 30 ||
		req.Caps.BudgetUSD != settings.TriggerRunBudgetUSD || req.Caps.MaxOutputChars != 2000 {
		t.Fatalf("%+v", req)
	}
	if protocol.TriggerRequest(settings.Settings{}, "q", "s", nil).Caps.Timeout.Seconds() != settings.TriggerTimeoutSeconds {
		t.Fatal("default timeout")
	}
}

func TestRunTrialAggregatesRuns(t *testing.T) {
	e := &exec{}
	ps := []model.Predicate{{Kind: model.KindSession, Event: "command", Match: "ls"}}
	res, err := protocol.RunTrial(context.Background(), env(e), protocol.Trial{Runs: 2, Predicates: ps, Judge: judge.Spec{}})
	if err != nil || e.runs != 2 || len(res.Observations) != 2 || res.Usage.CostUSD != 0.02 || res.DurationMs != 10 ||
		len(res.Predicates) != 1 || !res.Predicates[0].Passed || protocol.OutcomeOf(res.Predicates) != model.Passed {
		t.Fatalf("%v %+v", err, res)
	}
	miss := []model.Predicate{{Kind: model.KindSession, Event: "command", Match: "rm"}}
	res, _ = protocol.RunTrial(context.Background(), env(&exec{}), protocol.Trial{Runs: 1, Predicates: miss})
	if protocol.OutcomeOf(res.Predicates) != model.Failed {
		t.Fatal("a failed predicate fails the trial")
	}
}

func TestRunTrialFaults(t *testing.T) {
	if _, err := protocol.RunTrial(context.Background(), env(&exec{err: errors.New("down")}), protocol.Trial{Runs: 1}); err == nil {
		t.Fatal("executor fault")
	}
	bad := []model.Predicate{{Kind: "nope"}}
	if _, err := protocol.RunTrial(context.Background(), env(&exec{}), protocol.Trial{Runs: 1, Predicates: bad}); err == nil {
		t.Fatal("predicate fault")
	}
}

func TestHasKind(t *testing.T) {
	ps := []model.Predicate{{Kind: "a"}, {Kind: "b"}}
	if !protocol.HasKind(ps, "b") || protocol.HasKind(ps, "c") {
		t.Fatal("has kind")
	}
}
