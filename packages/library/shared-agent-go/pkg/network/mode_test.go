package network

import (
	"errors"
	"testing"
)

func TestParseMode(t *testing.T) {
	for _, s := range []string{"host", "none", "managed"} {
		if m, err := ParseMode(s); err != nil || string(m) != s {
			t.Errorf("ParseMode(%q) = (%q, %v), want (%q, nil)", s, m, err, s)
		}
	}
	for _, s := range []string{"proxy", "firewall", ""} {
		_, err := ParseMode(s)
		var invalid *InvalidModeError
		if !errors.As(err, &invalid) || invalid.Value != s {
			t.Errorf("ParseMode(%q) error = %v, want an InvalidModeError", s, err)
		}
	}
}

func TestInvalidModeErrorNamesTheModes(t *testing.T) {
	msg := (&InvalidModeError{Value: "x"}).Error()
	if msg != "unknown network mode x; valid: host, none, managed" {
		t.Errorf("message = %q", msg)
	}
}

func TestNeedsConfinement(t *testing.T) {
	if NeedsConfinement(NetworkHost) {
		t.Error("NeedsConfinement(host) = true, want false")
	}
	for _, m := range []NetworkMethod{NetworkNone, NetworkManaged} {
		if !NeedsConfinement(m) {
			t.Errorf("NeedsConfinement(%q) = false, want true", m)
		}
	}
}

func TestCheckManaged(t *testing.T) {
	if err := CheckManaged(NetworkManaged, TransportNone); !errors.Is(err, ErrManagedEgressUnenforceable) {
		t.Errorf("CheckManaged(managed, none) = %v, want %v", err, ErrManagedEgressUnenforceable)
	}
	for _, c := range []struct {
		method    NetworkMethod
		transport EgressTransport
	}{{NetworkManaged, TransportPolicy}, {NetworkNone, TransportNone}, {NetworkHost, TransportNone}} {
		if err := CheckManaged(c.method, c.transport); err != nil {
			t.Errorf("CheckManaged(%q, %q) = %v, want nil", c.method, c.transport, err)
		}
	}
}
