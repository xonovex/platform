package store_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/store"
)

func newStore(t *testing.T) store.Store {
	t.Helper()
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	fsys := memfs.New(nil)
	return store.Store{FS: fsys, Reader: document.Reader{FS: fsys, Validator: v}, Dir: "/results"}
}

func TestDigestIsStableAndSplitSensitive(t *testing.T) {
	first, second := store.Digest("a", "b"), store.Digest("a", "b")
	if first != second {
		t.Fatal("unstable")
	}
	if !strings.HasPrefix(store.Digest("a"), "sha256:") {
		t.Fatal("prefix")
	}
	if store.Digest("ab", "c") == store.Digest("a", "bc") {
		t.Fatal("split changes the digest")
	}
}

func TestRunID(t *testing.T) {
	id := store.RunID(model.Trigger, "my skill/x", "sha256:0123456789abcdef")
	if id != "trigger-my-skill-x-01234567" {
		t.Fatal(id)
	}
	if store.RunID(model.Output, "///", "ab") != "output-run-ab" {
		t.Fatal("an empty subject gets a placeholder")
	}
}

func TestSaveThenLoad(t *testing.T) {
	s := newStore(t)
	if _, found, err := s.Load("none"); err != nil || found {
		t.Fatalf("%v %v", found, err)
	}
	r := model.Results{
		Version: 1, RunID: "trigger-x-aaaaaaaa", Protocol: model.Trigger, Profile: "default",
		InputsDigest: store.Digest("d"), StartedAt: "2026-11-20T10:00:00Z", Status: model.Incomplete,
		Cases: []model.CaseResult{{ID: "q1", Outcome: model.Passed}},
	}
	if err := s.Save(r); err != nil {
		t.Fatal(err)
	}
	got, found, err := s.Load(r.RunID)
	if err != nil || !found || got.RunID != r.RunID || len(got.Cases) != 1 {
		t.Fatalf("%v %v %+v", found, err, got)
	}
	if !strings.HasSuffix(s.File(r.RunID), "/trigger-x-aaaaaaaa/results.json") {
		t.Fatal(s.File(r.RunID))
	}
	if err := s.Save(model.Results{}); err == nil {
		t.Fatal("a run id is required")
	}
}

func TestLoadRejectsAnInvalidCheckpoint(t *testing.T) {
	s := newStore(t)
	_ = s.FS.WriteFile(s.File("bad"), []byte(`{"version":1}`))
	if _, _, err := s.Load("bad"); err == nil {
		t.Fatal("want an error")
	}
}
