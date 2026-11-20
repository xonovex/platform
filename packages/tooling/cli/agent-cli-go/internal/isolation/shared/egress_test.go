package shared

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func policyRoute(t *testing.T) EgressRoute {
	t.Helper()
	allow, err := egress.NewAllowlist([]string{"api.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	return EgressRoute{Transport: network.TransportPolicy, Allow: allow}
}

func TestEgressRouteValidate(t *testing.T) {
	good := []EgressRoute{
		{Transport: network.TransportUnixSocket, Socket: "/run/p.sock", Bridge: "/bin/agent-cli"},
		{Transport: network.TransportLoopback, Proxy: "127.0.0.1:40000"},
		policyRoute(t),
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

func TestRunEnvMergesInOrder(t *testing.T) {
	cfg := RunConfig{CustomEnv: []string{"A=custom", "B=custom"}, Secrets: map[string]string{"B": "secret"}}
	env, err := RunEnv(cfg)
	if err != nil || env["A"] != "custom" || env["B"] != "secret" {
		t.Errorf("RunEnv = %v, %v; want a credential to win over custom env", env, err)
	}
	if _, bad := RunEnv(RunConfig{CustomEnv: []string{"no-equals"}}); bad == nil {
		t.Error("an invalid custom variable must fail")
	}
}

func TestRunEnvProxyWinsOverTheCaller(t *testing.T) {
	cfg := RunConfig{
		Network:   network.NetworkManaged,
		Egress:    EgressRoute{Transport: network.TransportLoopback, Proxy: "127.0.0.1:40000"},
		CustomEnv: []string{"HTTPS_PROXY=http://elsewhere:1"},
	}
	env, err := RunEnv(cfg)
	if err != nil || env["HTTPS_PROXY"] != "http://127.0.0.1:40000" {
		t.Errorf("RunEnv = %v, %v", env, err)
	}
	cfg.Egress = EgressRoute{Transport: network.TransportUnixSocket, Socket: "/s", Bridge: "/b"}
	if env, _ := RunEnv(cfg); env["HTTPS_PROXY"] != "http://elsewhere:1" {
		t.Error("only the loopback transport sets the proxy environment: the bridge sets it for unix-socket")
	}
}

func TestResolveBridge(t *testing.T) {
	fsys := systemtest.New().AddDir("/run/egress").AddExecutable("/opt/agent-cli")
	route := EgressRoute{Transport: network.TransportUnixSocket, Socket: "/run/egress/../egress/p.sock", Bridge: "/opt/agent-cli"}
	got, err := ResolveBridge(fsys, route)
	if err != nil || got.Socket != "/run/egress/p.sock" || got.Bridge != "/opt/agent-cli" {
		t.Errorf("ResolveBridge = %+v, %v", got, err)
	}
	cases := map[string]EgressRoute{
		"wrong transport":   {Transport: network.TransportLoopback, Proxy: "127.0.0.1:1"},
		"invalid route":     {Transport: network.TransportUnixSocket, Socket: "p.sock", Bridge: "/opt/agent-cli"},
		"missing directory": {Transport: network.TransportUnixSocket, Socket: "/nope/p.sock", Bridge: "/opt/agent-cli"},
		"missing bridge":    {Transport: network.TransportUnixSocket, Socket: "/run/egress/p.sock", Bridge: "/opt/missing"},
	}
	for name, route := range cases {
		if _, err := ResolveBridge(fsys, route); err == nil {
			t.Errorf("%s: ResolveBridge succeeded", name)
		}
	}
}
