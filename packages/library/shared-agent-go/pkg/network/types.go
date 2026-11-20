package network

// NetworkMethod is the network-egress axis.
//
//	host    = share host net with unrestricted egress; explicit opt-in that does
//	          not satisfy RequireEgressRestricted.
//	none    = no network (bwrap --unshare-net / docker --network none); satisfies
//	          RequireEgressRestricted.
//	managed = egress only to the hosts of an allowlist. It satisfies
//	          RequireEgressRestricted only when the selected isolator declares an
//	          EgressTransport that forces every connection through the allowlist.
type NetworkMethod string

const (
	NetworkHost    NetworkMethod = "host"
	NetworkNone    NetworkMethod = "none"
	NetworkManaged NetworkMethod = "managed"
)

// EgressTransport is how an isolator forces the traffic of a managed run through
// the allowlist. The empty value means the isolator has no such route.
//
//	unix-socket = the sandbox has no network at all; the only route out is a unix
//	              socket to the allowlist proxy that the CLI runs on the host.
//	loopback    = the sandbox may connect only to the loopback port of the
//	              allowlist proxy that the CLI runs on the host.
//	policy      = the sandbox runtime enforces the allowlist itself from a policy
//	              document the CLI hands it; the CLI runs no proxy.
type EgressTransport string

const (
	TransportNone       EgressTransport = ""
	TransportUnixSocket EgressTransport = "unix-socket"
	TransportLoopback   EgressTransport = "loopback"
	TransportPolicy     EgressTransport = "policy"
)

// RunsProxy reports whether the CLI itself runs the allowlist proxy for the
// transport.
func (t EgressTransport) RunsProxy() bool {
	return t == TransportUnixSocket || t == TransportLoopback
}

// EgressRestricted reports whether the network method, on an isolator that
// declares transport, restricts egress. Network is a closed enum, so the caller
// computes this boolean and passes it into policy.Capabilities; the policy engine
// stays method-agnostic. host never qualifies, and managed qualifies only when
// the isolator forces the traffic through the allowlist, so a transport-less
// isolator never satisfies the guarantee.
func EgressRestricted(n NetworkMethod, transport EgressTransport) bool {
	switch n {
	case NetworkNone:
		return true
	case NetworkManaged:
		return transport != TransportNone
	default:
		return false
	}
}
