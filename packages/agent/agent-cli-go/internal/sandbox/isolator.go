package sandbox

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"

// Isolator confines the agent process and applies a Provisioner's Contribution
// via its own mechanism (bwrap binds / docker -v / direct host exec). It applies
// the NetworkMethod explicitly and declares its host-tools / kernel-isolation
// capabilities so the policy engine never names a concrete isolator.
type Isolator interface {
	Available() (bool, error)
	Run(cfg *types.SandboxConfig, c types.Contribution) (int, error)
	Command(cfg *types.SandboxConfig, c types.Contribution) []string
	// HidesHost reports whether, for this request, host tools are off PATH and not
	// bind-reachable (so RequireHostToolsUnreachable can be honored). It depends on
	// the passthrough knob and, for image-based isolators, whether a pinned image
	// is supplied.
	HidesHost(passthrough bool, image string) bool
	// KernelIsolated reports whether this isolator plus runtime gives a kernel
	// boundary (so RequireKernelIsolation can be honored).
	KernelIsolated(runtime string) bool
}
