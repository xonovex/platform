package shared

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// BridgeListen is the loopback address the in-sandbox bridge listens on. A
// sandbox with no network of its own has a private loopback, so the port cannot
// clash with a host service.
const BridgeListen = "127.0.0.1:3128"

// WrapWithBridge runs command under the bridge: the bridge listens on
// BridgeListen, relays each connection to the proxy socket, points the proxy
// environment of command at itself, and exits with the code of command.
func WrapWithBridge(route isoshared.EgressRoute, command []string) []string {
	return append([]string{route.Bridge, "egress-bridge", "--socket", route.Socket, "--listen", BridgeListen, "--"}, command...)
}

// BridgeMounts returns the host directory that holds the proxy socket, which the
// sandbox binds read-write so it can connect, and the bridge executable, which
// it binds read-only.
func BridgeMounts(route isoshared.EgressRoute) (socketDir, executable string) {
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

// RunEnv merges the environment of the process, the credential values and, for
// the loopback transport, the proxy environment of cfg. The last source wins, and the
// proxy environment wins over the caller so a managed run cannot be pointed away
// from its proxy by a custom variable.
func RunEnv(cfg isoshared.RunConfig) (map[string]string, error) {
	custom, err := envutil.ParseCustomEnv(cfg.Process.Env)
	if err != nil {
		return nil, err
	}
	env := envutil.MergeEnvMaps(custom, cfg.Secrets)
	if cfg.Network == network.NetworkManaged && cfg.Egress.Transport == network.TransportLoopback {
		env = envutil.MergeEnvMaps(env, ProxyEnv(cfg.Egress.Proxy))
	}
	return env, nil
}

// ResolveBridge validates a unix-socket route against the host: the socket
// directory and the bridge executable must exist, and both paths come back clean.
func ResolveBridge(fsys FileSystem, route isoshared.EgressRoute) (isoshared.EgressRoute, error) {
	if route.Transport != network.TransportUnixSocket {
		return isoshared.EgressRoute{}, fmt.Errorf("transport %q does not run the bridge", route.Transport)
	}
	if err := route.Validate(); err != nil {
		return isoshared.EgressRoute{}, err
	}
	dir, err := ResolveDirectory(fsys, filepath.Dir(route.Socket), "egress socket directory")
	if err != nil {
		return isoshared.EgressRoute{}, err
	}
	bridge, err := ResolveExistingPath(fsys, route.Bridge, "egress bridge executable")
	if err != nil {
		return isoshared.EgressRoute{}, err
	}
	route.Socket = filepath.Join(dir, filepath.Base(route.Socket))
	route.Bridge = bridge
	return route, nil
}
