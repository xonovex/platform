package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
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

func (w *AgentProviderWebhook) validate(provider *agentv1alpha1.AgentProvider) (admission.Warnings, error) {
	if len(provider.Spec.AgentTypes) == 0 {
		return nil, fmt.Errorf("at least one agent type must be specified")
	}

	for _, at := range provider.Spec.AgentTypes {
		if at != agentv1alpha1.AgentTypeClaude && at != agentv1alpha1.AgentTypeOpencode {
			return nil, fmt.Errorf("invalid agent type: %s", at)
		}
	}

	return nil, nil
}
