package resolver

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// ResolveProvider resolves the provider configuration and returns environment variables
func ResolveProvider(ctx context.Context, c client.Client, run *agentv1alpha1.AgentRun, config *agentv1alpha1.AgentConfig) (map[string]string, error) {
	env := make(map[string]string)

	// Use inline provider if specified
	if run.Spec.Provider != nil {
		return resolveInlineProvider(ctx, c, run.Namespace, run.Spec.Provider)
	}

	// Use provider ref if specified
	providerRef := run.Spec.ProviderRef
	if providerRef == "" && config != nil {
		// Fall back to default provider from AgentConfig
		if defaultRef, ok := config.Spec.DefaultProviders[run.Spec.Agent]; ok {
			providerRef = defaultRef
		}
	}

	if providerRef == "" {
		return env, nil
	}

	// Fetch AgentProvider
	var provider agentv1alpha1.AgentProvider
	if err := c.Get(ctx, types.NamespacedName{Name: providerRef, Namespace: run.Namespace}, &provider); err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", providerRef, err)
	}

	// Copy environment from provider spec
	for k, v := range provider.Spec.Environment {
		env[k] = v
	}

	// Resolve auth token from secret
	if provider.Spec.AuthTokenSecretRef != nil {
		token, err := getSecretValue(ctx, c, run.Namespace, provider.Spec.AuthTokenSecretRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve auth token: %w", err)
		}
		// Inject token as ANTHROPIC_AUTH_TOKEN for Anthropic-compatible providers
		if _, hasBaseURL := env["ANTHROPIC_BASE_URL"]; hasBaseURL {
			env["ANTHROPIC_AUTH_TOKEN"] = token
		}
	}

	return env, nil
}

func resolveInlineProvider(ctx context.Context, c client.Client, namespace string, spec *agentv1alpha1.ProviderSpec) (map[string]string, error) {
	env := make(map[string]string)

	for k, v := range spec.Environment {
		env[k] = v
	}

	if spec.AuthSecretRef != nil {
		token, err := getSecretValue(ctx, c, namespace, spec.AuthSecretRef)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve inline auth token: %w", err)
		}
		if _, hasBaseURL := env["ANTHROPIC_BASE_URL"]; hasBaseURL {
			env["ANTHROPIC_AUTH_TOKEN"] = token
		}
	}

	return env, nil
}

func getSecretValue(ctx context.Context, c client.Client, namespace string, ref *agentv1alpha1.SecretKeyRef) (string, error) {
	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace}, &secret); err != nil {
		return "", fmt.Errorf("secret %s not found: %w", ref.Name, err)
	}

	value, ok := secret.Data[ref.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", ref.Key, ref.Name)
	}

	return string(value), nil
}
