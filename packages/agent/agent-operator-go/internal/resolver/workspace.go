package resolver

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ResolveWorkspace fetches an AgentWorkspace by name in the given namespace
func ResolveWorkspace(ctx context.Context, c client.Client, namespace, name string) (*agentv1alpha1.AgentWorkspace, error) {
	var ws agentv1alpha1.AgentWorkspace
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ws); err != nil {
		return nil, fmt.Errorf("failed to get workspace %s: %w", name, err)
	}
	return &ws, nil
}
