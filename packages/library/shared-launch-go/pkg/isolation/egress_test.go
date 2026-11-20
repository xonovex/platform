package isolation

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
)

func TestEgressRouteValidate(t *testing.T) {
	allow, err := egress.NewAllowlist([]string{"api.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	good := []EgressRoute{
		{Transport: network.TransportUnixSocket, Socket: "/run/p.sock", Bridge: "/bin/agent-cli"},
		{Transport: network.TransportLoopback, Proxy: "127.0.0.1:40000"},
		{Transport: network.TransportPolicy, Allow: allow},
	}
	for _, route := range good {
		if err := route.Validate(); err != nil {
			t.Errorf("Validate(%+v) = %v", route, err)
		}
	}
	bad := []EgressRoute{
		{},
		{Transport: network.TransportNone},
		{Transport: network.TransportUnixSocket, Socket: "run/p.sock", Bridge: "/bin/agent-cli"},
		{Transport: network.TransportUnixSocket, Socket: "/run/p.sock", Bridge: "agent-cli"},
		{Transport: network.TransportLoopback},
		{Transport: network.TransportPolicy},
	}
	for _, route := range bad {
		if err := route.Validate(); err == nil {
			t.Errorf("Validate(%+v) succeeded", route)
		}
	}
}
