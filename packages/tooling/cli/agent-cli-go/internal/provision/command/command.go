// Package command is the provision=command leaf: it contributes the init-command
// list, run once before the agent.
package command

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
)

// Provisioner contributes the init-command list.
type Provisioner struct{}

// New creates the command (init-command list) provisioner.
func New() *Provisioner { return &Provisioner{} }

// Contribute returns a Contribution carrying the init commands.
func (Provisioner) Contribute(in provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{InitCommands: in.InitCommands}, nil
}

// Pinned reports false: an init-command list is not a pinned source.
func (Provisioner) Pinned() bool { return false }

// CheckAgent reports true: the init commands run inside the sandbox and may
// install the agent, so no check before the launch can decide it.
func (Provisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return true, nil
}
