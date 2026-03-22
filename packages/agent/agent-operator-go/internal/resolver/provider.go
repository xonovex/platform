package resolver

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	sharedtypes "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// ResolveProvider resolves the provider configuration and returns environment variables
func ResolveProvider(ctx context.Context, c client.Client, run *agentv1alpha1.AgentRun, defaultProvider string) (map[string]string, error) {
	env := make(map[string]string)

	// Use inline provider if specified
	if run.Spec.Provider != nil {
		return resolveInlineProvider(ctx, c, run.Namespace, run.Spec.Provider)
	}

	// Use provider ref if specified
	providerRef := run.Spec.ProviderRef
	if providerRef == "" {
		providerRef = defaultProvider // from harness
	}

	if providerRef == "" {
		return env, nil
	}

	// Fetch AgentProvider
	var provider agentv1alpha1.AgentProvider
	if err := c.Get(ctx, types.NamespacedName{Name: providerRef, Namespace: run.Namespace}, &provider); err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", providerRef, err)
	}

	// Load preset env vars as defaults
	mergePresetEnv(env, provider.Spec.PresetRef, provider.Spec.AgentType)

	// Copy environment from provider spec (overrides preset)
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

	// Load preset env vars as defaults
	mergePresetEnv(env, spec.PresetRef, spec.AgentType)

	// Copy environment from inline spec (overrides preset)
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

// mergePresetEnv loads preset environment variables as defaults into env.
// Unknown presets are silently ignored for forward compatibility.
func mergePresetEnv(env map[string]string, presetRef, agentType string) {
	if presetRef == "" {
		return
	}
	at := sharedtypes.AgentType(agentType)
	if at == "" {
		at = sharedtypes.AgentClaude
	}
	preset, err := providers.GetProvider(presetRef, at)
	if err != nil {
		return // unknown preset — soft failure
	}
	for k, v := range preset.Environment {
		env[k] = v
	}
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
