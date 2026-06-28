package docker

import netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"

// networkArgs returns the docker network flags for the mode. none isolates the
// network entirely; host shares the host network; proxy keeps a reachable bridge
// so the allowlist proxy can be reached, with egress constrained by ProxyEnv.
//
// This is a cross-axis bridge (isolation -> network/shared): the dependent
// isolation leaf owns the glue; network/shared never reaches back into isolation.
func networkArgs(m netshared.Mode) []string {
	switch m {
	case netshared.ModeNone:
		return []string{"--network", "none"}
	case netshared.ModeHost:
		return []string{"--network", "host"}
	default: // proxy
		return []string{"--network", "bridge"}
	}
}
