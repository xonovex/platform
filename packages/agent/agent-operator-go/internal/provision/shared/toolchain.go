// Package shared is the provision axis core: the Toolchain port plus the
// type-keyed registry. The neutral handoff out of provision is the pod image (a
// pinned-source guarantee), not a method-specific type. The registry is
// package-level (lazy DI is the deferred follow-up).
package shared

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/provision/nix"
)

// Toolchain is a pluggable provisioning strategy for the agent pod. A toolchain
// that provisions via a pre-built image returns it from Image(); Pinned reports
// whether that source is pinned (satisfies RequirePinnedProvision). The controller
// and pod-hardening resolve a Toolchain and call these methods — they never name a
// concrete toolchain.
type Toolchain interface {
	Image() string
	Pinned() bool
}

// Factory builds a Toolchain from its spec.
type Factory func(*agentv1alpha1.ToolchainSpec) Toolchain

// toolchainFactories is the registry of toolchain resolvers keyed by type. Adding a
// toolchain is one entry here plus a leaf, not edits across the controller and
// pod-hardening.
var toolchainFactories = map[agentv1alpha1.ToolchainType]Factory{
	agentv1alpha1.ToolchainTypeNix: func(tc *agentv1alpha1.ToolchainSpec) Toolchain { return nix.New(tc.Nix) },
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
