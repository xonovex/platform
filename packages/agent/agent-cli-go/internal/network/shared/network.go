// Package shared is the network axis core. Network is a CLOSED enum
// (host|none|proxy) realized here with no registry — unlike isolation and
// provision, the variant set is fixed, so there is no lazy-factory plug-in
// machinery. The per-isolator network flags live in isolation bridge files
// (isolation/<type>/network.go), which depend on this package one-way only.
package shared

import (
	"errors"

	netenum "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/network"
)

var ErrProxyEnforcementUnavailable = errors.New("network=proxy has no enforceable transport backend")

// Mode is the network-egress selection. It aliases the shared closed enum so the
// CLI names the axis locally without redefining the variant set (one owner: the
// shared pkg/network package).
type Mode = netenum.NetworkMethod

const (
	ModeHost  = netenum.NetworkHost
	ModeNone  = netenum.NetworkNone
	ModeProxy = netenum.NetworkProxy
)

// EgressIsRestricted reports whether the mode restricts egress (none or proxy);
// host shares the host network unrestricted and does not qualify.
func EgressIsRestricted(m Mode) bool { return netenum.EgressIsRestricted(m) }

// ParseMode validates s and returns the corresponding Mode.
func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeHost, ModeNone, ModeProxy:
		return Mode(s), nil
	default:
		return "", &InvalidModeError{Value: s}
	}
}

// InvalidModeError reports an unrecognised network mode string.
type InvalidModeError struct{ Value string }

func (e *InvalidModeError) Error() string {
	return "unknown network mode " + e.Value + "; valid: host, none, proxy"
}
