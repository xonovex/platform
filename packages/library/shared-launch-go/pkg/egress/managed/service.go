package managed

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// Host is the port the service consumes: the listeners and temporary files of the
// proxy, and the path of this executable, which a sandbox runs as the bridge.
type Host interface {
	Dialer
	Listen(network, address string) (net.Listener, error)
	MkdirTemp(pattern string) (string, error)
	RemoveAll(path string) error
	Executable() (string, error)
}

// Service starts the allowlist proxy of one run.
type Service struct {
	Host Host
	Log  logging.Logger
}

// shutdownTimeout bounds how long Close waits for open tunnels.
const shutdownTimeout = 2 * time.Second

// Open starts a proxy for allow over transport and returns the route to it and
// the function that stops it. The unix-socket transport keeps the socket in a
// private directory; the loopback transport listens on an ephemeral port of the
// host loopback. The policy transport runs no proxy: the sandbox runtime enforces
// the allowlist, so its route only carries it.
func (s Service) Open(transport network.EgressTransport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error) {
	route := isoshared.EgressRoute{Transport: transport, Allow: allow}
	if !transport.RunsProxy() {
		return route, func() {}, route.Validate()
	}
	cleanup := func() {}
	var listener net.Listener
	switch transport {
	case network.TransportUnixSocket:
		bridge, err := s.Host.Executable()
		if err != nil {
			return isoshared.EgressRoute{}, nil, fmt.Errorf("locate the bridge executable: %w", err)
		}
		dir, err := s.Host.MkdirTemp("agent-egress-*")
		if err != nil {
			return isoshared.EgressRoute{}, nil, fmt.Errorf("create the proxy socket directory: %w", err)
		}
		cleanup = func() { _ = s.Host.RemoveAll(dir) }
		route.Socket = filepath.Join(dir, "proxy.sock")
		route.Bridge = bridge
		listener, err = s.Host.Listen("unix", route.Socket)
		if err != nil {
			cleanup()
			return isoshared.EgressRoute{}, nil, fmt.Errorf("listen on the proxy socket: %w", err)
		}
	default:
		var err error
		listener, err = s.Host.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return isoshared.EgressRoute{}, nil, fmt.Errorf("listen on the loopback proxy port: %w", err)
		}
		route.Proxy = listener.Addr().String()
	}
	server := &http.Server{Handler: Proxy{Allow: allow, Dialer: s.Host, Log: s.Log}, ReadHeaderTimeout: 30 * time.Second}
	go func() { _ = server.Serve(listener) }()
	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(ctx)
		_ = server.Close()
		cleanup()
	}
	return route, stop, nil
}
