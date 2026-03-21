package resolver

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ResolveToolchain resolves toolchain config from either inline or a named reference.
func ResolveToolchain(ctx context.Context, c client.Client, namespace, toolchainRef string, inline *agentv1alpha1.ToolchainSpec) (*agentv1alpha1.ToolchainSpec, error) {
	if inline != nil {
		return inline, nil
	}

	if toolchainRef == "" {
		return nil, nil
	}

	var tc agentv1alpha1.AgentToolchain
	if err := c.Get(ctx, types.NamespacedName{Name: toolchainRef, Namespace: namespace}, &tc); err != nil {
		return nil, err
	}

	return &tc.Spec, nil
}
