package launch

import (
	"errors"
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// openEgress starts the enforcement of a managed run for the transport the
// isolator declared, and returns the route the isolator applies and the stop
// function that ends the enforcement. Any other run keeps its planned route and
// gets a stop function that does nothing.
func openEgress(opener EgressOpener, plan Plan) (isoshared.EgressRoute, func(), error) {
	if plan.Run.Network != network.NetworkManaged {
		return plan.Run.Egress, func() {}, nil
	}
	if opener == nil {
		return isoshared.EgressRoute{}, nil, errors.New("this build has no managed egress")
	}
	route, stop, err := opener.Open(plan.Run.Egress.Transport, plan.Allow)
	if err != nil {
		return isoshared.EgressRoute{}, nil, fmt.Errorf("start managed egress: %w", err)
	}
	return route, stop, nil
}
