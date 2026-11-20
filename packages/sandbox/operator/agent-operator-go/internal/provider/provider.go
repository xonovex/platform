// Package provider is the provider axis: it resolves the model provider of an
// AgentRun (inline or referenced) into the route of the shared plan, and renders
// the environment of that route with Secret references, never Secret values.
package provider

import (
	"context"
	"fmt"
	"maps"
	"slices"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agentcmd"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	sharedtypes "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// Route is the model provider of a run and the Secret that holds its
// credential. Provider is nil when the run names no provider; AuthSecret is nil
// when the provider takes no credential.
type Route struct {
	Provider   *sharedtypes.ModelProvider
	AuthSecret *agentv1alpha1.SecretKeyRef
}

// Reader reads the AgentProvider and the Secret a provider resolution names.
type Reader interface {
	GetAgentProvider(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentProvider, error)
	GetSecret(ctx context.Context, key types.NamespacedName) (*corev1.Secret, error)
}

// ResolveProvider resolves the provider route of a run: the inline provider,
// else the AgentProvider the run names, else defaultProvider. A presetRef names
// a portable provider of presets, the catalog the operator loaded from its
// providers directory.
func ResolveProvider(ctx context.Context, reader Reader, presets providers.Catalog, run *agentv1alpha1.AgentRun, defaultProvider string) (Route, error) {
	if run.Spec.Provider != nil {
		spec := run.Spec.Provider
		route, err := buildRoute(presets, providerFields{
			name: spec.PresetRef, presetRef: spec.PresetRef, agentType: spec.AgentType,
			authSecret: spec.AuthSecretRef, authTokenEnv: spec.AuthTokenEnv, environment: spec.Environment, cliArgs: spec.CliArgs,
		})
		if err != nil {
			return Route{}, err
		}
		if err := validateAuthToken(ctx, reader, run.Namespace, route); err != nil {
			return Route{}, fmt.Errorf("failed to resolve inline auth token: %w", err)
		}
		return route, nil
	}

	providerRef := run.Spec.ProviderRef
	if providerRef == "" {
		providerRef = defaultProvider
	}
	if providerRef == "" {
		return Route{}, nil
	}
	provider, err := reader.GetAgentProvider(ctx, types.NamespacedName{Name: providerRef, Namespace: run.Namespace})
	if err != nil {
		return Route{}, fmt.Errorf("failed to get provider %s: %w", providerRef, err)
	}
	spec := provider.Spec
	route, err := buildRoute(presets, providerFields{
		name: providerRef, displayName: spec.DisplayName, presetRef: spec.PresetRef, agentType: spec.AgentType,
		authSecret: spec.AuthTokenSecretRef, authTokenEnv: spec.AuthTokenEnv, environment: spec.Environment, cliArgs: spec.CliArgs,
	})
	if err != nil {
		return Route{}, err
	}
	if err := validateAuthToken(ctx, reader, run.Namespace, route); err != nil {
		return Route{}, fmt.Errorf("failed to resolve auth token: %w", err)
	}
	return route, nil
}

// providerFields are the fields an inline provider and an AgentProvider share.
type providerFields struct {
	name, displayName, presetRef, agentType, authTokenEnv string
	authSecret                                            *agentv1alpha1.SecretKeyRef
	environment                                           map[string]string
	cliArgs                                               []string
}

// buildRoute layers the fields of a provider over its preset: the environment
// overrides the preset environment, the arguments follow the preset arguments,
// and authTokenEnv overrides the credential target of the preset. The route
// names a credential target only when a Secret holds the credential, and never
// a host variable: the kubelet reads the credential from the Secret.
func buildRoute(presets providers.Catalog, fields providerFields) (Route, error) {
	agentType := sharedtypes.AgentType(fields.agentType)
	if agentType == "" {
		agentType = sharedtypes.AgentClaude
	}
	route := &sharedtypes.ModelProvider{Name: fields.name, DisplayName: fields.displayName, AgentType: agentType, Portable: true, Environment: map[string]string{}}
	if route.Name == "" {
		route.Name = "inline"
	}
	credentialTarget := ""
	if fields.presetRef != "" {
		preset, err := presets.ResolvePortable(fields.presetRef, agentType)
		if err != nil {
			return Route{}, fmt.Errorf("provider preset %q is unavailable for agent type %q: %w", fields.presetRef, agentType, err)
		}
		maps.Copy(route.Environment, preset.Environment)
		route.CliArgs = slices.Clone(preset.CliArgs)
		route.Slots = slices.Clone(preset.Slots)
		credentialTarget = preset.CredentialTargetEnv
	}
	maps.Copy(route.Environment, fields.environment)
	route.CliArgs = append(route.CliArgs, fields.cliArgs...)
	if fields.authTokenEnv != "" {
		credentialTarget = fields.authTokenEnv
	}
	if fields.authSecret != nil {
		route.CredentialTargetEnv = credentialTarget
	}
	return Route{Provider: route, AuthSecret: fields.authSecret}, nil
}

// Environment renders the provider environment agent reads from route: the
// planned values sorted by name, then the credential as a reference to its
// Secret.
func Environment(agent *sharedtypes.AgentConfig, route Route) []corev1.EnvVar {
	plan := agentcmd.PlanProviderEnv(agent, route.Provider)
	env := make([]corev1.EnvVar, 0, len(plan.Values)+1)
	for _, name := range slices.Sorted(maps.Keys(plan.Values)) {
		env = append(env, corev1.EnvVar{Name: name, Value: plan.Values[name]})
	}
	if plan.Credential.Env != "" && route.AuthSecret != nil {
		env = append(env, corev1.EnvVar{
			Name: plan.Credential.Env,
			ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: route.AuthSecret.Name},
				Key:                  route.AuthSecret.Key,
			}},
		})
	}
	return env
}

// validateAuthToken checks that the Secret of a route has a target variable
// that no plain variable of the route sets, and that the Secret holds its key.
func validateAuthToken(ctx context.Context, reader Reader, namespace string, route Route) error {
	if route.AuthSecret == nil {
		return nil
	}
	target := route.Provider.CredentialTargetEnv
	if target == "" {
		return fmt.Errorf("authTokenEnv is required when an auth token Secret is configured")
	}
	if _, exists := route.Provider.Environment[target]; exists {
		return fmt.Errorf("environment variable %q conflicts with the auth token Secret", target)
	}
	return validateSecretKey(ctx, reader, namespace, route.AuthSecret)
}

func validateSecretKey(ctx context.Context, reader Reader, namespace string, ref *agentv1alpha1.SecretKeyRef) error {
	secret, err := reader.GetSecret(ctx, types.NamespacedName{Name: ref.Name, Namespace: namespace})
	if err != nil {
		return fmt.Errorf("secret %s not found: %w", ref.Name, err)
	}

	if _, ok := secret.Data[ref.Key]; !ok {
		return fmt.Errorf("key %s not found in secret %s", ref.Key, ref.Name)
	}
	return nil
}
