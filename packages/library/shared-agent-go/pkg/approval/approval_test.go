package approval

import (
	"strings"
	"testing"
)

const (
	digestA = "sha256:" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	digestB = "sha256:" + "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestCheck(t *testing.T) {
	if err := Check(digestA, digestA, "--approve-digest"); err != nil {
		t.Errorf("the matching digest failed: %v", err)
	}
	if err := Check(digestA, "", "--approve-digest"); err == nil || !strings.Contains(err.Error(), "--approve-digest "+digestA) {
		t.Errorf("the preview step must name the flag and the digest: %v", err)
	}
	if err := Check(digestA, "nonsense", "--approve-digest"); err == nil || !strings.Contains(err.Error(), "not a sha256") {
		t.Errorf("a malformed digest passed: %v", err)
	}
	if err := Check(digestA, digestB, "--approve-digest"); err == nil || !strings.Contains(err.Error(), "differs") {
		t.Errorf("a different digest passed: %v", err)
	}
}

func TestLedgerWithKeepsOrderAndReplacesTheSameApproval(t *testing.T) {
	ledger := EmptyLedger().
		With(Approved{Kind: "provider", Subject: "old", Digest: digestB}).
		With(Approved{Kind: "provider", Subject: "first", Digest: digestA}).
		With(Approved{Kind: "provider", Subject: "new", Digest: digestB}).
		With(Approved{Kind: "another", Subject: "x", Digest: digestB})
	if len(ledger.Approved) != 3 {
		t.Fatalf("Approved = %+v, want three records", ledger.Approved)
	}
	if got := ledger.Approved[0].Kind; got != "another" {
		t.Errorf("first kind = %q, want records sorted by kind", got)
	}
	if ledger.Approved[2].Subject != "new" {
		t.Errorf("the repeated approval kept subject %q, want new", ledger.Approved[2].Subject)
	}
	if !ledger.Has("provider", digestA) || ledger.Has("provider", "sha256:"+strings.Repeat("c", 64)) || ledger.Has("other", digestA) {
		t.Error("Has() must match the kind and the digest")
	}
}

func TestLedgerRoundTrips(t *testing.T) {
	ledger := EmptyLedger().With(Approved{Kind: "provider", Subject: "repo", Digest: digestA})
	data, err := EncodeLedger(ledger)
	if err != nil {
		t.Fatalf("EncodeLedger() error = %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Error("the ledger file must end with a line ending")
	}
	decoded, err := DecodeLedger(data)
	if err != nil || !decoded.Has("provider", digestA) {
		t.Fatalf("DecodeLedger() = %+v, %v", decoded, err)
	}
	empty, err := EncodeLedger(EmptyLedger())
	if err != nil || !strings.Contains(string(empty), `"approved": []`) {
		t.Errorf("an empty ledger = %q, %v, want an empty list", empty, err)
	}
}

func TestLedgerFailsClosed(t *testing.T) {
	bad := map[string]string{
		"version":   `{"version":2,"approved":[]}`,
		"unknown":   `{"version":1,"approved":[],"extra":1}`,
		"trailing":  `{"version":1,"approved":[]} {}`,
		"syntax":    `{`,
		"no kind":   `{"version":1,"approved":[{"kind":"","subject":"s","digest":"` + digestA + `"}]}`,
		"digest":    `{"version":1,"approved":[{"kind":"k","subject":"s","digest":"nope"}]}`,
		"duplicate": `{"version":1,"approved":[{"kind":"k","subject":"s","digest":"` + digestA + `"},{"kind":"k","subject":"t","digest":"` + digestA + `"}]}`,
	}
	for name, data := range bad {
		if _, err := DecodeLedger([]byte(data)); err == nil {
			t.Errorf("%s: DecodeLedger() error = nil", name)
		}
	}
	if _, err := EncodeLedger(Ledger{Version: 9}); err == nil {
		t.Error("EncodeLedger() accepted an unsupported version")
	}
}
