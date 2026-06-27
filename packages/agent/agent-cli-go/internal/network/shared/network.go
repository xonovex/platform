// Package shared is the network axis core. Network is a CLOSED enum
// (host|none|proxy) realized here with no registry — unlike isolation and
// provision, the variant set is fixed, so there is no lazy-factory plug-in
// machinery. The per-isolator network flags live in isolation bridge files
// (isolation/<type>/network.go), which depend on this package one-way only.
package shared

import (
	"os"

	netenum "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/network"
)

// Mode is the network-egress selection. It aliases the shared closed enum so the
// CLI names the axis locally without redefining the variant set (one owner: the
// shared pkg/network package).
type Mode = netenum.NetworkMethod

const (
	ModeHost  = netenum.NetworkHost
	ModeNone  = netenum.NetworkNone
	ModeProxy = netenum.NetworkProxy
)

// ProxyEnvVar names the host environment variable holding the egress-allowlist
// proxy URL used when Mode=proxy. An empty value means no proxy is configured;
// the isolator still applies network isolation (fail closed, not open).
const ProxyEnvVar = "AGENT_SANDBOX_PROXY"

// EgressIsRestricted reports whether the mode restricts egress (none or proxy);
// host shares the host network unrestricted and does not qualify.
func EgressIsRestricted(m Mode) bool { return netenum.EgressIsRestricted(m) }

// ProxyURL returns the configured egress-allowlist proxy URL, or "" if unset.
func ProxyURL() string { return os.Getenv(ProxyEnvVar) }

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

// ProxyEnv returns the proxy environment for Mode=proxy: all egress is routed
// through proxyURL and NO_PROXY is empty so nothing bypasses it. It returns nil
// for any other mode or when proxyURL is empty.
func ProxyEnv(m Mode, proxyURL string) map[string]string {
	if m != ModeProxy || proxyURL == "" {
		return nil
	}
	return map[string]string{
		"HTTP_PROXY":  proxyURL,
		"HTTPS_PROXY": proxyURL,
		"http_proxy":  proxyURL,
		"https_proxy": proxyURL,
		"NO_PROXY":    "",
		"no_proxy":    "",
	}
}
