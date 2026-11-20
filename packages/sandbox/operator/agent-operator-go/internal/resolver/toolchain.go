package resolver

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// ToolchainReader reads a named AgentToolchain.
type ToolchainReader interface {
	GetAgentToolchain(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentToolchain, error)
}

// ResolveToolchain resolves toolchain config from either inline or a named reference.
func ResolveToolchain(ctx context.Context, reader ToolchainReader, namespace, toolchainRef string, inline *agentv1alpha1.ToolchainSpec) (*agentv1alpha1.ToolchainSpec, error) {
	if inline != nil {
		return inline, nil
	}

	if toolchainRef == "" {
		return nil, nil
	}

	toolchain, err := reader.GetAgentToolchain(ctx, types.NamespacedName{Name: toolchainRef, Namespace: namespace})
	if err != nil {
		return nil, err
	}

	return &toolchain.Spec, nil
}
