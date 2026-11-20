// Package none is the provision=none leaf of the shared plan: the pod image
// supplies every tool, a nix toolchain image included, so the provisioner
// contributes nothing to the pod.
package none

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

// Provisioner contributes nothing.
type Provisioner struct{}

var _ provshared.Provisioner = Provisioner{}

// Contribute returns an empty Contribution.
func (Provisioner) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}

// Pinned reports false: the kube confinement declares the pin of the image.
func (Provisioner) Pinned() bool { return false }

// CheckAgent reports false: the pod image supplies the agent.
func (Provisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}
