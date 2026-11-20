package webhook

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// AgentHarnessWebhook implements validation for AgentHarness
type AgentHarnessWebhook struct {
	catalog Catalog
}

// NewAgentHarnessWebhook returns the AgentHarness webhook over the catalog of
// registered agent types.
func NewAgentHarnessWebhook(catalog Catalog) *AgentHarnessWebhook {
	return &AgentHarnessWebhook{catalog: catalog}
}

var _ admission.Validator[*agentv1alpha1.AgentHarness] = &AgentHarnessWebhook{}

func (w *AgentHarnessWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &agentv1alpha1.AgentHarness{}).
		WithValidator(w).
		Complete()
}

func (w *AgentHarnessWebhook) ValidateCreate(_ context.Context, h *agentv1alpha1.AgentHarness) (admission.Warnings, error) {
	return w.validate(h)
}

func (w *AgentHarnessWebhook) ValidateUpdate(_ context.Context, _ *agentv1alpha1.AgentHarness, newObj *agentv1alpha1.AgentHarness) (admission.Warnings, error) {
	return w.validate(newObj)
}

func (w *AgentHarnessWebhook) ValidateDelete(_ context.Context, _ *agentv1alpha1.AgentHarness) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentHarnessWebhook) validate(h *agentv1alpha1.AgentHarness) (admission.Warnings, error) {
	if h.Spec.Type != "" {
		if !w.catalog.HasHarness(h.Spec.Type) {
			return nil, fmt.Errorf("invalid agent type: %s", h.Spec.Type)
		}
	}
	return nil, nil
}
