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

// ResolvedProvider contains the environment and command arguments contributed by a model provider.
type ResolvedProvider struct {
	Environment []corev1.EnvVar
	CliArgs     []string
}

// ResolveProvider resolves the provider configuration for an agent execution.
func ResolveProvider(ctx context.Context, c client.Client, run *agentv1alpha1.AgentRun, defaultProvider string) (ResolvedProvider, error) {
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
		return ResolvedProvider{}, nil
	}

	var provider agentv1alpha1.AgentProvider
	if err := c.Get(ctx, types.NamespacedName{Name: providerRef, Namespace: run.Namespace}, &provider); err != nil {
		return ResolvedProvider{}, fmt.Errorf("failed to get provider %s: %w", providerRef, err)
	}

	preset, err := mergePreset(env, provider.Spec.PresetRef, provider.Spec.AgentType)
	if err != nil {
		return ResolvedProvider{}, err
	}
	cliArgs := append(preset.cliArgs, provider.Spec.CliArgs...)

	// Copy environment from provider spec (overrides preset).
	for k, v := range provider.Spec.Environment {
		env[k] = v
	}

	authTokenEnv := provider.Spec.AuthTokenEnv
	if authTokenEnv == "" {
		authTokenEnv = preset.authTokenEnv
	}
	if err := validateAuthToken(ctx, c, run.Namespace, authTokenEnv, provider.Spec.AuthTokenSecretRef, env); err != nil {
		return ResolvedProvider{}, fmt.Errorf("failed to resolve auth token: %w", err)
	}

	return ResolvedProvider{Environment: environmentVariables(env, authTokenEnv, provider.Spec.AuthTokenSecretRef), CliArgs: cliArgs}, nil
}

func resolveInlineProvider(ctx context.Context, c client.Client, namespace string, spec *agentv1alpha1.ProviderSpec) (ResolvedProvider, error) {
	env := make(map[string]string)

	preset, err := mergePreset(env, spec.PresetRef, spec.AgentType)
	if err != nil {
		return ResolvedProvider{}, err
	}
	cliArgs := append(preset.cliArgs, spec.CliArgs...)

	for k, v := range spec.Environment {
		env[k] = v
	}

	authTokenEnv := spec.AuthTokenEnv
	if authTokenEnv == "" {
		authTokenEnv = preset.authTokenEnv
	}
	if err := validateAuthToken(ctx, c, namespace, authTokenEnv, spec.AuthSecretRef, env); err != nil {
		return ResolvedProvider{}, fmt.Errorf("failed to resolve inline auth token: %w", err)
	}

	return ResolvedProvider{Environment: environmentVariables(env, authTokenEnv, spec.AuthSecretRef), CliArgs: cliArgs}, nil
}

func environmentVariables(values map[string]string, authTokenEnv string, authSecretRef *agentv1alpha1.SecretKeyRef) []corev1.EnvVar {
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
			Name: authTokenEnv,
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: authSecretRef.Name},
				Key:                  authSecretRef.Key,
			}},
		})
	}
	return env
}

type presetConfig struct {
	authTokenEnv string
	cliArgs      []string
}

func mergePreset(env map[string]string, presetRef, agentType string) (presetConfig, error) {
	if presetRef == "" {
		return presetConfig{}, nil
	}
	at := sharedtypes.AgentType(agentType)
	if at == "" {
		at = sharedtypes.AgentClaude
	}
	preset, err := providers.GetPortableProvider(presetRef, at)
	if err != nil {
		return presetConfig{}, fmt.Errorf("provider preset %q is unavailable for agent type %q: %w", presetRef, at, err)
	}
	for k, v := range preset.Environment {
		env[k] = v
	}
	return presetConfig{
		authTokenEnv: preset.CredentialTargetEnv,
		cliArgs:      append([]string{}, preset.CliArgs...),
	}, nil
}

func validateAuthToken(ctx context.Context, c client.Client, namespace, authTokenEnv string, ref *agentv1alpha1.SecretKeyRef, environment map[string]string) error {
	if ref == nil {
		return nil
	}
	if authTokenEnv == "" {
		return fmt.Errorf("authTokenEnv is required when an auth token Secret is configured")
	}
	if _, exists := environment[authTokenEnv]; exists {
		return fmt.Errorf("environment variable %q conflicts with the auth token Secret", authTokenEnv)
	}
	return validateSecretKey(ctx, c, namespace, ref)
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
