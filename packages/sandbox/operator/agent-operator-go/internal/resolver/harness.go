package resolver

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// HarnessReader reads a named AgentHarness.
type HarnessReader interface {
	GetAgentHarness(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentHarness, error)
}

// ResolveHarness resolves an AgentHarness from either inline config or a named reference.
// Inline config takes priority. Returns nil if neither is provided.
func ResolveHarness(ctx context.Context, reader HarnessReader, namespace, harnessRef string, inline *agentv1alpha1.AgentSpec) (*agentv1alpha1.AgentHarness, error) {
	if inline != nil {
		return &agentv1alpha1.AgentHarness{Spec: *inline}, nil
	}

	if harnessRef == "" {
		return nil, nil
	}

	return reader.GetAgentHarness(ctx, types.NamespacedName{Name: harnessRef, Namespace: namespace})
}
