package network

import "testing"

func TestEgressRestricted(t *testing.T) {
	cases := []struct {
		name      string
		method    NetworkMethod
		transport EgressTransport
		want      bool
	}{
		{"host", NetworkHost, TransportUnixSocket, false},
		{"none", NetworkNone, TransportNone, true},
		{"managed without a route", NetworkManaged, TransportNone, false},
		{"managed over a unix socket", NetworkManaged, TransportUnixSocket, true},
		{"managed over loopback", NetworkManaged, TransportLoopback, true},
		{"managed by policy", NetworkManaged, TransportPolicy, true},
		{"unknown", NetworkMethod("bogus"), TransportUnixSocket, false},
	}
	for _, c := range cases {
		if got := EgressRestricted(c.method, c.transport); got != c.want {
			t.Errorf("%s: EgressRestricted(%q, %q) = %t, want %t", c.name, c.method, c.transport, got, c.want)
		}
	}
}

func TestRunsProxy(t *testing.T) {
	cases := map[EgressTransport]bool{
		TransportNone:       false,
		TransportUnixSocket: true,
		TransportLoopback:   true,
		TransportPolicy:     false,
	}
	for transport, want := range cases {
		if got := transport.RunsProxy(); got != want {
			t.Errorf("%q.RunsProxy() = %t, want %t", transport, got, want)
		}
	}
}
