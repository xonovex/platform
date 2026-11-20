package shared

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

func TestEgressRestricted(t *testing.T) {
	cases := []struct {
		mode      Mode
		transport Transport
		want      bool
	}{
		{ModeHost, TransportUnixSocket, false},
		{ModeNone, TransportNone, true},
		{ModeManaged, TransportNone, false},
		{ModeManaged, TransportLoopback, true},
	}
	for _, c := range cases {
		if got := EgressRestricted(c.mode, c.transport); got != c.want {
			t.Errorf("EgressRestricted(%q, %q) = %t, want %t", c.mode, c.transport, got, c.want)
		}
	}
}

func TestNeedsConfinement(t *testing.T) {
	if NeedsConfinement(ModeHost) {
		t.Error("NeedsConfinement(host) = true, want false")
	}
	for _, m := range []Mode{ModeNone, ModeManaged} {
		if !NeedsConfinement(m) {
			t.Errorf("NeedsConfinement(%q) = false, want true", m)
		}
	}
}

func TestCheckManaged(t *testing.T) {
	if err := CheckManaged(ModeManaged, TransportNone); !errors.Is(err, ErrManagedEgressUnenforceable) {
		t.Errorf("CheckManaged(managed, none) = %v, want %v", err, ErrManagedEgressUnenforceable)
	}
	for _, c := range []struct {
		mode      Mode
		transport Transport
	}{{ModeManaged, TransportPolicy}, {ModeNone, TransportNone}, {ModeHost, TransportNone}} {
		if err := CheckManaged(c.mode, c.transport); err != nil {
			t.Errorf("CheckManaged(%q, %q) = %v, want nil", c.mode, c.transport, err)
		}
	}
}
