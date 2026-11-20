package env

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestLookup(t *testing.T) {
	host := systemtest.New()
	host.Env["TOKEN"] = "s3cret"
	store := New(host)
	if !store.Available() {
		t.Error("the env store is always available")
	}
	secret, err := store.Lookup("TOKEN")
	if err != nil || secret.Reveal() != "s3cret" {
		t.Errorf("Lookup(TOKEN) = %v, %v", secret, err)
	}
	if _, err := store.Lookup("MISSING"); !errors.Is(err, credential.ErrNotFound) {
		t.Errorf("an unset variable error = %v, want ErrNotFound", err)
	}
	if _, err := store.Lookup("not a name"); err == nil {
		t.Error("an invalid variable name must fail")
	}
}
