package bwrap

import netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"

// networkArgs returns the bubblewrap network flags for the mode, applied
// EXPLICITLY. host shares the host network; none and proxy BOTH unshare it.
// Emitting --unshare-net for every non-host method is the regression guard
// against silently leaving egress open.
//
// This is a cross-axis bridge (isolation -> network/shared): the dependent
// isolation leaf owns the glue; network/shared never reaches back into isolation.
//
// Under proxy the namespace is still unshared; reaching the allowlist proxy
// requires the host to wire it into the namespace (slirp/pasta) — that wiring is
// out of scope here, but the env injection (ProxyEnv) and the isolation flag are
// applied so egress is never accidentally left open.
func networkArgs(m netshared.Mode) []string {
	if m == netshared.ModeHost {
		return []string{"--share-net"}
	}
	return []string{"--unshare-net"}
}
