package sandbox

import (
	"errors"
	"fmt"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// Policy errors. EnforcePolicy wraps one of these (with context) for the first
// unmet guarantee; callers match with errors.Is to react to a specific failure.
var (
	// ErrPinnedProvisioningUnmet: RequirePinnedProvisioning requested but the
	// provisioning method is not pinned.
	ErrPinnedProvisioningUnmet = errors.New("pinned provisioning unmet")
	// ErrHostToolsReachable: RequireHostToolsUnreachable requested but the
	// isolation axis leaves host tools reachable.
	ErrHostToolsReachable = errors.New("host tools reachable")
	// ErrEgressUnrestricted: RequireEgressRestricted requested but the network
	// method does not restrict egress.
	ErrEgressUnrestricted = errors.New("egress unrestricted")
	// ErrKernelIsolationUnmet: RequireKernelIsolation requested but the isolation
	// axis plus runtime gives no kernel boundary.
	ErrKernelIsolationUnmet = errors.New("kernel isolation unmet")
	// ErrInvalidAxes: an axis value lies outside the curated matrix.
	ErrInvalidAxes = errors.New("invalid sandbox axes")
)

// ProvisioningIsPinned reports whether the provisioning method comes from a
// pinned source. Only nix is pinned on this axis; images are pinned separately
// (the pin lives in the image reference, not the provisioning method).
func ProvisioningIsPinned(p types.ProvisioningMethod) bool {
	return p == types.ProvisionNix
}

// IsolationHidesHost reports whether the isolation axis can hide host tools.
// Host-tools-unreachable ALSO depends on closure-only store binds, no
// host-$HOME bind, and (docker) a pinned image; that bind discipline is enforced
// by the isolators. This classifier only gates the isolation axis:
// passthrough=true forfeits the guarantee for bwrap.
func IsolationHidesHost(i types.IsolationMethod, passthrough bool) bool {
	switch i {
	case types.IsolationBwrap:
		return !passthrough
	case types.IsolationDocker:
		return true
	default: // IsolationNone
		return false
	}
}

// EgressIsRestricted reports whether the network method restricts egress.
// Network=host shares the host network unrestricted and does not qualify.
func EgressIsRestricted(n types.NetworkMethod) bool {
	return n == types.NetworkNone || n == types.NetworkProxy
}

// KernelIsolated reports whether the isolation method plus runtime gives a
// kernel boundary. Only docker with a sandboxed runtime (runsc/gVisor) qualifies
// here; bwrap, isolation=none, and default runc are attack-surface reduction,
// not a kernel boundary. Operator pods reach kernel isolation through a
// sandboxed runtimeClass, classified by the operator.
func KernelIsolated(i types.IsolationMethod, runtime string) bool {
	return i == types.IsolationDocker && (runtime == "runsc" || runtime == "gvisor")
}

// ValidAxes reports whether each axis value lies within the curated matrix
// ({none,bwrap,docker} × {none,nix,command} × {host,none,proxy}), returning a
// wrapped ErrInvalidAxes naming the offending axis otherwise.
func ValidAxes(iso types.IsolationMethod, prov types.ProvisioningMethod, net types.NetworkMethod) error {
	switch iso {
	case types.IsolationNone, types.IsolationBwrap, types.IsolationDocker:
	default:
		return fmt.Errorf("isolation %q is not one of none/bwrap/docker: %w", iso, ErrInvalidAxes)
	}
	switch prov {
	case types.ProvisionNone, types.ProvisionNix, types.ProvisionCommand:
	default:
		return fmt.Errorf("provisioning %q is not one of none/nix/command: %w", prov, ErrInvalidAxes)
	}
	switch net {
	case types.NetworkHost, types.NetworkNone, types.NetworkProxy:
	default:
		return fmt.Errorf("network %q is not one of host/none/proxy: %w", net, ErrInvalidAxes)
	}
	return nil
}

// EnforcePolicy validates the requested axes against the demanded guarantees,
// failing CLOSED: it returns a distinct named error (wrapped with context) for
// the first unmet guarantee and never silently downgrades. It first rejects any
// axis value outside the curated matrix.
func EnforcePolicy(
	iso types.IsolationMethod,
	prov types.ProvisioningMethod,
	net types.NetworkMethod,
	passthrough bool,
	runtime string,
	pol types.SandboxPolicy,
) error {
	if err := ValidAxes(iso, prov, net); err != nil {
		return err
	}
	if pol.RequirePinnedProvisioning && !ProvisioningIsPinned(prov) {
		return fmt.Errorf("provisioning %q is not pinned (require-pinned-provisioning needs nix or a pinned image): %w", prov, ErrPinnedProvisioningUnmet)
	}
	if pol.RequireHostToolsUnreachable && !IsolationHidesHost(iso, passthrough) {
		return fmt.Errorf("isolation %q (passthrough=%t) leaves host tools reachable: %w", iso, passthrough, ErrHostToolsReachable)
	}
	if pol.RequireEgressRestricted && !EgressIsRestricted(net) {
		return fmt.Errorf("network %q does not restrict egress (require-egress-restricted needs none or proxy): %w", net, ErrEgressUnrestricted)
	}
	if pol.RequireKernelIsolation && !KernelIsolated(iso, runtime) {
		return fmt.Errorf("isolation %q runtime %q gives no kernel boundary (require-kernel-isolation needs docker --runtime runsc/gVisor or a sandboxed runtimeClass): %w", iso, runtime, ErrKernelIsolationUnmet)
	}
	return nil
}
