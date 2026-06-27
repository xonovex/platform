package sandbox

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"

// Provisioner produces a Contribution (tools/binds/env/init) that an Isolator
// applies, and declares its own guarantees so the policy engine never names it.
type Provisioner interface {
	Contribute(cfg *types.SandboxConfig) (types.Contribution, error)
	// Pinned reports whether provisioning comes from a pinned source (so
	// RequirePinnedProvision can be honored).
	Pinned() bool
}

// noneProvisioner contributes nothing; the host or base image supplies tools
// (subject to HostPassthrough).
type noneProvisioner struct{}

func (noneProvisioner) Contribute(*types.SandboxConfig) (types.Contribution, error) {
	return types.Contribution{}, nil
}
func (noneProvisioner) Pinned() bool { return false }

// commandProvisioner contributes the init-command list, running it once before
// the agent.
type commandProvisioner struct{}

func (commandProvisioner) Contribute(cfg *types.SandboxConfig) (types.Contribution, error) {
	return types.Contribution{InitCommands: cfg.SandboxInitCommands}, nil
}
func (commandProvisioner) Pinned() bool { return false }

// NewNoneProvisioner returns the built-in no-op provisioner.
func NewNoneProvisioner() Provisioner { return noneProvisioner{} }

// NewCommandProvisioner returns the built-in command (init-command list) provisioner.
func NewCommandProvisioner() Provisioner { return commandProvisioner{} }
