package resolver

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ResolveConfig finds the AgentConfig in the namespace (singleton by convention)
func ResolveConfig(ctx context.Context, c client.Client, namespace string) (*agentv1alpha1.AgentConfig, error) {
	var configList agentv1alpha1.AgentConfigList
	if err := c.List(ctx, &configList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	if len(configList.Items) == 0 {
		return nil, nil
	}

	// Return the first config found (singleton by convention)
	return &configList.Items[0], nil
}
