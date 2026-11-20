package cost_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/cost"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

func entry(id string, p model.Protocol, runs int, usd float64) model.CostEntry {
	return model.CostEntry{
		RunID: id, RecordedAt: "2026-11-20T10:00:00Z", Protocol: p, Profile: "p", Model: "m",
		ModelRuns: runs, Usage: model.Usage{CostUSD: usd},
	}
}

func TestAppendReplacesTheSameRun(t *testing.T) {
	h := cost.Append(model.CostHistory{}, entry("a", model.Trigger, 3, 0.1))
	h = cost.Append(h, entry("b", model.Output, 4, 0.2))
	h = cost.Append(h, entry("a", model.Trigger, 6, 0.3))
	if len(h.Entries) != 2 || h.Entries[1].ModelRuns != 6 || h.Version != model.Version {
		t.Fatalf("%+v", h)
	}
}

func TestSummarize(t *testing.T) {
	h := cost.Append(cost.Append(cost.Append(model.CostHistory{},
		entry("a", model.Trigger, 3, 0.1)), entry("b", model.Output, 4, 0.2)), entry("c", model.Trigger, 1, 0.05))
	s := cost.Summarize(h, "")
	if s.All.Runs != 3 || s.All.ModelRuns != 8 || s.ByProtocol[model.Trigger].Runs != 2 {
		t.Fatalf("%+v", s)
	}
	if ps := s.Protocols(); len(ps) != 2 || ps[0] != model.Output {
		t.Fatalf("%v", ps)
	}
	only := cost.Summarize(h, model.Output)
	if only.All.Runs != 1 || len(only.ByProtocol) != 1 {
		t.Fatalf("%+v", only)
	}
}

func TestStore(t *testing.T) {
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	fsys := memfs.New(nil)
	s := cost.Store{FS: fsys, Reader: document.Reader{FS: fsys, Validator: v}}
	empty, err := s.Load("/c.json")
	if err != nil || len(empty.Entries) != 0 || empty.Version != model.Version {
		t.Fatalf("%v %+v", err, empty)
	}
	h := cost.Append(empty, entry("a", model.Trigger, 3, 0.1))
	if err := s.Save("/c.json", h); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load("/c.json")
	if err != nil || len(got.Entries) != 1 {
		t.Fatalf("%v %+v", err, got)
	}
}
