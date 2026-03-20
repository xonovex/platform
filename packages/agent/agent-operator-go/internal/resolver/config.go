package resolver

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ResolveConfig looks up an AgentConfig by name in the given namespace.
// Returns nil if configRef is empty (no config referenced).
func ResolveConfig(ctx context.Context, c client.Client, namespace, configRef string) (*agentv1alpha1.AgentConfig, error) {
	if configRef == "" {
		return nil, nil
	}

	var config agentv1alpha1.AgentConfig
	if err := c.Get(ctx, types.NamespacedName{Name: configRef, Namespace: namespace}, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
