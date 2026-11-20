package trust_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/configtest"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

const ledgerPath = "/state/agentic/trust.json"

func stateDir() (string, error) { return "/state/agentic", nil }

func project(t *testing.T) layer.Document {
	t.Helper()
	d, err := layer.Parse(spec.Project, "/repo/agentic.json", []byte(`{"version":1,"eval":{}}`+"\n"))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func ledger(files map[string]string) (trust.Ledger, *configtest.Files) {
	store := configtest.NewFiles(files)
	return trust.Ledger{Store: store, StateDir: stateDir}, store
}

func TestAGateWithoutSectionsPasses(t *testing.T) {
	l, _ := ledger(nil)
	if err := trust.Gate(trust.Request{Document: project(t)}, l, &configtest.Prompter{}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
}

func TestAnApprovedDigestPasses(t *testing.T) {
	doc := project(t)
	l, _ := ledger(map[string]string{ledgerPath: `{"version":1,"approvals":[{"digest":"` + doc.Digest + `","path":"/x"}]}`})
	if err := trust.Gate(trust.Request{Document: doc, Sections: []string{"eval"}}, l, &configtest.Prompter{}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
}

func TestWithoutATerminalTheGateFailsClosedAndNamesTheDigest(t *testing.T) {
	doc := project(t)
	l, _ := ledger(nil)
	err := trust.Gate(trust.Request{Document: doc, Sections: []string{"eval"}}, l, &configtest.Prompter{}, &strings.Builder{})
	if !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), "--approve-config-digest "+doc.Digest) {
		t.Fatalf("err = %v", err)
	}
}

func TestTheFlagApprovesAndIsKept(t *testing.T) {
	doc := project(t)
	l, _ := ledger(nil)
	if err := trust.Gate(trust.Request{Document: doc, Sections: []string{"eval"}, Flag: doc.Digest}, l, &configtest.Prompter{}, &strings.Builder{}); err != nil {
		t.Fatal(err)
	}
	if ok, err := l.Approved(doc.Digest); !ok || err != nil {
		t.Fatalf("Approved = %v, %v", ok, err)
	}
	if err := trust.Gate(trust.Request{Document: doc, Sections: []string{"eval"}}, l, &configtest.Prompter{}, &strings.Builder{}); err != nil {
		t.Fatalf("a kept approval must pass: %v", err)
	}
}

func TestAFlagForAnotherDigestFails(t *testing.T) {
	l, _ := ledger(nil)
	err := trust.Gate(trust.Request{Document: project(t), Sections: []string{"eval"}, Flag: "sha256:00"}, l, &configtest.Prompter{}, &strings.Builder{})
	if !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), "differs from the digest") {
		t.Fatalf("err = %v", err)
	}
}

func TestTheTerminalShowsTheFileAndAsks(t *testing.T) {
	doc := project(t)
	for answer, approve := range map[string]bool{"y": true, " YES ": true, "n": false, "": false} {
		l, _ := ledger(nil)
		var out strings.Builder
		err := trust.Gate(trust.Request{Document: doc, Sections: []string{"eval", "run"}}, l, &configtest.Prompter{Terminal: true, Answers: []string{answer}}, &out)
		shown := out.String()
		if !strings.Contains(shown, "/repo/agentic.json sets the eval, run configuration") || !strings.Contains(shown, `{"version":1,"eval":{}}`) ||
			!strings.Contains(shown, "Digest: "+doc.Digest) || !strings.Contains(shown, "Approve this digest? [y/N] ") {
			t.Fatalf("shown = %q", shown)
		}
		if approve != (err == nil) {
			t.Fatalf("answer %q: err = %v", answer, err)
		}
		if ok, _ := l.Approved(doc.Digest); ok != approve {
			t.Fatalf("answer %q: kept = %v", answer, ok)
		}
	}
}

func TestAClosedInputDeclinesAndNamesTheFlag(t *testing.T) {
	l, _ := ledger(nil)
	doc := project(t)
	err := trust.Gate(trust.Request{Document: doc, Sections: []string{"eval"}}, l, &configtest.Prompter{Terminal: true}, &strings.Builder{})
	if !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), "--approve-config-digest "+doc.Digest) {
		t.Fatalf("err = %v", err)
	}
}

var errBoom = errors.New("boom")

type brokenPrompter struct{}

func (brokenPrompter) IsTerminal() bool          { return true }
func (brokenPrompter) ReadLine() (string, error) { return "", errBoom }

type brokenWriter struct{ after int }

func (w *brokenWriter) Write(p []byte) (int, error) {
	if w.after == 0 {
		return 0, errBoom
	}
	w.after--
	return len(p), nil
}

func TestTerminalFaultsSurface(t *testing.T) {
	request := trust.Request{Document: project(t), Sections: []string{"eval"}}
	l, _ := ledger(nil)
	if err := trust.Gate(request, l, brokenPrompter{}, &strings.Builder{}); !errors.Is(err, errBoom) {
		t.Fatalf("read fault = %v", err)
	}
	for after := range 2 {
		if err := trust.Gate(request, l, &configtest.Prompter{Terminal: true}, &brokenWriter{after: after}); !errors.Is(err, errBoom) {
			t.Fatalf("write fault after %d writes = %v", after, err)
		}
	}
}

