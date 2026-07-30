package bwrap

import netshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/network/shared"

// networkArgs returns the bubblewrap network flags for the mode. Proxy mode is
// rejected until a transport exists that reaches an allowlist proxy without
// also exposing unrestricted network routes.
//
// This is a cross-axis bridge (isolation -> network/shared): the dependent
// isolation leaf owns the glue; network/shared never reaches back into isolation.
func networkArgs(m netshared.Mode) ([]string, error) {
	if m == netshared.ModeHost {
		return []string{"--share-net"}, nil
	}
	if m == netshared.ModeProxy {
		return nil, netshared.ErrProxyEnforcementUnavailable
	}
	return []string{"--unshare-net"}, nil
}
