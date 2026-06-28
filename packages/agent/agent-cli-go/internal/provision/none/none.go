// Package none is the provision=none leaf: it contributes nothing; the host or
// base image supplies tools (subject to the isolator's host passthrough).
package none

import (
	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
)

// Provisioner contributes nothing.
type Provisioner struct{}

// New creates the no-op provisioner.
func New() *Provisioner { return &Provisioner{} }

// Contribute returns an empty Contribution.
func (Provisioner) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}

// Pinned reports false: no provisioning source.
func (Provisioner) Pinned() bool { return false }
