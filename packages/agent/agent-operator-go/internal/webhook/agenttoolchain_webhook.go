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

// AgentToolchainWebhook implements validation for AgentToolchain
type AgentToolchainWebhook struct{}

var _ webhook.CustomValidator = &AgentToolchainWebhook{}

func (w *AgentToolchainWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&agentv1alpha1.AgentToolchain{}).
		WithValidator(w).
		Complete()
}

func (w *AgentToolchainWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	tc, ok := obj.(*agentv1alpha1.AgentToolchain)
	if !ok {
		return nil, fmt.Errorf("expected AgentToolchain, got %T", obj)
	}
	return w.validate(tc)
}

func (w *AgentToolchainWebhook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	tc, ok := newObj.(*agentv1alpha1.AgentToolchain)
	if !ok {
		return nil, fmt.Errorf("expected AgentToolchain, got %T", newObj)
	}
	return w.validate(tc)
}

func (w *AgentToolchainWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentToolchainWebhook) validate(tc *agentv1alpha1.AgentToolchain) (admission.Warnings, error) {
	validTypes := map[agentv1alpha1.ToolchainType]bool{
		agentv1alpha1.ToolchainTypeNix: true,
	}
	if tc.Spec.Type != "" && !validTypes[tc.Spec.Type] {
		return nil, fmt.Errorf("invalid toolchain type: %s", tc.Spec.Type)
	}
	return nil, nil
}
