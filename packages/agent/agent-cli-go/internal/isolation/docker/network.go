package docker

import netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"

// networkArgs returns the docker network flags for the mode. Proxy mode is
// rejected because a bridge plus proxy environment variables does not prevent
// untrusted code from opening direct sockets.
//
// This is a cross-axis bridge (isolation -> network/shared): the dependent
// isolation leaf owns the glue; network/shared never reaches back into isolation.
func networkArgs(m netshared.Mode) ([]string, error) {
	switch m {
	case netshared.ModeNone:
		return []string{"--network", "none"}, nil
	case netshared.ModeHost:
		return []string{"--network", "host"}, nil
	default:
		return nil, netshared.ErrProxyEnforcementUnavailable
	}
}
