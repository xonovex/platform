package baseline_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/baseline"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

func score(v float64) *float64 { return &v }

func results(cases ...model.CaseResult) model.Results {
	return model.Results{Protocol: model.Output, Profile: "p", InputsDigest: "sha256:0", Cases: cases}
}

func TestRecordUsesScoreThenOutcome(t *testing.T) {
	b := baseline.Record(results(
		model.CaseResult{ID: "a", Score: score(0.5), Runs: 3},
		model.CaseResult{ID: "b", Outcome: model.Passed},
		model.CaseResult{ID: "c", Outcome: model.Failed, Condition: model.Absent},
	), 0.1, "2026-11-20T10:00:00Z")
	if len(b.Cases) != 3 || b.Cases[0].PassRate != 0.5 || b.Cases[0].Runs != 3 ||
		b.Cases[1].PassRate != 1 || b.Cases[1].Runs != 1 || b.Cases[2].PassRate != 0 || b.Cases[2].Condition != model.Absent {
		t.Fatalf("%+v", b)
	}
	if b.Tolerance != 0.1 || b.Version != model.Version {
		t.Fatalf("%+v", b)
	}
}

func TestCompare(t *testing.T) {
	base := model.Baseline{Tolerance: 0.1, Cases: []model.BaselineCase{
		{ID: "same", PassRate: 1}, {ID: "worse", PassRate: 1}, {ID: "better", PassRate: 0.5},
		{ID: "gone", PassRate: 1}, {ID: "wobble", PassRate: 1, Condition: model.Present},
	}}
	got := baseline.Compare(base, results(
		model.CaseResult{ID: "same", Outcome: model.Passed},
		model.CaseResult{ID: "worse", Outcome: model.Failed},
		model.CaseResult{ID: "better", Outcome: model.Passed},
		model.CaseResult{ID: "wobble", Condition: model.Present, Score: score(0.95)},
		model.CaseResult{ID: "new", Outcome: model.Passed},
	))
	if len(got.Regressions) != 1 || got.Regressions[0].Key != "worse" || len(got.Improvements) != 1 ||
		got.Improvements[0].Key != "better" || len(got.Missing) != 1 || got.Missing[0] != "gone" ||
		len(got.Added) != 1 || got.Added[0] != "new" || !got.Regressed() {
		t.Fatalf("%+v", got)
	}
	clean := baseline.Compare(model.Baseline{Cases: []model.BaselineCase{{ID: "a", PassRate: 1}}}, results(model.CaseResult{ID: "a", Outcome: model.Passed}))
	if clean.Regressed() {
		t.Fatal("no change, no regression")
	}
}

func TestStoreRoundTrip(t *testing.T) {
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	fsys := memfs.New(nil)
	s := baseline.Store{FS: fsys, Reader: document.Reader{FS: fsys, Validator: v}}
	if _, found, err := s.Load("/b.json"); found || err != nil {
		t.Fatalf("%v %v", found, err)
	}
	b := baseline.Record(model.Results{
		Protocol: model.Trigger, Profile: "p",
		InputsDigest: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		Cases:        []model.CaseResult{{ID: "q", Outcome: model.Passed}},
	}, 0.1, "2026-11-20T10:00:00Z")
	if err := s.Save("/b.json", b); err != nil {
		t.Fatal(err)
	}
	got, found, err := s.Load("/b.json")
	if err != nil || !found || len(got.Cases) != 1 {
		t.Fatalf("%v %v %+v", found, err, got)
	}
	_ = fsys.WriteFile("/bad.json", []byte("{}"))
	if _, _, err := s.Load("/bad.json"); err == nil {
		t.Fatal("an invalid baseline is an error")
	}
}
