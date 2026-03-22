package webhook

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
)

// AgentProviderWebhook implements validation for AgentProvider
type AgentProviderWebhook struct{}

var _ webhook.CustomValidator = &AgentProviderWebhook{}

func (w *AgentProviderWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&agentv1alpha1.AgentProvider{}).
		WithValidator(w).
		Complete()
}

func (w *AgentProviderWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	provider, ok := obj.(*agentv1alpha1.AgentProvider)
	if !ok {
		return nil, fmt.Errorf("expected AgentProvider, got %T", obj)
	}
	return w.validate(provider)
}

func (w *AgentProviderWebhook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	provider, ok := newObj.(*agentv1alpha1.AgentProvider)
	if !ok {
		return nil, fmt.Errorf("expected AgentProvider, got %T", newObj)
	}
	return w.validate(provider)
}

func (w *AgentProviderWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

var envKeyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

var k8sNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{0,251}[a-z0-9]$|^[a-z0-9]$`)

var blockedEnvKeyPrefixes = []string{
	"LD_",
	"DYLD_",
	"PYTHONPATH",
	"RUBYOPT",
	"NODE_OPTIONS",
	"JAVA_TOOL_OPTIONS",
}

func (w *AgentProviderWebhook) validate(provider *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
	if provider.Spec.AuthTokenSecretRef != nil {
		ref := provider.Spec.AuthTokenSecretRef
		if ref.Name == "" {
			return nil, fmt.Errorf("authTokenSecretRef.name is required")
		}
		if ref.Key == "" {
			return nil, fmt.Errorf("authTokenSecretRef.key is required")
		}
		if !k8sNamePattern.MatchString(ref.Name) {
			return nil, fmt.Errorf("authTokenSecretRef.name %q is not a valid Kubernetes resource name", ref.Name)
		}
	}

	for key := range provider.Spec.Environment {
		if !envKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("environment key %q is not a valid env var name", key)
		}
		upperKey := strings.ToUpper(key)
		for _, blocked := range blockedEnvKeyPrefixes {
			if strings.HasPrefix(upperKey, blocked) {
				return nil, fmt.Errorf("environment key %q is not allowed (blocked prefix %q)", key, blocked)
			}
		}
	}

	for i, arg := range provider.Spec.CliArgs {
		if arg == "" {
			return nil, fmt.Errorf("cliArgs[%d] is empty", i)
		}
		if validator.ContainsShellMetachars(arg) {
			return nil, fmt.Errorf("cliArgs[%d] %q contains shell metacharacters", i, arg)
		}
	}

	return nil, nil
}
