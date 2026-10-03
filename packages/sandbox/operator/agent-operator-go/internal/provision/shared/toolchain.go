// Package shared is the provision axis core: the leaf-free Toolchain port. The
// neutral handoff out of provision is the pod image (a pinned-source guarantee),
// not a method-specific type. The registry that wires concrete leaves lives in the
// composition root (internal/plugins), so this core names no leaf.
package shared

// Toolchain is a pluggable provisioning strategy for the agent pod. A toolchain
// that provisions via a pre-built image returns it from Image(); Pinned reports
// whether that source is pinned (satisfies RequirePinnedProvision). The controller
// and pod-hardening resolve a Toolchain and call these methods — they never name a
// concrete toolchain.
type Toolchain interface {
	Image() string
	Pinned() bool
}
