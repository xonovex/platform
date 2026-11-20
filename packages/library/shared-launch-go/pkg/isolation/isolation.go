// Package isolation is the isolation axis core: the Confinement and Launcher
// ports, the Process an isolation confines, the neutral RunConfig every
// isolation receives and the EgressRoute of a managed run. It names no concrete
// isolation and starts nothing.
package isolation

import (
	"errors"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// ErrNoImage reports a run on an isolation that starts the agent from a
// container image without an image. No image ships with the platform: the image
// is a setting of the user, so a caller names the setting that supplies it.
var ErrNoImage = errors.New("the isolation starts the agent from a container image, and no image is set")

// Process is the agent process of a run: its argument vector, its environment
// as KEY=VALUE entries, and its working directory. The planner computes it from
// the facts of the agent, and an isolation confines it without changing what it
// runs.
type Process struct {
	Argv []string
	Env  []string
	Dir  string
}

// RunConfig carries the run-wide inputs every isolator needs to confine the
// agent process. Every path in it is absolute, and HomeDir is always set. The
// per-isolator knobs (bwrap HostPassthrough, docker Image and Runtime) sit
// here because a registry builds the isolators before the request of a run is
// known.
type RunConfig struct {
	HomeDir string
	RepoDir string
	// Process is the agent process to confine. Its Dir is the work directory,
	// which the isolation binds read-write and starts the process in.
	Process Process
	// Interactive attaches the process to the terminal of the caller. A run
	// without it gets no terminal and no standard input, so a caller that reads
	// the output of the process can start it without one.
	Interactive bool
	// Name names the sandbox instance of an isolation that keeps one apart from
	// the process that starts it, so StopCommand can end it. Empty leaves the
	// instance unnamed.
	Name string

	// Network is the egress mode; the isolator emits its own network flags via its
	// network.go bridge.
	Network network.NetworkMethod
	// Egress is the route of a managed run: the allowlist and, for a transport
	// that runs a proxy, where to reach it. It is the zero value for any other mode.
	Egress EgressRoute

	// HostPassthrough is the bwrap knob: expose host tool dirs as a fallback.
	HostPassthrough bool
	// Image and Runtime are the docker knobs: the pinned image and a kernel-isolating
	// container runtime (e.g. runsc); empty Runtime means default runc.
	Image   string
	Runtime string

	BindPaths   []string
	RoBindPaths []string
	// ProviderEnv is the environment of the provider route, resolved for this run
	// from the host. Secrets are credential values resolved for this run. Both
	// reach the agent as environment variables only, never as an argument, a file
	// or a log line.
	ProviderEnv map[string]string
	Secrets     map[string]string

	Agent    *types.AgentConfig
	Provider *types.ModelProvider
}

// Confinement declares what an isolation guarantees, so the policy engine and
// the selection never name a concrete isolator. It holds no way to start a
// process: an executor that launches the agent elsewhere implements only this.
type Confinement interface {
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
	// EgressTransport declares how this isolator forces a managed run through the
	// allowlist, or TransportNone when it has no such route. Select refuses a
	// managed run on an isolator that declares none, and a managed run satisfies
	// the egress guarantee only through a declared transport.
	EgressTransport() network.EgressTransport
	// Bounded reports whether a boundary surrounds the agent, so the planner
	// starts the agent without its own permission prompts.
	Bounded() bool
	// AcceptsHostBinds reports whether the isolation can bind a host path into
	// the sandbox. The planner binds the source directory only into an isolation
	// that accepts it, and selection refuses the bind paths of a request for one
	// that does not.
	AcceptsHostBinds() bool
	// RequiresImage reports whether the isolation starts the agent from a
	// container image. Selection refuses a request without an image for one
	// that does, with an error that wraps ErrNoImage.
	RequiresImage() bool
}

// Launcher starts a confined agent process on this machine and applies a
// Provisioner's Contribution through its own mechanism (bwrap binds, docker -v,
// direct host exec). It applies the network mode explicitly.
type Launcher interface {
	// Available reports whether the isolation tool is installed on this machine.
	Available() (bool, error)
	// CheckAgent reports an error when the base of this isolator, the host or
	// the image, cannot supply the agent executable for cfg. It applies only
	// when the provisioner supplies no tools.
	CheckAgent(cfg RunConfig) error
	Run(cfg RunConfig, c provision.Contribution) (int, error)
	Command(cfg RunConfig, c provision.Contribution) ([]string, error)
	// TerminalCommand returns the command AND the environment to launch it under a
	// terminal wrapper. Isolators that bake the environment into the command
	// (bwrap/docker) return the host environment; the host (none) isolator, which
	// bakes nothing, returns the agent command with its full resolved environment
	// (provider tokens, custom env, and the provisioner's PATH/env).
	TerminalCommand(cfg RunConfig, c provision.Contribution) (command []string, env []string, err error)
	// StopCommand returns the command that ends the sandbox instance cfg.Name
	// names after the process of Command ended before the agent did, or nil when
	// ending that process ends the sandbox.
	StopCommand(cfg RunConfig) []string
}

// Isolator is an isolation that an executor both declares and launches on the
// machine it runs on.
type Isolator interface {
	Confinement
	Launcher
}
