package builder

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// Toolchain is a pluggable provisioning strategy for the agent pod. A toolchain
// that provisions via a pre-built image returns it from Image(); Pinned reports
// whether that source is pinned (satisfies RequirePinnedProvision). The
// controller and pod-hardening resolve a Toolchain and call these methods — they
// never name a concrete toolchain.
type Toolchain interface {
	// Image returns the pre-built pod image for this toolchain, or "" if it does
	// not provision via an image.
	Image() string
	// Pinned reports whether the toolchain provisions from a pinned source.
	Pinned() bool
}

// ToolchainFactory builds a Toolchain from its spec.
type ToolchainFactory func(*agentv1alpha1.ToolchainSpec) Toolchain

// toolchainFactories is the registry of toolchain resolvers keyed by type. It is
// an immutable lookup table — adding a toolchain is one entry here plus a
// Toolchain implementation, not edits across the controller and pod-hardening.
var toolchainFactories = map[agentv1alpha1.ToolchainType]ToolchainFactory{
	agentv1alpha1.ToolchainTypeNix: func(tc *agentv1alpha1.ToolchainSpec) Toolchain { return nixToolchain{tc.Nix} },
}

// ResolveToolchain returns the Toolchain for a spec, or nil if the spec is nil or
// its type is not registered.
func ResolveToolchain(tc *agentv1alpha1.ToolchainSpec) Toolchain {
	if tc == nil {
		return nil
	}
	factory, ok := toolchainFactories[tc.Type]
	if !ok {
		return nil
	}
	return factory(tc)
}

// nixToolchain provisions via a pre-built, digest-pinned nix OCI image — the same
// content-addressed closure the CLI resolves. There is no per-pod nix install.
type nixToolchain struct{ nix *agentv1alpha1.NixSpec }

func (n nixToolchain) Image() string {
	if n.nix == nil {
		return ""
	}
	return n.nix.Image
}

func (n nixToolchain) Pinned() bool { return true }
