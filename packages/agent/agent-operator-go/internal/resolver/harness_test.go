package resolver

import (
	"context"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestResolveHarness_InlineSpecTakesPriorityOverRef(t *testing.T) {
	stored := testutil.NewAgentHarness("default", "stored",
		testutil.WithDefaultImage("stored-image"))
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(stored).
		Build()
	inline := &agentv1alpha1.AgentSpec{Type: agentv1alpha1.AgentTypeOpencode}

	harness, err := ResolveHarness(context.Background(), c, "default", "stored", inline)

	if err != nil {
		t.Fatalf("ResolveHarness() error = %v", err)
	}
	if harness == nil {
		t.Fatal("ResolveHarness() = nil, want the inline spec")
	}
	if harness.Spec.Type != agentv1alpha1.AgentTypeOpencode {
		t.Errorf("resolved type = %q, want the inline %q",
			harness.Spec.Type, agentv1alpha1.AgentTypeOpencode)
	}
	if harness.Spec.DefaultImage == "stored-image" {
		t.Error("inline spec must not be overwritten by the referenced harness")
	}
}

func TestResolveHarness_ResolvesAReference(t *testing.T) {
	stored := testutil.NewAgentHarness("default", "stored",
		testutil.WithDefaultImage("stored-image"))
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(stored).
		Build()

	harness, err := ResolveHarness(context.Background(), c, "default", "stored", nil)

	if err != nil {
		t.Fatalf("ResolveHarness() error = %v", err)
	}
	if harness == nil {
		t.Fatal("ResolveHarness() = nil, want the stored harness")
	}
	if harness.Spec.DefaultImage != "stored-image" {
		t.Errorf("resolved image = %q, want %q", harness.Spec.DefaultImage, "stored-image")
	}
}

// Neither an inline spec nor a reference means the run carries no harness, which
// is a valid state rather than an error.
func TestResolveHarness_ReturnsNothingWithoutInlineOrRef(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testutil.NewScheme()).Build()

	harness, err := ResolveHarness(context.Background(), c, "default", "", nil)

	if err != nil {
		t.Fatalf("ResolveHarness() error = %v, want nil", err)
	}
	if harness != nil {
		t.Errorf("ResolveHarness() = %v, want nil", harness)
	}
}

func TestResolveHarness_PropagatesAMissingReference(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testutil.NewScheme()).Build()

	harness, err := ResolveHarness(context.Background(), c, "default", "absent", nil)

	if err == nil {
		t.Fatal("ResolveHarness() must return an error for an absent reference")
	}
	if harness != nil {
		t.Errorf("ResolveHarness() = %v, want nil alongside the error", harness)
	}
}

// The reference resolves inside the run's namespace, so an identically named
// harness elsewhere must not satisfy it.
func TestResolveHarness_DoesNotCrossNamespaces(t *testing.T) {
	stored := testutil.NewAgentHarness("other", "stored",
		testutil.WithDefaultImage("stored-image"))
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(stored).
		Build()

	if _, err := ResolveHarness(
		context.Background(), c, "default", "stored", nil,
	); err == nil {
		t.Fatal("ResolveHarness() must not resolve a harness from another namespace")
	}
}
