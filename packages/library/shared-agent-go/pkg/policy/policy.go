package policy

import (
	"errors"
	"fmt"
)

// Policy errors. EnforcePolicy wraps one of these (with context) for the first
// unmet guarantee; callers match with errors.Is to react to a specific failure.
var (
	// ErrPinnedProvisionUnmet: RequirePinnedProvision requested but the resolved
	// provisioner is not pinned.
	ErrPinnedProvisionUnmet = errors.New("pinned provision unmet")
	// ErrHostToolsReachable: RequireHostToolsUnreachable requested but the resolved
	// isolator leaves host tools reachable.
	ErrHostToolsReachable = errors.New("host tools reachable")
	// ErrEgressUnrestricted: RequireEgressRestricted requested but the network
	// method does not restrict egress.
	ErrEgressUnrestricted = errors.New("egress unrestricted")
	// ErrKernelIsolationUnmet: RequireKernelIsolation requested but the resolved
	// isolator plus runtime gives no kernel boundary.
	ErrKernelIsolationUnmet = errors.New("kernel isolation unmet")
)

// Capabilities are the guarantees a SELECTED sandbox actually provides, computed
// by the caller from the resolved isolator/provisioner/network. The policy engine
// is method-agnostic: it never names a concrete isolator or provisioner, so a new
// plugin (a pinned-image provisioner, a kernel-isolating isolator) declares its
// own capability and needs zero changes here.
type Capabilities struct {
	// Pinned: the provisioner resolves from a pinned source (Provisioner.Pinned()).
	Pinned bool
	// HostToolsUnreachable: the isolator keeps host tools off PATH and not
	// bind-reachable for this request (Isolator.HidesHost(passthrough, image)).
	HostToolsUnreachable bool
	// EgressRestricted: the network method restricts egress
	// (network.EgressRestricted).
	EgressRestricted bool
	// KernelIsolated: the isolator plus runtime gives a kernel boundary
	// (Isolator.KernelIsolated(runtime)).
	KernelIsolated bool
}

// SandboxPolicy expresses the isolation guarantees the caller demands of the
// selected sandbox, independent of which axes were requested. Each guarantee is
// independently requestable and is enforced fail-closed (see EnforcePolicy):
// the engine refuses to run when a requested guarantee cannot be established,
// never silently degrades.
type SandboxPolicy struct {
	// RequirePinnedProvision mandates provisioning from a pinned source
	// (nix closure or a pinned image), enforced at resolve via
	// --frozen / --no-write-lock-file against a committed lock (fail closed).
	RequirePinnedProvision bool
	// RequireHostToolsUnreachable mandates that host tools are off PATH AND not
	// bind-reachable. Conditioned on closure-only store binds, no host-$HOME bind,
	// and (docker) a pinned image.
	RequireHostToolsUnreachable bool
	// RequireEgressRestricted mandates Network=none, or Network=managed on an isolator whose egress transport forces every connection through the allowlist; host is unrestricted.
	RequireEgressRestricted bool
	// RequireKernelIsolation mandates a kernel boundary: docker --runtime
	// runsc/gVisor, or a pod with a sandboxed runtimeClass (gVisor/Kata/kata-cc).
	// NOT satisfied by bwrap or default runc.
	RequireKernelIsolation bool
}

// Guarantee names one isolation guarantee a policy can demand.
type Guarantee string

// The guarantees of a sandbox, in the order EnforcePolicy checks them.
const (
	GuaranteePinnedProvision      Guarantee = "pinned-provision"
	GuaranteeHostToolsUnreachable Guarantee = "host-tools-unreachable"
	GuaranteeEgressRestricted     Guarantee = "egress-restricted"
	GuaranteeKernelIsolation      Guarantee = "kernel-isolation"
)

// Outcome is the evaluation of one guarantee: whether the policy demands it,
// whether the selected sandbox provides it, and why a demand is unmet.
type Outcome struct {
	Guarantee Guarantee
	// Required reports that the policy demands the guarantee.
	Required bool
	// Provided reports that the selected sandbox establishes the guarantee.
	Provided bool
	// Err is the named policy error of a demand that is not provided, nil
	// otherwise. It carries the reason.
	Err error
}

// Met reports that the policy holds for the guarantee: it is provided, or the
// policy does not demand it.
func (o Outcome) Met() bool { return !o.Required || o.Provided }

// Evaluate reports every guarantee against the capabilities, in the order
// EnforcePolicy checks them. It reports a demand that is not provided in Err and
// never stops at the first one.
func Evaluate(caps Capabilities, pol SandboxPolicy) []Outcome {
	outcomes := []Outcome{
		{Guarantee: GuaranteePinnedProvision, Required: pol.RequirePinnedProvision, Provided: caps.Pinned,
			Err: fmt.Errorf("provisioning is not pinned (require-pinned-provision needs nix or a pinned image): %w", ErrPinnedProvisionUnmet)},
		{Guarantee: GuaranteeHostToolsUnreachable, Required: pol.RequireHostToolsUnreachable, Provided: caps.HostToolsUnreachable,
			Err: fmt.Errorf("the selected isolation leaves host tools reachable: %w", ErrHostToolsReachable)},
		{Guarantee: GuaranteeEgressRestricted, Required: pol.RequireEgressRestricted, Provided: caps.EgressRestricted,
			Err: fmt.Errorf("the network method does not restrict egress (require-egress-restricted needs none, or managed on an isolator that forces the allowlist): %w", ErrEgressUnrestricted)},
		{Guarantee: GuaranteeKernelIsolation, Required: pol.RequireKernelIsolation, Provided: caps.KernelIsolated,
			Err: fmt.Errorf("the selected isolation gives no kernel boundary (require-kernel-isolation needs docker --runtime runsc/gVisor or a sandboxed runtimeClass): %w", ErrKernelIsolationUnmet)},
	}
	for index := range outcomes {
		if outcomes[index].Met() {
			outcomes[index].Err = nil
		}
	}
	return outcomes
}

// EnforcePolicy validates the provided capabilities against the demanded
// guarantees, failing CLOSED: it returns a distinct named error (wrapped with
// context) for the first unmet guarantee and never silently downgrades.
func EnforcePolicy(caps Capabilities, pol SandboxPolicy) error {
	for _, outcome := range Evaluate(caps, pol) {
		if outcome.Err != nil {
			return outcome.Err
		}
	}
	return nil
}
