package resolver

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func newStoredToolchain(namespace, name, image string) *agentv1alpha1.AgentToolchain {
	return &agentv1alpha1.AgentToolchain{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: agentv1alpha1.ToolchainSpec{
			Type: agentv1alpha1.ToolchainTypeNix,
			Nix:  &agentv1alpha1.NixSpec{Image: image},
		},
	}
}

func TestResolveToolchain_InlineSpecTakesPriorityOverRef(t *testing.T) {
	stored := newStoredToolchain("default", "stored", "stored-image")
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(stored).
		Build()
	inline := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Image: "inline-image"},
	}

	resolved, err := ResolveToolchain(context.Background(), c, "default", "stored", inline)

	if err != nil {
		t.Fatalf("ResolveToolchain() error = %v", err)
	}
	if resolved == nil || resolved.Nix == nil {
		t.Fatalf("ResolveToolchain() = %v, want the inline spec", resolved)
	}
	if resolved.Nix.Image != "inline-image" {
		t.Errorf("resolved image = %q, want the inline %q", resolved.Nix.Image, "inline-image")
	}
}

func TestResolveToolchain_ResolvesAReference(t *testing.T) {
	stored := newStoredToolchain("default", "stored", "stored-image")
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(stored).
		Build()

	resolved, err := ResolveToolchain(context.Background(), c, "default", "stored", nil)

	if err != nil {
		t.Fatalf("ResolveToolchain() error = %v", err)
	}
	if resolved == nil || resolved.Nix == nil {
		t.Fatalf("ResolveToolchain() = %v, want the stored spec", resolved)
	}
	if resolved.Nix.Image != "stored-image" {
		t.Errorf("resolved image = %q, want %q", resolved.Nix.Image, "stored-image")
	}
}

// A run may legitimately carry no toolchain, so the absence of both inputs is
// not an error.
func TestResolveToolchain_ReturnsNothingWithoutInlineOrRef(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testutil.NewScheme()).Build()

	resolved, err := ResolveToolchain(context.Background(), c, "default", "", nil)

	if err != nil {
		t.Fatalf("ResolveToolchain() error = %v, want nil", err)
	}
	if resolved != nil {
		t.Errorf("ResolveToolchain() = %v, want nil", resolved)
	}
}

func TestResolveToolchain_PropagatesAMissingReference(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testutil.NewScheme()).Build()

	resolved, err := ResolveToolchain(context.Background(), c, "default", "absent", nil)

	if err == nil {
		t.Fatal("ResolveToolchain() must return an error for an absent reference")
	}
	if resolved != nil {
		t.Errorf("ResolveToolchain() = %v, want nil alongside the error", resolved)
	}
}

func TestResolveToolchain_DoesNotCrossNamespaces(t *testing.T) {
	stored := newStoredToolchain("other", "stored", "stored-image")
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(stored).
		Build()

	if _, err := ResolveToolchain(
		context.Background(), c, "default", "stored", nil,
	); err == nil {
		t.Fatal("ResolveToolchain() must not resolve a toolchain from another namespace")
	}
}
