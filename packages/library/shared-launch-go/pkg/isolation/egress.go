package isolation

import (
	"fmt"
	"path/filepath"

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
