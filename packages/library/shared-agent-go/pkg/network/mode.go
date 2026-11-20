package network

import (
	"errors"
	"fmt"
)

// ErrManagedEgressUnenforceable reports a managed run on an isolator that has no
// transport to force the traffic through the allowlist.
var ErrManagedEgressUnenforceable = errors.New("network=managed has no enforceable transport on this isolation")

// NeedsConfinement reports whether the method differs from the unrestricted host
// network, so that only an isolator with a network boundary can apply it.
func NeedsConfinement(n NetworkMethod) bool { return n != NetworkHost }

// ParseMode validates s and returns the network method it names.
func ParseMode(s string) (NetworkMethod, error) {
	switch NetworkMethod(s) {
	case NetworkHost, NetworkNone, NetworkManaged:
		return NetworkMethod(s), nil
	default:
		return "", &InvalidModeError{Value: s}
	}
}

// InvalidModeError reports a string that names no network method.
type InvalidModeError struct{ Value string }

func (e *InvalidModeError) Error() string {
	return "unknown network mode " + e.Value + "; valid: host, none, managed"
}

// CheckManaged fails closed when a managed run meets an isolator that declares no
// transport.
func CheckManaged(n NetworkMethod, transport EgressTransport) error {
	if n == NetworkManaged && transport == TransportNone {
		return fmt.Errorf("%w; select an isolation that declares an egress transport", ErrManagedEgressUnenforceable)
	}
	return nil
}
