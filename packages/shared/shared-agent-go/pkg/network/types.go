package network

// NetworkMethod is the network-egress axis, replacing the old boolean Network.
//
//	host  = share host net, unrestricted egress — today's de-facto behavior, now
//	        an EXPLICIT opt-in; does NOT satisfy RequireEgressRestricted.
//	none  = no network (bwrap --unshare-net / docker --network none); satisfies
//	        RequireEgressRestricted.
//	proxy = egress ONLY via a host-side allowlist HTTP(S) proxy; link-local,
//	        metadata (169.254.169.254), RFC1918, and loopback denied; satisfies
//	        RequireEgressRestricted. Recommended default for untrusted code that
//	        still needs the model API.
type NetworkMethod string

const (
	NetworkHost  NetworkMethod = "host"
	NetworkNone  NetworkMethod = "none"
	NetworkProxy NetworkMethod = "proxy"
)

// DefaultEgressAllowlist seeds NetworkProxy: provider API endpoints plus common
// package registries and git forges. The CLI `--egress-allow` (repeatable)
// EXTENDS, not replaces, this set.
var DefaultEgressAllowlist = []string{
	// Provider API endpoints
	"api.anthropic.com",
	"api.z.ai",
	// Package registries
	"registry.npmjs.org",
	"pypi.org",
	"files.pythonhosted.org",
	"crates.io",
	"static.crates.io",
	"proxy.golang.org",
	"sum.golang.org",
	// Git forges
	"github.com",
	"codeload.github.com",
	"raw.githubusercontent.com",
	"objects.githubusercontent.com",
	"gitlab.com",
}

// EgressIsRestricted reports whether the network method restricts egress. Network
// is a closed enum (no per-network plugin object), so the caller computes this
// boolean and passes it into policy.Capabilities; the policy engine stays
// method-agnostic. NetworkHost shares the host network unrestricted and does not
// qualify.
func EgressIsRestricted(n NetworkMethod) bool {
	return n == NetworkNone || n == NetworkProxy
}
