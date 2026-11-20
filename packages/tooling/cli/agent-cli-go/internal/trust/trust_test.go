package trust

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/approval"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

const digest = "sha256:" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func terminal(answers ...string) *systemtest.Fake {
	host := systemtest.New()
	host.Terminal = true
	host.Input = answers
	return host
}

func TestDecideAsksOnATerminalAndApprovesOnYes(t *testing.T) {
	for _, answer := range []string{"y", "Y", "yes", " YES "} {
		var out bytes.Buffer
		approved, err := Decide(terminal(answer), &out, digest, "")
		if err != nil || approved != digest {
			t.Errorf("answer %q: Decide() = %q, %v, want the digest", answer, approved, err)
		}
		if !strings.Contains(out.String(), "Approve this digest? [y/N] ") {
			t.Errorf("answer %q: output = %q", answer, out.String())
		}
	}
}

func TestDecideDeclinesEveryOtherAnswer(t *testing.T) {
	for _, answer := range []string{"", "n", "no", "yess", "ok"} {
		if _, err := Decide(terminal(answer), io.Discard, digest, ""); !errors.Is(err, ErrDeclined) {
			t.Errorf("answer %q: error = %v, want ErrDeclined", answer, err)
		}
	}
	var out bytes.Buffer
	if _, err := Decide(terminal(), &out, digest, ""); !errors.Is(err, ErrDeclined) || !strings.HasSuffix(out.String(), "[y/N] \n") {
		t.Errorf("end of input: error = %v, output = %q", err, out.String())
	}
}

func TestDecideStopsWhenTheAnswerCannotBeRead(t *testing.T) {
	host := terminal()
	host.InputErr = errors.New("interrupted")
	if _, err := Decide(host, io.Discard, digest, ""); err == nil || errors.Is(err, ErrDeclined) || !strings.Contains(err.Error(), "interrupted") {
		t.Errorf("error = %v, want the read failure", err)
	}
}

func TestDecideStopsWhenTheQuestionCannotBeWritten(t *testing.T) {
	if _, err := Decide(terminal("y"), failingWriter{}, digest, ""); err == nil {
		t.Error("Decide() ignored the write failure")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestDecidePassesAFlagThroughAndNeverAsksWithoutATerminal(t *testing.T) {
	host := terminal("y")
	approved, err := Decide(host, io.Discard, digest, "sha256:given")
	if err != nil || approved != "sha256:given" || len(host.Input) != 1 {
		t.Errorf("Decide(flag) = %q, %v, input left %v", approved, err, host.Input)
	}
	plain := systemtest.New()
	approved, err = Decide(plain, io.Discard, digest, "")
	if err != nil || approved != "" {
		t.Errorf("Decide(no terminal) = %q, %v, want the empty approval", approved, err)
	}
}

func TestDefaultFilePrefersAnAbsoluteXDGDataHome(t *testing.T) {
	env := map[string]string{"XDG_DATA_HOME": "/data"}
	if got := DefaultFile(func(k string) string { return env[k] }, "/home/u"); got != "/data/agent-cli/approvals.json" {
		t.Errorf("DefaultFile() = %q", got)
	}
	env["XDG_DATA_HOME"] = "relative"
	if got := DefaultFile(func(k string) string { return env[k] }, "/home/u"); got != "/home/u/.local/share/agent-cli/approvals.json" {
		t.Errorf("DefaultFile() = %q", got)
	}
}

func TestNewStoreNeedsACleanAbsoluteFile(t *testing.T) {
	for _, file := range []string{"relative.json", "/a/../b.json", "/a//b.json"} {
		if _, err := NewStore(systemtest.New(), file); err == nil {
			t.Errorf("NewStore(%q) error = nil", file)
		}
	}
	store, err := NewStore(systemtest.New(), "/a/approvals.json")
	if err != nil || store.File() != "/a/approvals.json" {
		t.Errorf("NewStore() = %+v, %v", store, err)
	}
}

func TestAnEmptyStoreApprovesNothingAndApproveRecordsTheDigest(t *testing.T) {
	host := systemtest.New()
	store, _ := NewStore(host, "/data/agent-cli/approvals.json")

	if ok, err := store.Approved("kind", digest); ok || err != nil {
		t.Fatalf("Approved() of an empty store = %v, %v", ok, err)
	}
	if err := store.Approve(approval.Approved{Kind: "kind", Subject: "repo", Digest: digest}); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if ok, err := store.Approved("kind", digest); !ok || err != nil {
		t.Errorf("Approved() after Approve = %v, %v", ok, err)
	}
	if ok, _ := store.Approved("other", digest); ok {
		t.Error("an approval of one kind approved another")
	}
	if host.Exists("/data/agent-cli/approvals.json.tmp") {
		t.Error("the staged file was left behind")
	}
}

func TestACorruptLedgerFailsClosed(t *testing.T) {
	host := systemtest.New()
	host.AddFile("/a/approvals.json", "{", 0o600)
	store, _ := NewStore(host, "/a/approvals.json")

	if _, err := store.Ledger(); err == nil {
		t.Error("Ledger() accepted a corrupt file")
	}
	if _, err := store.Approved("k", digest); err == nil {
		t.Error("Approved() accepted a corrupt file")
	}
	if err := store.Approve(approval.Approved{Kind: "k", Digest: digest}); err == nil {
		t.Error("Approve() replaced a corrupt file")
	}
}

// faulty fails one named operation of the host.
type faulty struct {
	*systemtest.Fake
	fail string
}

func (f faulty) err(op string) error {
	if f.fail == op {
		return errors.New(op + " failed")
	}
	return nil
}

func (f faulty) ReadFile(p string) ([]byte, error) {
	if err := f.err("read"); err != nil {
		return nil, err
	}
	return f.Fake.ReadFile(p)
}

func (f faulty) MkdirAll(p string, perm fs.FileMode) error {
	if err := f.err("mkdir"); err != nil {
		return err
	}
	return f.Fake.MkdirAll(p, perm)
}

func (f faulty) WriteFile(p string, data []byte, perm fs.FileMode) error {
	if err := f.err("write"); err != nil {
		return err
	}
	return f.Fake.WriteFile(p, data, perm)
}

func (f faulty) Rename(a, b string) error {
	if err := f.err("rename"); err != nil {
		return err
	}
	return f.Fake.Rename(a, b)
}

func TestEveryFailureOfTheMachineStopsApprove(t *testing.T) {
	for _, op := range []string{"read", "mkdir", "write", "rename"} {
		store, _ := NewStore(faulty{Fake: systemtest.New(), fail: op}, "/a/approvals.json")
		err := store.Approve(approval.Approved{Kind: "k", Subject: "s", Digest: digest})
		if err == nil || !strings.Contains(err.Error(), op+" failed") {
			t.Errorf("%s: Approve() error = %v", op, err)
		}
	}
	store, _ := NewStore(faulty{Fake: systemtest.New(), fail: "read"}, "/a/approvals.json")
	if _, err := store.Ledger(); err == nil {
		t.Error("Ledger() ignored the read failure")
	}
}

func TestApproveRefusesARecordTheLedgerCannotEncode(t *testing.T) {
	store, _ := NewStore(systemtest.New(), "/a/approvals.json")
	if err := store.Approve(approval.Approved{Kind: "", Digest: digest}); err == nil {
		t.Error("Approve() accepted a record without a kind")
	}
}
