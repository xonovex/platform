package hostlaunch

import (
	"fmt"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

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
		env = envutil.MergeEnvMaps(env, isoshared.ProxyEnv(cfg.Egress.Proxy))
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
