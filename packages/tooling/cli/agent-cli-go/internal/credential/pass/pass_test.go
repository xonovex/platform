package pass

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestAvailable(t *testing.T) {
	host := systemtest.New()
	host.Env["PATH"] = "/usr/bin"
	if New(host).Available() {
		t.Error("the store must be unavailable without pass")
	}
	host.AddExecutable("/usr/bin/pass")
	if !New(host).Available() {
		t.Error("the store must be available with pass")
	}
}

func TestLookup(t *testing.T) {
	host := systemtest.New()
	host.Outcomes["pass show work/api"] = systemtest.Outcome{Stdout: "line-one\r\nuser: me\n"}
	host.Outcomes["pass show work/blank"] = systemtest.Outcome{Stdout: ""}
	store := New(host)
	if secret, err := store.Lookup("work/api"); err != nil || secret.Reveal() != "line-one" {
		t.Errorf("Lookup = %v, %v; want only the first line", secret, err)
	}
	if _, err := store.Lookup("work/blank"); !errors.Is(err, credential.ErrNotFound) {
		t.Errorf("an empty first line error = %v, want ErrNotFound", err)
	}
	if _, err := store.Lookup("work/absent"); !errors.Is(err, credential.ErrNotFound) {
		t.Errorf("a failed tool error = %v, want ErrNotFound", err)
	}
	for _, bad := range []string{"../etc", "a/../../b", "-x", ""} {
		if _, err := store.Lookup(bad); err == nil {
			t.Errorf("Lookup(%q) succeeded", bad)
		}
	}
}