func TestLedgerFaultsSurface(t *testing.T) {
	doc := project(t)
	request := trust.Request{Document: doc, Sections: []string{"eval"}, Flag: doc.Digest}
	cases := map[string]string{
		"not json":       `{`,
		"unknown member": `{"version":1,"approvals":[],"extra":1}`,
		"wrong version":  `{"version":2,"approvals":[]}`,
		"bad digest":     `{"version":1,"approvals":[{"digest":"md5:1","path":"/x"}]}`,
	}
	for name, content := range cases {
		l, _ := ledger(map[string]string{ledgerPath: content})
		if err := trust.Gate(request, l, &configtest.Prompter{}, &strings.Builder{}); err == nil || !strings.Contains(err.Error(), "approval ledger") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if err := (trust.Ledger{Store: failingStore{}, StateDir: stateDir}).Approve(trust.Approval{Digest: doc.Digest}); !errors.Is(err, errBoom) {
		t.Fatalf("read fault = %v", err)
	}
	if err := (trust.Ledger{Store: failingStore{read: true}, StateDir: stateDir}).Approve(trust.Approval{Digest: doc.Digest}); !errors.Is(err, errBoom) {
		t.Fatalf("write fault = %v", err)
	}
	l, _ := ledger(nil)
	if err := l.Approve(trust.Approval{Digest: "sha256:1"}); err == nil {
		t.Fatal("a malformed digest must not be approved")
	}
	homeless := trust.Ledger{Store: configtest.NewFiles(nil), StateDir: func() (string, error) { return "", errBoom }}
	if _, err := homeless.Approved(doc.Digest); !errors.Is(err, errBoom) {
		t.Fatalf("approved without a state directory = %v", err)
	}
	if err := homeless.Approve(trust.Approval{Digest: doc.Digest}); !errors.Is(err, errBoom) {
		t.Fatalf("approve without a state directory = %v", err)
	}
}

func TestApproveKeepsOneSortedRecordPerDigest(t *testing.T) {
	a, b := "sha256:"+strings.Repeat("a", 64), "sha256:"+strings.Repeat("b", 64)
	l, store := ledger(nil)
	for _, digest := range []string{b, a, b} {
		if err := l.Approve(trust.Approval{Digest: digest, Path: "/p"}); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := store.ReadFile(ledgerPath)
	want := `{
  "version": 1,
  "approvals": [
    {
      "digest": "` + a + `",
      "path": "/p"
    },
    {
      "digest": "` + b + `",
      "path": "/p"
    }
  ]
}
`
	if string(data) != want {
		t.Fatalf("ledger = %s", data)
	}
}

// failingStore fails every write, and every read unless read is set; held is
// the digest a successful read holds.
type failingStore struct {
	read bool
	held string
}

func (s failingStore) ReadFile(string) ([]byte, error) {
	if s.read && s.held != "" {
		return []byte(`{"version":1,"approvals":[{"digest":"` + s.held + `","path":"/p"}]}`), nil
	}
	if s.read {
		return []byte(`{"version":1,"approvals":[]}`), nil
	}
	return nil, errBoom
}

func (failingStore) WriteFile(string, []byte) error { return errBoom }

func TestDecideShowsTheReviewOfAnySubject(t *testing.T) {
	digest := "sha256:" + strings.Repeat("c", 64)
	subject := trust.Subject{Digest: digest, Path: "plugin hello", Notice: "plugin hello is installed only after you approve its digest", Review: "files: 3\n"}
	l, _ := ledger(nil)
	var out strings.Builder
	if err := trust.Decide(subject, "", l, &configtest.Prompter{Terminal: true, Answers: []string{"y"}}, &out); err != nil {
		t.Fatal(err)
	}
	if want := "plugin hello is installed only after you approve its digest.\n\nfiles: 3\n\nDigest: " + digest + "\nApprove this digest? [y/N] "; out.String() != want {
		t.Fatalf("shown = %q", out.String())
	}
	l, _ = ledger(nil)
	err := trust.Decide(subject, "", l, &configtest.Prompter{}, &strings.Builder{})
	if !errors.Is(err, trust.ErrNotApproved) || !strings.Contains(err.Error(), "not approved: plugin hello is installed only after you approve its digest: review it") {
		t.Fatalf("err = %v", err)
	}
}

func TestRevokeRemovesOnlyTheDigest(t *testing.T) {
	a, b := "sha256:"+strings.Repeat("a", 64), "sha256:"+strings.Repeat("b", 64)
	l, _ := ledger(nil)
	if removed, err := l.Revoke(a); removed || err != nil {
		t.Fatalf("revoke from an empty ledger = %v, %v", removed, err)
	}
	for _, digest := range []string{a, b} {
		if err := l.Approve(trust.Approval{Digest: digest, Path: "/p"}); err != nil {
			t.Fatal(err)
		}
	}
	if removed, err := l.Revoke(a); !removed || err != nil {
		t.Fatalf("revoke = %v, %v", removed, err)
	}
	if ok, _ := l.Approved(a); ok {
		t.Fatal("the revoked digest is still approved")
	}
	if ok, _ := l.Approved(b); !ok {
		t.Fatal("revoke removed another digest")
	}
	homeless := trust.Ledger{Store: configtest.NewFiles(nil), StateDir: func() (string, error) { return "", errBoom }}
	if _, err := homeless.Revoke(a); !errors.Is(err, errBoom) {
		t.Fatalf("revoke without a state directory = %v", err)
	}
	if _, err := (trust.Ledger{Store: failingStore{}, StateDir: stateDir}).Revoke(a); !errors.Is(err, errBoom) {
		t.Fatalf("read fault = %v", err)
	}
	broken := trust.Ledger{Store: failingStore{read: true, held: a}, StateDir: stateDir}
	if _, err := broken.Revoke(a); !errors.Is(err, errBoom) {
		t.Fatalf("write fault = %v", err)
	}
}
