package keychain

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
		t.Error("the store must be unavailable without the security tool")
	}
	host.AddExecutable("/usr/bin/security")
	if !New(host).Available() {
		t.Error("the store must be available with the security tool")
	}
}

func TestLookup(t *testing.T) {
	host := systemtest.New()
	host.Outcomes["security find-generic-password -s svc -a acct -w"] = systemtest.Outcome{Stdout: "pw1\n"}
	host.Outcomes["security find-generic-password -s svc -w"] = systemtest.Outcome{Stdout: "pw2\n"}
	host.Outcomes["security find-generic-password -s blank -w"] = systemtest.Outcome{Stdout: "\n"}
	store := New(host)
	if secret, err := store.Lookup("svc/acct"); err != nil || secret.Reveal() != "pw1" {
		t.Errorf("Lookup(svc/acct) = %v, %v", secret, err)
	}
	if secret, err := store.Lookup("svc"); err != nil || secret.Reveal() != "pw2" {
		t.Errorf("Lookup(svc) = %v, %v", secret, err)
	}
	if _, err := store.Lookup("blank"); !errors.Is(err, credential.ErrNotFound) {
		t.Errorf("an empty value error = %v, want ErrNotFound", err)
	}
	if _, err := store.Lookup("absent"); !errors.Is(err, credential.ErrNotFound) {
		t.Errorf("a failed tool error = %v, want ErrNotFound", err)
	}
	for _, bad := range []string{"-s", "svc/", ""} {
		if _, err := store.Lookup(bad); err == nil {
			t.Errorf("Lookup(%q) succeeded", bad)
		}
	}
}
