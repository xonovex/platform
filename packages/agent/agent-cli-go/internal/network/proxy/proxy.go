// Package proxy is the network=proxy leaf: egress only via a host-side allowlist
// HTTP(S) proxy. It owns the proxy-specific Options (allowlist + URL); host and
// none carry no per-type data, so the closed enum has no host/none leaf — the
// asymmetry is deliberate (see network/shared).
package proxy

import (
	"strings"

	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
)

// EgressAllowEnvVar names the environment variable the host-side proxy reads to
// learn the extra allowlist hosts. The proxy itself enforces the allowlist;
// this carries it into the sandbox env so the configured proxy can honor it.
const EgressAllowEnvVar = "SANDBOX_EGRESS_ALLOW"

// Options is the proxy variant's immutable configuration: the egress allowlist
// (extends the shared default) and the proxy URL.
type Options struct {
	EgressAllowlist []string
	URL             string
}

// Env returns the proxy egress environment for these options, or nil when no
// proxy URL is configured. The allowlist is exported via EgressAllowEnvVar so it
// is honored rather than written-and-ignored.
func (o Options) Env() map[string]string {
	env := netshared.ProxyEnv(netshared.ModeProxy, o.URL)
	if env == nil {
		return nil
	}
	if len(o.EgressAllowlist) > 0 {
		env[EgressAllowEnvVar] = strings.Join(o.EgressAllowlist, ",")
	}
	return env
}
