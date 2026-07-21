package webhook

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/providers"
	sharedtypes "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
)

// AgentProviderWebhook implements validation for AgentProvider
type AgentProviderWebhook struct{}

var _ admission.Validator[*agentv1alpha1.AgentProvider] = &AgentProviderWebhook{}

func (w *AgentProviderWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &agentv1alpha1.AgentProvider{}).
		WithValidator(w).
		Complete()
}

func (w *AgentProviderWebhook) ValidateCreate(_ context.Context, provider *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
	return w.validate(provider)
}

func (w *AgentProviderWebhook) ValidateUpdate(_ context.Context, _ *agentv1alpha1.AgentProvider, newObj *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
	return w.validate(newObj)
}

func (w *AgentProviderWebhook) ValidateDelete(_ context.Context, _ *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
	return nil, nil
}

var envKeyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

var k8sNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{0,251}[a-z0-9]$|^[a-z0-9]$`)

func (w *AgentProviderWebhook) validate(provider *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
	err := validateProviderConfig(
		provider.Spec.PresetRef,
		provider.Spec.AgentType,
		provider.Spec.AuthTokenSecretRef,
		provider.Spec.AuthTokenEnv,
		provider.Spec.Environment,
		provider.Spec.CliArgs,
	)
	return nil, err
}

func validateProviderConfig(presetRef, agentType string, secretRef *agentv1alpha1.SecretKeyRef, authTokenEnv string, environment map[string]string, cliArgs []string) error {
	presetAuthTokenEnv := ""
	if presetRef != "" {
		at := sharedtypes.AgentType(agentType)
		if at == "" {
			at = sharedtypes.AgentClaude
		}
		preset, err := providers.GetPortableProvider(presetRef, at)
		if err != nil {
			return fmt.Errorf("presetRef %q is not a portable provider preset for agent type %q: %w", presetRef, at, err)
		}
		presetAuthTokenEnv = preset.CredentialTargetEnv
	}

	if secretRef != nil {
		ref := secretRef
		if ref.Name == "" {
			return fmt.Errorf("authTokenSecretRef.name is required")
		}
		if ref.Key == "" {
			return fmt.Errorf("authTokenSecretRef.key is required")
		}
		if !k8sNamePattern.MatchString(ref.Name) {
			return fmt.Errorf("authTokenSecretRef.name %q is not a valid Kubernetes resource name", ref.Name)
		}
		effectiveAuthTokenEnv := authTokenEnv
		if effectiveAuthTokenEnv == "" {
			effectiveAuthTokenEnv = presetAuthTokenEnv
		}
		if effectiveAuthTokenEnv == "" {
			return fmt.Errorf("authTokenEnv is required when authTokenSecretRef is configured")
		}
		if err := validateEnvironmentKey(effectiveAuthTokenEnv); err != nil {
			return fmt.Errorf("authTokenEnv: %w", err)
		}
		if _, exists := environment[effectiveAuthTokenEnv]; exists {
			return fmt.Errorf("environment key %q conflicts with authTokenSecretRef", effectiveAuthTokenEnv)
		}
	} else if authTokenEnv != "" {
		return fmt.Errorf("authTokenEnv requires authTokenSecretRef")
	}

	for key := range environment {
		if err := validateEnvironmentKey(key); err != nil {
			return fmt.Errorf("environment key %q: %w", key, err)
		}
	}

	for i, arg := range cliArgs {
		if arg == "" {
			return fmt.Errorf("cliArgs[%d] is empty", i)
		}
		if shell.ContainsMetachars(arg) {
			return fmt.Errorf("cliArgs[%d] %q contains shell metacharacters", i, arg)
		}
	}

	return nil
}

func validateEnvironmentKey(key string) error {
	if !envKeyPattern.MatchString(key) {
		return fmt.Errorf("is not a valid env var name")
	}
	upperKey := strings.ToUpper(key)
	for _, blocked := range []string{"LD_", "DYLD_", "PYTHONPATH", "RUBYOPT", "NODE_OPTIONS", "JAVA_TOOL_OPTIONS"} {
		if strings.HasPrefix(upperKey, blocked) {
			return fmt.Errorf("is not allowed (blocked prefix %q)", blocked)
		}
	}
	return nil
}
