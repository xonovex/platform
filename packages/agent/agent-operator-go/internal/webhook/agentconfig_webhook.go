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

// AgentConfigWebhook implements validation for AgentConfig
type AgentConfigWebhook struct{}

var _ webhook.CustomValidator = &AgentConfigWebhook{}

func (w *AgentConfigWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&agentv1alpha1.AgentConfig{}).
		WithValidator(w).
		Complete()
}

func (w *AgentConfigWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	config, ok := obj.(*agentv1alpha1.AgentConfig)
	if !ok {
		return nil, fmt.Errorf("expected AgentConfig, got %T", obj)
	}
	return w.validate(config)
}

func (w *AgentConfigWebhook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	config, ok := newObj.(*agentv1alpha1.AgentConfig)
	if !ok {
		return nil, fmt.Errorf("expected AgentConfig, got %T", newObj)
	}
	return w.validate(config)
}

func (w *AgentConfigWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentConfigWebhook) validate(config *agentv1alpha1.AgentConfig) (admission.Warnings, error) {
	if config.Spec.DefaultAgent != "" {
		if config.Spec.DefaultAgent != agentv1alpha1.AgentTypeClaude && config.Spec.DefaultAgent != agentv1alpha1.AgentTypeOpencode {
			return nil, fmt.Errorf("invalid default agent type: %s", config.Spec.DefaultAgent)
		}
	}
	return nil, nil
}
