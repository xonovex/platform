package sandboxutil

import (
	"os"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// ProxyEnvVar names the host environment variable holding the egress-allowlist
// proxy URL used when NetworkMethod=proxy. An empty value means no proxy is
// configured; the isolator still applies network isolation (fail closed, not
// open).
const ProxyEnvVar = "AGENT_SANDBOX_PROXY"

// BwrapNetworkArgs returns the bubblewrap network flags for net, applied
// EXPLICITLY. host shares the host network; none and proxy BOTH unshare it.
// Emitting --unshare-net for every non-host method is the regression guard
// against collapsing the old nix/nixflake tiers (which unshared the net) and
// silently dropping that isolation.
//
// Under proxy the namespace is still unshared; reaching the allowlist proxy
// requires the host to wire it into the namespace (slirp/pasta) — that wiring is
// out of scope here, but the env injection (ProxyEnv) and the isolation flag are
// applied so egress is never accidentally left open.
func BwrapNetworkArgs(net types.NetworkMethod) []string {
	if net == types.NetworkHost {
		return []string{"--share-net"}
	}
	return []string{"--unshare-net"}
}

// DockerNetworkArgs returns the docker network flags for net. none isolates the
// network entirely; host shares the host network; proxy keeps a reachable bridge
// so the allowlist proxy can be reached, with egress constrained by ProxyEnv.
func DockerNetworkArgs(net types.NetworkMethod) []string {
	switch net {
	case types.NetworkNone:
		return []string{"--network", "none"}
	case types.NetworkHost:
		return []string{"--network", "host"}
	default: // proxy
		return []string{"--network", "bridge"}
	}
}

// ProxyURL returns the configured egress-allowlist proxy URL, or "" if unset.
func ProxyURL() string {
	return os.Getenv(ProxyEnvVar)
}

// ProxyEnv returns the proxy environment for NetworkMethod=proxy: all egress is
// routed through proxyURL and NO_PROXY is empty so nothing bypasses it. It
// returns nil for any other method or when proxyURL is empty.
func ProxyEnv(net types.NetworkMethod, proxyURL string) map[string]string {
	if net != types.NetworkProxy || proxyURL == "" {
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
