package network

import "testing"

func TestEgressIsRestricted(t *testing.T) {
	cases := map[NetworkMethod]bool{
		NetworkHost:  false,
		NetworkNone:  true,
		NetworkProxy: true,
	}
	for net, want := range cases {
		if got := EgressIsRestricted(net); got != want {
			t.Errorf("EgressIsRestricted(%q) = %t, want %t", net, got, want)
		}
	}
}
