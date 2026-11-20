// Package shared is the network axis core. Network is a closed enum: host, none
// and managed. Managed reaches only the hosts of an allowlist, and the isolator
// declares the transport that forces every connection through it. The
// per-isolator network flags live in isolation bridge files, which depend on this
// package one-way only.
package shared

import (
	"errors"
	"fmt"

	netenum "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
)

// ErrManagedEgressUnenforceable reports a managed run on an isolator that has no
// transport to force the traffic through the allowlist.
var ErrManagedEgressUnenforceable = errors.New("network=managed has no enforceable transport on this isolation")

// Mode is the network-egress selection. It aliases the shared closed enum so the
// CLI names the axis locally without redefining the variant set (one owner: the
// shared pkg/network package).
type Mode = netenum.NetworkMethod

// Transport is how an isolator forces a managed run through the allowlist.
type Transport = netenum.EgressTransport

const (
	ModeHost    = netenum.NetworkHost
	ModeNone    = netenum.NetworkNone
	ModeManaged = netenum.NetworkManaged

	TransportNone       = netenum.TransportNone
	TransportUnixSocket = netenum.TransportUnixSocket
	TransportLoopback   = netenum.TransportLoopback
	TransportPolicy     = netenum.TransportPolicy
)

// EgressRestricted reports whether the mode, on an isolator that declares
// transport, enforceably restricts egress.
func EgressRestricted(m Mode, transport Transport) bool {
	return netenum.EgressRestricted(m, transport)
}

// NeedsConfinement reports whether the mode differs from the unrestricted host
// network, so that only an isolator with a network boundary can apply it.
func NeedsConfinement(m Mode) bool { return m != ModeHost }

// ParseMode validates s and returns the corresponding Mode.
func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeHost, ModeNone, ModeManaged:
		return Mode(s), nil
	default:
		return "", &InvalidModeError{Value: s}
	}
}

// InvalidModeError reports an unrecognised network mode string.
type InvalidModeError struct{ Value string }

func (e *InvalidModeError) Error() string {
	return "unknown network mode " + e.Value + "; valid: host, none, managed"
}

// CheckManaged fails closed when a managed run meets an isolator that declares no
// transport.
func CheckManaged(m Mode, transport Transport) error {
	if m == ModeManaged && transport == TransportNone {
		return fmt.Errorf("%w; select an isolation that declares an egress transport", ErrManagedEgressUnenforceable)
	}
	return nil
}
