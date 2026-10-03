package shared

import (
	"errors"
	"testing"
)

func TestParseMode(t *testing.T) {
	for _, s := range []string{"host", "none"} {
		if m, err := ParseMode(s); err != nil || string(m) != s {
			t.Errorf("ParseMode(%q) = (%q, %v), want (%q, nil)", s, m, err, s)
		}
	}
	if _, err := ParseMode("proxy"); !errors.Is(err, ErrProxyEnforcementUnavailable) {
		t.Errorf("ParseMode(proxy) error = %v, want %v", err, ErrProxyEnforcementUnavailable)
	}
	if _, err := ParseMode("firewall"); err == nil {
		t.Error("ParseMode(firewall) = nil, want error")
	}
}

func TestEgressIsRestricted(t *testing.T) {
	cases := map[Mode]bool{ModeHost: false, ModeNone: true, ModeProxy: false}
	for m, want := range cases {
		if got := EgressIsRestricted(m); got != want {
			t.Errorf("EgressIsRestricted(%q) = %t, want %t", m, got, want)
		}
	}
}
