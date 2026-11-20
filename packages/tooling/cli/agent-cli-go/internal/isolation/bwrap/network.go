package bwrap

import netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"

// networkArgs returns the bubblewrap network flags for the mode. A managed run
// gets no network at all, like none: its only route out is the proxy socket that
// buildArgs binds in, so a direct socket has no interface to use.
//
// This is a cross-axis bridge (isolation -> network/shared): the dependent
// isolation leaf owns the glue; network/shared never reaches back into isolation.
func networkArgs(m netshared.Mode) []string {
	if m == netshared.ModeHost {
		return []string{"--share-net"}
	}
	return []string{"--unshare-net"}
}
