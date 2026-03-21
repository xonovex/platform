package resolver

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ResolveHarness resolves an AgentHarness from either inline config or a named reference.
// Inline config takes priority. Returns nil if neither is provided.
func ResolveHarness(ctx context.Context, c client.Client, namespace, harnessRef string, inline *agentv1alpha1.AgentSpec) (*agentv1alpha1.AgentHarness, error) {
	if inline != nil {
		return &agentv1alpha1.AgentHarness{Spec: *inline}, nil
	}

	if harnessRef == "" {
		return nil, nil
	}

	var harness agentv1alpha1.AgentHarness
	if err := c.Get(ctx, types.NamespacedName{Name: harnessRef, Namespace: namespace}, &harness); err != nil {
		return nil, err
	}

	return &harness, nil
}
