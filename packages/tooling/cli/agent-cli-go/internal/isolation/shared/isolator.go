// Package shared is the isolation axis core: the Isolator port, the Host port
// the isolators consume, the neutral RunConfig the composition root hands every
// isolator, and the path checks they share. It names no concrete
// isolator (none/bwrap/docker) and no sibling axis leaf; the per-isolator network
// flags live in isolation/<type>/network.go bridge files.
package shared

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
)

// RunConfig carries the run-wide inputs every isolator needs to launch the
// agent. Every path in it is absolute, and HomeDir is always set. The
// per-isolator knobs (bwrap HostPassthrough, docker Image and Runtime) sit
// here because the composition root builds the isolators before the flags are
// parsed.
type RunConfig struct {
	HomeDir string
	WorkDir string
	RepoDir string

	// Network is the egress mode; the isolator emits its own network flags via its
	// network.go bridge.
	Network netshared.Mode

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
}

// Isolator confines the agent process and applies a Provisioner's Contribution
// via its own mechanism (bwrap binds / docker -v / direct host exec). It applies
// the network mode explicitly and declares its host-tools / kernel-isolation
// capabilities so the policy engine never names a concrete isolator.
type Isolator interface {
	Available() (bool, error)
	Run(cfg RunConfig, c provision.Contribution) (int, error)
	Command(cfg RunConfig, c provision.Contribution) ([]string, error)
	// TerminalCommand returns the command AND the environment to launch it under a
	// terminal wrapper. Isolators that bake the environment into the command
	// (bwrap/docker) return the host environment; the host (none) isolator, which
	// bakes nothing, returns the agent command with its full resolved environment
	// (provider tokens, custom env, and the provisioner's PATH/env).
	TerminalCommand(cfg RunConfig, c provision.Contribution) (command []string, env []string, err error)
	// PinnedProvision reports whether all tool inputs used by this isolator are
	// immutable. Image-based isolators must account for the image as well as any
	// separate provisioner contribution.
	PinnedProvision(method provision.ProvisionMethod, provisionerPinned bool, image string) bool
	// HidesHost reports whether, for this request, host tools are off PATH and not
	// bind-reachable (so RequireHostToolsUnreachable can be honored). It depends on
	// the passthrough knob and, for image-based isolators, whether a pinned image
	// is supplied.
	HidesHost(passthrough bool, image string) bool
	// KernelIsolated reports whether this isolator plus runtime gives a kernel
	// boundary (so RequireKernelIsolation can be honored).
	KernelIsolated(runtime string) bool
	// ConfinesNetwork reports whether this isolator can apply a network mode
	// other than host.
	ConfinesNetwork() bool
	// CheckAgent reports an error when the base of this isolator, the host or
	// the image, cannot supply the agent executable for cfg. It applies only
	// when the provisioner supplies no tools.
	CheckAgent(cfg RunConfig) error
}
