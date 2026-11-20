package run_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/run"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/store"
)

func newStore(t *testing.T) store.Store {
	t.Helper()
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	fsys := memfs.New(nil)
	return store.Store{FS: fsys, Reader: document.Reader{FS: fsys, Validator: v}, Dir: "/r"}
}

var now = func() time.Time { return time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC) }

type counter struct {
	mu  sync.Mutex
	ran map[string]int
}

func unit(c *counter, id string, runs int, fail error) protocol.Unit {
	return protocol.Unit{ID: id, ModelRuns: runs, Run: func(context.Context) (model.CaseResult, error) {
		c.mu.Lock()
		c.ran[id]++
		c.mu.Unlock()
		if fail != nil {
			return model.CaseResult{}, fail
		}
		return model.CaseResult{ID: id, Outcome: model.Passed, Usage: &model.Usage{CostUSD: 0.01}}, nil
	}}
}

func decide(cases []model.CaseResult) protocol.Verdict {
	return protocol.Verdict{Pass: len(cases) == 3, Lines: []string{"n"}}
}

func options(s store.Store, g *budget.Guard) run.Options {
	return run.Options{Protocol: model.Trigger, Profile: "p", Concurrency: 2, Guard: g, Store: s, Now: now}
}

func plan(c *counter, fail error) protocol.Plan {
	return protocol.Plan{
		Subject: "skill", Inputs: []string{"doc"}, Decide: decide,
		Units: []protocol.Unit{unit(c, "a", 1, nil), unit(c, "b", 1, fail), unit(c, "c", 1, nil)},
	}
}

func TestRunsEveryUnitAndDecides(t *testing.T) {
	c := &counter{ran: map[string]int{}}
	rep, err := run.Execute(context.Background(), plan(c, nil), options(newStore(t), budget.New(budget.Limits{})))
	if err != nil || rep.Results.Status != model.Complete || !rep.Verdict.Pass || len(rep.Results.Cases) != 3 ||
		rep.Results.FinishedAt == "" || rep.Results.Usage == nil || rep.Results.Usage.CostUSD < 0.029 {
		t.Fatalf("%v %+v", err, rep)
	}
}

func TestDryRunPlansWithoutRunning(t *testing.T) {
	c := &counter{ran: map[string]int{}}
	o := options(newStore(t), budget.New(budget.Limits{}))
	o.DryRun = true
	rep, err := run.Execute(context.Background(), plan(c, nil), o)
	if err != nil || !rep.Dry || rep.Planned != 3 || rep.Pending != 3 || rep.ModelRuns != 3 || len(c.ran) != 0 {
		t.Fatalf("%v %+v", err, rep)
	}
}

func TestBudgetStopsTheRunAndResumeFinishesIt(t *testing.T) {
	s := newStore(t)
	c := &counter{ran: map[string]int{}}
	rep, err := run.Execute(context.Background(), plan(c, nil), options(s, budget.New(budget.Limits{MaxModelRuns: 2})))
	if err != nil || rep.Results.Status != model.Incomplete || rep.Results.Stop != model.StopBudget || len(rep.Results.Cases) != 2 || rep.Pending != 1 {
		t.Fatalf("%v %+v", err, rep)
	}
	o := options(s, budget.New(budget.Limits{}))
	o.Resume = true
	rep, err = run.Execute(context.Background(), plan(c, nil), o)
	if err != nil || rep.Results.Status != model.Complete || rep.Resumed != 2 || len(rep.Results.Cases) != 3 || c.ran["a"] != 1 || c.ran["c"] != 1 {
		t.Fatalf("%v %+v %v", err, rep, c.ran)
	}
	again, err := run.Execute(context.Background(), plan(c, nil), o)
	if err != nil || !again.Verdict.Pass || c.ran["a"] != 1 {
		t.Fatalf("a complete run is returned, not rerun: %v %+v", err, again)
	}
}

