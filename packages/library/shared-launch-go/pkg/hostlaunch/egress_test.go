package hostlaunch

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

func TestRunEnvMergesInOrder(t *testing.T) {
	cfg := isoshared.RunConfig{Process: isoshared.Process{Env: []string{"A=custom", "B=custom"}}, Secrets: map[string]string{"B": "secret"}}
	env, err := RunEnv(cfg)
	if err != nil || env["A"] != "custom" || env["B"] != "secret" {
		t.Errorf("RunEnv = %v, %v; want a credential to win over custom env", env, err)
	}
	if _, bad := RunEnv(isoshared.RunConfig{Process: isoshared.Process{Env: []string{"no-equals"}}}); bad == nil {
		t.Error("an invalid custom variable must fail")
	}
}

func TestRunEnvProxyWinsOverTheCaller(t *testing.T) {
	cfg := isoshared.RunConfig{
		Network: network.NetworkManaged,
		Egress:  isoshared.EgressRoute{Transport: network.TransportLoopback, Proxy: "127.0.0.1:40000"},
		Process: isoshared.Process{Env: []string{"HTTPS_PROXY=http://elsewhere:1"}},
	}
	env, err := RunEnv(cfg)
	if err != nil || env["HTTPS_PROXY"] != "http://127.0.0.1:40000" {
		t.Errorf("RunEnv = %v, %v", env, err)
	}
	cfg.Egress = isoshared.EgressRoute{Transport: network.TransportUnixSocket, Socket: "/s", Bridge: "/b"}
	if env, _ := RunEnv(cfg); env["HTTPS_PROXY"] != "http://elsewhere:1" {
		t.Error("only the loopback transport sets the proxy environment: the bridge sets it for unix-socket")
	}
}

func TestResolveBridge(t *testing.T) {
	fsys := oshosttest.New().AddDir("/run/egress").AddExecutable("/opt/agent-cli")
	route := isoshared.EgressRoute{Transport: network.TransportUnixSocket, Socket: "/run/egress/../egress/p.sock", Bridge: "/opt/agent-cli"}
	got, err := ResolveBridge(fsys, route)
	if err != nil || got.Socket != "/run/egress/p.sock" || got.Bridge != "/opt/agent-cli" {
		t.Errorf("ResolveBridge = %+v, %v", got, err)
	}
	cases := map[string]isoshared.EgressRoute{
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
