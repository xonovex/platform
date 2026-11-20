// Package none is the isolation=none leaf: direct host execution, no namespace
// boundary. It is the single source of the direct-exec command/env.
package none

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/hostlaunch"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// Isolator runs the agent directly on the host — no namespace boundary. It
// applies the provisioner's Contribution (PATH/env/init) but cannot hide host
// tools or restrict network egress. It declares itself unbounded, so the agent
// runs WITHOUT the sandbox permission bypass and keeps its normal host
// permission prompts.
type Isolator struct {
	host hostlaunch.Host
	log  logging.Logger
}

// NewIsolator creates a host (no-isolation) isolator over host, writing through log.
func NewIsolator(host hostlaunch.Host, log logging.Logger) *Isolator {
	return &Isolator{host: host, log: log}
}

// Available always reports true — direct host execution is always possible.
func (i *Isolator) Available() (bool, error) { return true, nil }

// HidesHost reports false: host execution always exposes the host.
func (i *Isolator) HidesHost(bool, string) bool { return false }

// PinnedProvision mirrors the selected provisioner's immutability because host
// execution has no image input.
func (i *Isolator) PinnedProvision(_ provision.ProvisionMethod, provisionerPinned bool, _ string) bool {
	return provisionerPinned
}

// KernelIsolated reports false: there is no namespace or kernel boundary.
func (i *Isolator) KernelIsolated(string) bool { return false }

// ConfinesNetwork reports false: with no namespace, egress stays unrestricted.
func (i *Isolator) ConfinesNetwork() bool { return false }

// EgressTransport declares none: host execution has no boundary to route traffic
// through an allowlist.
func (i *Isolator) EgressTransport() network.EgressTransport { return network.TransportNone }

// Bounded reports false: with no boundary the agent keeps its own permission
// prompts.
func (i *Isolator) Bounded() bool { return false }

// AcceptsHostBinds reports true: the agent runs on the host, which reaches
// every host path already.
func (i *Isolator) AcceptsHostBinds() bool { return true }

// CheckAgent requires the agent executable on the PATH of the host.
func (i *Isolator) CheckAgent(cfg isoshared.RunConfig) error {
	if _, err := i.host.LookPath(cfg.Agent.Binary); err != nil {
		return fmt.Errorf("agent executable %q is not available on host PATH: %w", cfg.Agent.Binary, err)
	}
	return nil
}

// Run executes the agent on the host. It fails CLOSED when a network restriction
// is requested: with no namespace there is no way to enforce restricted egress.
func (i *Isolator) Run(cfg isoshared.RunConfig, c provision.Contribution) (int, error) {
	cmd, env, err := i.hostCommand(cfg, c)
	if err != nil {
		return 1, err
	}
	return hostlaunch.SpawnSandbox(i.host, i.log, cmd[0], cmd[1:], env, cfg.Process.Dir, "host execution")
}

// Command returns the host command (for display / terminal wrappers).
func (i *Isolator) Command(cfg isoshared.RunConfig, c provision.Contribution) ([]string, error) {
	cmd, _, err := i.hostCommand(cfg, c)
	return cmd, err
}

// TerminalCommand returns the host command AND its full resolved environment
// (provider tokens, custom env, and the provisioner's PATH/env). Unlike
// bwrap/docker, host execution bakes nothing into the command, so the wrapper
// must carry this environment for the agent to see its toolchain.
func (i *Isolator) TerminalCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	return i.hostCommand(cfg, c)
}

// hostCommand builds the host command and environment, applying the Contribution.
func (i *Isolator) hostCommand(cfg isoshared.RunConfig, c provision.Contribution) ([]string, []string, error) {
	if cfg.Network != network.NetworkHost {
		return nil, nil, fmt.Errorf("isolation=none cannot restrict network egress (network=%q); use bwrap or docker", cfg.Network)
	}
	if _, err := hostlaunch.ResolveDirectory(i.host, cfg.Process.Dir, "work directory"); err != nil {
		return nil, nil, err
	}
	return hostlaunch.HostExecution(i.host, cfg, c)
}
