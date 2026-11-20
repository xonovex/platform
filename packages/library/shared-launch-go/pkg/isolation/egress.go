package isolation

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
)

// EgressRoute is the route the traffic of a managed run takes.
type EgressRoute struct {
	Transport network.EgressTransport
	// Allow is the allowlist: the proxy enforces it for a proxy transport, and the
	// sandbox runtime enforces it for the policy transport.
	Allow egress.Allowlist
	// Socket is the host path of the proxy socket (unix-socket transport).
	Socket string
	// Proxy is the host:port of the proxy on the host loopback (loopback transport).
	Proxy string
	// Bridge is the absolute path of the executable that runs inside the sandbox
	// to relay its loopback port to Socket (unix-socket transport).
	Bridge string
}

// Validate reports whether the route holds what its transport needs.
func (r EgressRoute) Validate() error {
	switch r.Transport {
	case network.TransportUnixSocket:
		if !filepath.IsAbs(r.Socket) || !filepath.IsAbs(r.Bridge) {
			return fmt.Errorf("managed egress over a unix socket needs an absolute socket and bridge executable")
		}
	case network.TransportLoopback:
		if r.Proxy == "" {
			return fmt.Errorf("managed egress over loopback needs the proxy address")
		}
	case network.TransportPolicy:
		if len(r.Allow.Rules()) == 0 {
			return fmt.Errorf("managed egress by policy needs an allowlist")
		}
	default:
		return fmt.Errorf("managed egress needs a transport")
	}
	return nil
}

// BridgeListen is the loopback address the in-sandbox bridge listens on. A
// sandbox with no network of its own has a private loopback, so the port cannot
// clash with a host service.
const BridgeListen = "127.0.0.1:3128"

// WrapWithBridge runs command under the bridge: the bridge listens on
// BridgeListen, relays each connection to the proxy socket, points the proxy
// environment of command at itself, and exits with the code of command.
func WrapWithBridge(route EgressRoute, command []string) []string {
	return append([]string{route.Bridge, "egress-bridge", "--socket", route.Socket, "--listen", BridgeListen, "--"}, command...)
}

// BridgeMounts returns the host directory that holds the proxy socket, which the
// sandbox binds read-write so it can connect, and the bridge executable, which
// it binds read-only.
func BridgeMounts(route EgressRoute) (socketDir, executable string) {
	return filepath.Dir(route.Socket), route.Bridge
}

// ProxyEnv returns the environment that points an agent at an HTTP proxy.
func ProxyEnv(address string) map[string]string {
	url := "http://" + address
	env := map[string]string{"NO_PROXY": "", "no_proxy": ""}
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY"} {
		env[name] = url
		env[strings.ToLower(name)] = url
	}
	return env
}
