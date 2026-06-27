// Package nix is the provision=nix leaf: provisioning via a pre-built,
// digest-pinned nix OCI image — the same content-addressed closure the CLI
// resolves. There is no per-pod nix install.
package nix

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// Toolchain provisions via a pre-built, digest-pinned nix OCI image.
type Toolchain struct{ nix *agentv1alpha1.NixSpec }

// New creates a nix toolchain from its spec.
func New(spec *agentv1alpha1.NixSpec) Toolchain { return Toolchain{nix: spec} }

// Image returns the pre-built pod image, or "" if unset.
func (n Toolchain) Image() string {
	if n.nix == nil {
		return ""
	}
	return n.nix.Image
}

// Pinned reports true: a nix image resolves from a flake.lock/digest-pinned source.
func (n Toolchain) Pinned() bool { return true }
