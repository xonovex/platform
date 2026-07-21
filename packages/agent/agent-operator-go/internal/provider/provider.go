// Package provider is the provider axis: it resolves an AgentRun's model-provider
// configuration (inline or referenced) into the environment the agent container
// needs while preserving Secret references.
package provider

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	sharedtypes "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// ResolveProvider resolves the provider configuration and returns environment variables.
func ResolveProvider(ctx context.Context, c client.Client, run *agentv1alpha1.AgentRun, defaultProvider string) ([]corev1.EnvVar, error) {
	env := make(map[string]string)

	// Use inline provider if specified.
	if run.Spec.Provider != nil {
		return resolveInlineProvider(ctx, c, run.Namespace, run.Spec.Provider)
	}

	// Use provider ref if specified.
	providerRef := run.Spec.ProviderRef
	if providerRef == "" {
		providerRef = defaultProvider // from harness
	}
	if providerRef == "" {
		return nil, nil
	}

	var provider agentv1alpha1.AgentProvider
	if err := c.Get(ctx, types.NamespacedName{Name: providerRef, Namespace: run.Namespace}, &provider); err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", providerRef, err)
	}

	// Load preset env vars as defaults.
	if err := mergePresetEnv(env, provider.Spec.PresetRef, provider.Spec.AgentType); err != nil {
		return nil, err
	}

	// Copy environment from provider spec (overrides preset).
	for k, v := range provider.Spec.Environment {
		env[k] = v
	}

	var authSecretRef *agentv1alpha1.SecretKeyRef
	if provider.Spec.AuthTokenSecretRef != nil {
		if err := validateSecretKey(ctx, c, run.Namespace, provider.Spec.AuthTokenSecretRef); err != nil {
			return nil, fmt.Errorf("failed to resolve auth token: %w", err)
		}
		if _, hasBaseURL := env["ANTHROPIC_BASE_URL"]; hasBaseURL {
			authSecretRef = provider.Spec.AuthTokenSecretRef
		}
	}

	return environmentVariables(env, authSecretRef), nil
}

func resolveInlineProvider(ctx context.Context, c client.Client, namespace string, spec *agentv1alpha1.ProviderSpec) ([]corev1.EnvVar, error) {
	env := make(map[string]string)

	if err := mergePresetEnv(env, spec.PresetRef, spec.AgentType); err != nil {
		return nil, err
	}

	for k, v := range spec.Environment {
		env[k] = v
	}

	var authSecretRef *agentv1alpha1.SecretKeyRef
	if spec.AuthSecretRef != nil {
		if err := validateSecretKey(ctx, c, namespace, spec.AuthSecretRef); err != nil {
			return nil, fmt.Errorf("failed to resolve inline auth token: %w", err)
		}
		if _, hasBaseURL := env["ANTHROPIC_BASE_URL"]; hasBaseURL {
			authSecretRef = spec.AuthSecretRef
		}
	}

	return environmentVariables(env, authSecretRef), nil
}

func environmentVariables(values map[string]string, authSecretRef *agentv1alpha1.SecretKeyRef) []corev1.EnvVar {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	env := make([]corev1.EnvVar, 0, len(values)+1)
	for _, key := range keys {
		env = append(env, corev1.EnvVar{Name: key, Value: values[key]})
	}
	if authSecretRef != nil {
		env = append(env, corev1.EnvVar{
			Name: "ANTHROPIC_AUTH_TOKEN",
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: authSecretRef.Name},
				Key:                  authSecretRef.Key,
			}},
		})
	}
	return env
}

// mergePresetEnv loads preset environment variables as defaults into env.
func mergePresetEnv(env map[string]string, presetRef, agentType string) error {
	if presetRef == "" {
		return nil
	}
	at := sharedtypes.AgentType(agentType)
	if at == "" {
		at = sharedtypes.AgentClaude
	}
	preset, err := providers.GetProvider(presetRef, at)
	if err != nil {
		return fmt.Errorf("unknown provider preset %q for agent type %q: %w", presetRef, at, err)
	}
	for k, v := range preset.Environment {
		env[k] = v
	}
	return nil
}

func validateSecretKey(ctx context.Context, c client.Client, namespace string, ref *agentv1alpha1.SecretKeyRef) error {
	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace}, &secret); err != nil {
		return fmt.Errorf("secret %s not found: %w", ref.Name, err)
	}

	if _, ok := secret.Data[ref.Key]; !ok {
		return fmt.Errorf("key %s not found in secret %s", ref.Key, ref.Name)
	}
	return nil
}
