package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/resolver"
)

// PolicySource lists the AgentPolicies of a namespace.
type PolicySource interface {
	ListAgentPolicies(ctx context.Context, namespace string) ([]agentv1alpha1.AgentPolicy, error)
}

// ReferenceReader reads the harness, provider and toolchain an AgentRun names.
type ReferenceReader interface {
	resolver.HarnessReader
	resolver.ToolchainReader
	GetAgentProvider(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentProvider, error)
}

// Catalog reports which variant of each axis the operator registers.
type Catalog interface {
	HasHarness(agent agentv1alpha1.AgentType) bool
	HasWorkspace(wsType agentv1alpha1.WorkspaceType) bool
	HasToolchain(tcType agentv1alpha1.ToolchainType) bool
}

// namespacePolicy reads the AgentPolicy of a namespace through source.
func namespacePolicy(ctx context.Context, source PolicySource, namespace string) (*agentv1alpha1.AgentPolicy, error) {
	policies, err := source.ListAgentPolicies(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("list AgentPolicy in namespace %q: %w", namespace, err)
	}
	return selectPolicy(namespace, policies)
}

// selectPolicy returns the single AgentPolicy of a namespace, nil when it has
// none, and an error when it has more than one.
func selectPolicy(namespace string, policies []agentv1alpha1.AgentPolicy) (*agentv1alpha1.AgentPolicy, error) {
	switch len(policies) {
	case 0:
		return nil, nil
	case 1:
		return policies[0].DeepCopy(), nil
	default:
		return nil, fmt.Errorf("namespace %q has %d AgentPolicies; exactly one is supported", namespace, len(policies))
	}
}