func TestResumeIgnoresARunWithOtherInputs(t *testing.T) {
	s := newStore(t)
	c := &counter{ran: map[string]int{}}
	_, _ = run.Execute(context.Background(), plan(c, nil), options(s, budget.New(budget.Limits{MaxModelRuns: 1})))
	changed := plan(c, nil)
	changed.Inputs = []string{"edited"}
	o := options(s, budget.New(budget.Limits{}))
	o.Resume = true
	rep, err := run.Execute(context.Background(), changed, o)
	if err != nil || rep.Resumed != 0 || c.ran["a"] != 2 {
		t.Fatalf("%v %+v %v", err, rep, c.ran)
	}
}

func TestBatchSizeLimitsUnitsPerInvocation(t *testing.T) {
	c := &counter{ran: map[string]int{}}
	o := options(newStore(t), budget.New(budget.Limits{}))
	o.BatchSize = 2
	rep, err := run.Execute(context.Background(), plan(c, nil), o)
	if err != nil || rep.Results.Status != model.Incomplete || rep.Results.Stop != model.StopBudget || len(rep.Results.Cases) != 2 || rep.Pending != 1 {
		t.Fatalf("%v %+v", err, rep)
	}
}

func TestUnitFaultInvalidatesTheRun(t *testing.T) {
	c := &counter{ran: map[string]int{}}
	s := newStore(t)
	rep, err := run.Execute(context.Background(), plan(c, errors.New("claude down")), options(s, budget.New(budget.Limits{})))
	if err != nil || rep.Results.Status != model.Invalid || rep.Results.Stop != model.StopInfrastructure {
		t.Fatalf("%v %+v", err, rep)
	}
	var errored model.CaseResult
	for _, cr := range rep.Results.Cases {
		if cr.Outcome == model.Errored {
			errored = cr
		}
	}
	if errored.ID != "b" || errored.Error != "b: claude down" || c.ran["c"] != 0 {
		t.Fatalf("%+v %v", errored, c.ran)
	}
	o := options(s, budget.New(budget.Limits{}))
	o.Resume = true
	again, err := run.Execute(context.Background(), plan(c, nil), o)
	if err != nil || again.Resumed != 0 || again.Results.Status != model.Complete {
		t.Fatalf("an invalid run is never resumed: %v %+v", err, again)
	}
}

func TestCancelledContextIsAnError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &counter{ran: map[string]int{}}
	p := plan(c, nil)
	p.Units[0].Run = func(context.Context) (model.CaseResult, error) { cancel(); return model.CaseResult{}, context.Canceled }
	if _, err := run.Execute(ctx, p, options(newStore(t), budget.New(budget.Limits{}))); !errors.Is(err, context.Canceled) {
		t.Fatalf("%v", err)
	}
}

func TestStoreFaults(t *testing.T) {
	s := newStore(t)
	c := &counter{ran: map[string]int{}}
	_ = s.FS.WriteFile(s.File(store.RunID(model.Trigger, "skill", store.Digest("trigger", "p", "skill", "doc"))), []byte("{}"))
	o := options(s, budget.New(budget.Limits{}))
	o.Resume = true
	if _, err := run.Execute(context.Background(), plan(c, nil), o); err == nil {
		t.Fatal("an unreadable checkpoint is an error")
	}
	o.Resume = false
	o.Store.Dir = ""
	o.Store.FS = failingFS{}
	if _, err := run.Execute(context.Background(), plan(c, nil), o); err == nil {
		t.Fatal("a failed save is an error")
	}
}

func TestZeroConcurrencyRunsOneAtATime(t *testing.T) {
	c := &counter{ran: map[string]int{}}
	o := options(newStore(t), budget.New(budget.Limits{}))
	o.Concurrency = 0
	rep, err := run.Execute(context.Background(), plan(c, nil), o)
	if err != nil || rep.Results.Status != model.Complete {
		t.Fatalf("%v %+v", err, rep)
	}
}
