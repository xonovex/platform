// Package shared is the isolation axis core: the Isolator port plus the neutral
// RunConfig the composition root hands every isolator. It names no concrete
// isolator (none/bwrap/docker) and no sibling axis leaf; the per-isolator network
// flags live in isolation/<type>/network.go bridge files.
package shared

import (
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// RunConfig carries the run-wide inputs every isolator needs to launch the agent.
// It replaces the old flat SandboxConfig god-struct: it holds only the fields the
// isolators actually consume (data coupling), and the provisioning inputs live in
// the provision axis, not here. Per-isolator knobs (docker Image/Runtime, bwrap
// HostPassthrough) are assembled into their leaf Options at the composition root
// and copied onto the fields below.
type RunConfig struct {
	HomeDir string
	WorkDir string
	RepoDir string

	// Network is the egress mode; the isolator emits its own network flags via its
	// network.go bridge and applies ProxyEnv (already resolved, allowlist folded in).
	Network  netshared.Mode
	ProxyEnv map[string]string

	// HostPassthrough is the bwrap knob: expose host tool dirs as a fallback.
	HostPassthrough bool
	// Image and Runtime are the docker knobs: the pinned image and a kernel-isolating
	// container runtime (e.g. runsc); empty Runtime means default runc.
	Image   string
	Runtime string

	BindPaths   []string
	RoBindPaths []string
	CustomEnv   []string

	Agent     *types.AgentConfig
	Provider  *types.ModelProvider
	AgentArgs []string

	Verbose bool
}

// Isolator confines the agent process and applies a Provisioner's Contribution
// via its own mechanism (bwrap binds / docker -v / direct host exec). It applies
// the network mode explicitly and declares its host-tools / kernel-isolation
// capabilities so the policy engine never names a concrete isolator.
type Isolator interface {
	Available() (bool, error)
	Run(cfg RunConfig, c provision.Contribution) (int, error)
	Command(cfg RunConfig, c provision.Contribution) []string
	// TerminalCommand returns the command AND the environment to launch it under a
	// terminal wrapper. Isolators that bake the environment into the command
	// (bwrap/docker) return the host environment; the host (none) isolator, which
	// bakes nothing, returns the agent command with its full resolved environment
	// (provider tokens, custom env, and the provisioner's PATH/env).
	TerminalCommand(cfg RunConfig, c provision.Contribution) (command []string, env []string)
	// HidesHost reports whether, for this request, host tools are off PATH and not
	// bind-reachable (so RequireHostToolsUnreachable can be honored). It depends on
	// the passthrough knob and, for image-based isolators, whether a pinned image
	// is supplied.
	HidesHost(passthrough bool, image string) bool
	// KernelIsolated reports whether this isolator plus runtime gives a kernel
	// boundary (so RequireKernelIsolation can be honored).
	KernelIsolated(runtime string) bool
}
