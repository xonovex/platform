package nix

import (
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestToolchainReturnsConfiguredImage(t *testing.T) {
	toolchain := New(&agentv1alpha1.NixSpec{Image: "registry.example/agent@sha256:digest"})

	image := toolchain.Image()

	if image != "registry.example/agent@sha256:digest" {
		t.Errorf("Image() = %q, want configured image", image)
	}
}

func TestToolchainWithoutSpecReturnsEmptyImage(t *testing.T) {
	toolchain := New(nil)

	image := toolchain.Image()

	if image != "" {
		t.Errorf("Image() = %q, want empty image", image)
	}
}

func TestToolchainIsPinned(t *testing.T) {
	toolchain := New(nil)

	pinned := toolchain.Pinned()

	if !pinned {
		t.Error("Pinned() = false, want true")
	}
}
