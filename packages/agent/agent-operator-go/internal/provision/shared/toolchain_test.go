package shared

import (
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestResolveToolchain(t *testing.T) {
	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Image: "ghcr.io/xonovex/agent@sha256:abc"},
	}
	tcl := ResolveToolchain(tc)
	if tcl == nil || tcl.Image() != "ghcr.io/xonovex/agent@sha256:abc" {
		t.Fatalf("ResolveToolchain(nix).Image() = %v, want the pre-built image", tcl)
	}
	if !tcl.Pinned() {
		t.Error("nix toolchain must report Pinned()=true")
	}
	if ResolveToolchain(nil) != nil {
		t.Error("ResolveToolchain(nil) must be nil")
	}
	if ResolveToolchain(&agentv1alpha1.ToolchainSpec{Type: "bogus"}) != nil {
		t.Error("ResolveToolchain(unknown type) must be nil")
	}
}
