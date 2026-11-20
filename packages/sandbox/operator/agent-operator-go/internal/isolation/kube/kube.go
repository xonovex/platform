// Package kube is the isolation=kube leaf: the Confinement of an agent that runs
// as a Job pod. The operator builds the pod from the plan and never starts a
// process itself, so the leaf declares guarantees and launches nothing.
package kube

import (
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/validator"
)

// Method is the isolation method of the leaf.
const Method isolation.IsolationMethod = "kube"

// Confinement is a hardened Job pod: no host path, a read-only root file system,
// a per-run NetworkPolicy, and the runtime classes the namespace policy
// approves as sandboxed.
type Confinement struct {
	runtimeClasses []string
}

var _ isoshared.Confinement = Confinement{}

// New returns the confinement of a namespace whose AgentPolicy approves
// runtimeClasses as sandboxed runtimes.
func New(runtimeClasses []string) Confinement {
	return Confinement{runtimeClasses: slices.Clone(runtimeClasses)}
}

// PinnedProvision reports whether the pod image is digest-pinned. The image
// supplies every tool of the pod, a nix toolchain included, so the digest is
// the whole provisioning source.
func (Confinement) PinnedProvision(_ provision.ProvisionMethod, _ bool, image string) bool {
	return validator.ValidatePinnedImageReference(image) == nil
}

// HidesHost reports true: a pod mounts no host path and has no host tool on
// its PATH.
func (Confinement) HidesHost(bool, string) bool { return true }

// KernelIsolated reports whether the namespace policy approves runtime as a
// sandboxed runtime class.
func (c Confinement) KernelIsolated(runtime string) bool {
	return strings.TrimSpace(runtime) != "" && slices.Contains(c.runtimeClasses, runtime)
}

// ConfinesNetwork reports true: the per-run NetworkPolicy applies the network
// mode.
func (Confinement) ConfinesNetwork() bool { return true }

// EgressTransport reports TransportNone: no backend forces pod traffic through
// an FQDN allowlist, so a managed run is refused.
func (Confinement) EgressTransport() network.EgressTransport { return network.TransportNone }

// Bounded reports true: the pod surrounds the agent.
func (Confinement) Bounded() bool { return true }

// AcceptsHostBinds reports false: a pod reaches only its volumes.
func (Confinement) AcceptsHostBinds() bool { return false }

// RequiresImage reports true: a pod starts from its image.
func (Confinement) RequiresImage() bool { return true }
