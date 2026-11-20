// Package provision is the provision axis core: the Provisioner port and the
// neutral Input every provisioner receives. It names no concrete provisioner.
// Input carries the nix source type of the shared agent module, so this core
// depends on no provisioner leaf.
package provision

import (
	agentprovision "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// Input is the neutral per-run carrier handed to Contribute. Each provisioner
// reads only the field it needs (none reads nothing); it holds no method-specific
// behavior flag. NixSource is the value type of the shared agent module, keeping
// this core free of any provisioner leaf.
type Input struct {
	InitCommands []string
	NixSource    sharednix.NixSource
}

// Provisioner produces a Contribution (tools/binds/env/init) that an Isolator
// applies, and declares its own guarantees so the policy engine never names it.
type Provisioner interface {
	Contribute(in Input) (agentprovision.Contribution, error)
	// Pinned reports whether provisioning comes from a pinned source (so
	// RequirePinnedProvision can be honored).
	Pinned() bool
	// CheckAgent reports whether the provisioner accounts for the agent
	// executable. A provisioner that supplies no tools returns false so that the
	// isolator checks its own base; one that must supply the agent and did not
	// returns an error.
	CheckAgent(c agentprovision.Contribution, agent *types.AgentConfig) (bool, error)
}
