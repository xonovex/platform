package isolation

import (
	"strings"
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

func TestWrapWithBridgeAndMounts(t *testing.T) {
	route := EgressRoute{Transport: network.TransportUnixSocket, Socket: "/run/egress/p.sock", Bridge: "/bin/agent-cli"}
	got := strings.Join(WrapWithBridge(route, []string{"agent", "--x"}), " ")
	want := "/bin/agent-cli egress-bridge --socket /run/egress/p.sock --listen 127.0.0.1:3128 -- agent --x"
	if got != want {
		t.Errorf("WrapWithBridge = %q, want %q", got, want)
	}
	dir, exe := BridgeMounts(route)
	if dir != "/run/egress" || exe != "/bin/agent-cli" {
		t.Errorf("BridgeMounts = %q, %q", dir, exe)
	}
}

func TestProxyEnvCoversBothCases(t *testing.T) {
	env := ProxyEnv("127.0.0.1:3128")
	for _, name := range []string{"HTTP_PROXY", "http_proxy", "HTTPS_PROXY", "https_proxy", "ALL_PROXY", "all_proxy"} {
		if env[name] != "http://127.0.0.1:3128" {
			t.Errorf("%s = %q", name, env[name])
		}
	}
	if v, ok := env["NO_PROXY"]; !ok || v != "" {
		t.Error("NO_PROXY must be empty so nothing bypasses the proxy")
	}
}
