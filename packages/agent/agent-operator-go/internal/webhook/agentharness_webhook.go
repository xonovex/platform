package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/builder"
)

// AgentHarnessWebhook implements validation for AgentHarness
type AgentHarnessWebhook struct{}

var _ webhook.CustomValidator = &AgentHarnessWebhook{}

func (w *AgentHarnessWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&agentv1alpha1.AgentHarness{}).
		WithValidator(w).
		Complete()
}

func (w *AgentHarnessWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	h, ok := obj.(*agentv1alpha1.AgentHarness)
	if !ok {
		return nil, fmt.Errorf("expected AgentHarness, got %T", obj)
	}
	return w.validate(h)
}

func (w *AgentHarnessWebhook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	h, ok := newObj.(*agentv1alpha1.AgentHarness)
	if !ok {
		return nil, fmt.Errorf("expected AgentHarness, got %T", newObj)
	}
	return w.validate(h)
}

func (w *AgentHarnessWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentHarnessWebhook) validate(h *agentv1alpha1.AgentHarness) (admission.Warnings, error) {
	if h.Spec.Type != "" {
		if _, err := builder.GetHarnessCommand(h.Spec.Type); err != nil {
			return nil, fmt.Errorf("invalid agent type: %s", h.Spec.Type)
		}
	}
	return nil, nil
}
