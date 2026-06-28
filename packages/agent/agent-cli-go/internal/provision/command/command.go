// Package command is the provision=command leaf: it contributes the init-command
// list, run once before the agent.
package command

import (
	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
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
