package network

// NetworkMethod is the network-egress axis.
//
//	host  = share host net with unrestricted egress; explicit opt-in that does
//	        not satisfy RequireEgressRestricted.
//	none  = no network (bwrap --unshare-net / docker --network none); satisfies
//	        RequireEgressRestricted.
//	proxy = reserved for an enforceable allowlist transport; it does not satisfy
//	        RequireEgressRestricted until a realizer provides that transport.
type NetworkMethod string

const (
	NetworkHost  NetworkMethod = "host"
	NetworkNone  NetworkMethod = "none"
	NetworkProxy NetworkMethod = "proxy"
)

// EgressIsRestricted reports whether the network method restricts egress. Network
// is a closed enum (no per-network plugin object), so the caller computes this
// boolean and passes it into policy.Capabilities; the policy engine stays
// method-agnostic. NetworkHost shares the host network unrestricted and does not
// qualify.
func EgressIsRestricted(n NetworkMethod) bool {
	return n == NetworkNone
}
