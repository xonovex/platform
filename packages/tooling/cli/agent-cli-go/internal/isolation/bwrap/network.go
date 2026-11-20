package bwrap

import "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"

// networkArgs returns the bubblewrap network flags for the mode. A managed run
// gets no network at all, like none: its only route out is the proxy socket that
// buildArgs binds in, so a direct socket has no interface to use.
//
// This is a cross-axis bridge (isolation -> the shared network package): the
// dependent isolation leaf owns the glue, and the network package never reaches
// back into isolation.
func networkArgs(m network.NetworkMethod) []string {
	if m == network.NetworkHost {
		return []string{"--share-net"}
	}
	return []string{"--unshare-net"}
}
