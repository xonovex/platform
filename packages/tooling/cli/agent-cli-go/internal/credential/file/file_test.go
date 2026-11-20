package file

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestLookup(t *testing.T) {
	host := systemtest.New().
		AddFile("/secrets/token", "abc123\n", 0o600).
		AddFile("/secrets/open", "abc123", 0o644).
		AddFile("/secrets/empty", "\n", 0o600).
		AddDir("/secrets/dir")
	store := New(host)
	if !store.Available() {
		t.Error("the file store is always available")
	}
	secret, err := store.Lookup("/secrets/token")
	if err != nil || secret.Reveal() != "abc123" {
		t.Errorf("Lookup = %v, %v; want the value without its line break", secret, err)
	}
	cases := map[string]string{
		"relative":    "secrets/token",
		"missing":     "/secrets/none",
		"group read":  "/secrets/open",
		"a directory": "/secrets/dir",
	}
	for name, locator := range cases {
		if _, err := store.Lookup(locator); err == nil {
			t.Errorf("%s: Lookup(%q) succeeded", name, locator)
		}
	}
	if _, err := store.Lookup("/secrets/empty"); !errors.Is(err, credential.ErrNotFound) {
		t.Errorf("an empty file error = %v, want ErrNotFound", err)
	}
	if _, err := store.Lookup("/secrets/open"); err == nil || !strings.Contains(err.Error(), "chmod 600") {
		t.Errorf("the open file error should say how to fix it: %v", err)
	}
}

type unreadable struct{ *systemtest.Fake }

func (unreadable) ReadFile(string) ([]byte, error) { return nil, errors.New("denied") }

func TestLookupReportsAReadFailure(t *testing.T) {
	fake := systemtest.New().AddFile("/secrets/token", "x", 0o600)
	if _, err := New(unreadable{fake}).Lookup("/secrets/token"); err == nil {
		t.Error("a read failure must fail")
	}
}
