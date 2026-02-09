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

// AgentWorkspaceWebhook implements defaulting and validation for AgentWorkspace
type AgentWorkspaceWebhook struct{}

var _ webhook.CustomDefaulter = &AgentWorkspaceWebhook{}
var _ webhook.CustomValidator = &AgentWorkspaceWebhook{}

// SetupWebhookWithManager sets up the webhook with the Manager
func (w *AgentWorkspaceWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&agentv1alpha1.AgentWorkspace{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements webhook.CustomDefaulter
func (w *AgentWorkspaceWebhook) Default(_ context.Context, obj runtime.Object) error {
	ws, ok := obj.(*agentv1alpha1.AgentWorkspace)
	if !ok {
		return fmt.Errorf("expected AgentWorkspace, got %T", obj)
	}

	if ws.Spec.StorageSize == "" {
		ws.Spec.StorageSize = "10Gi"
	}

	for i := range ws.Spec.SharedVolumes {
		if ws.Spec.SharedVolumes[i].StorageSize == "" {
			ws.Spec.SharedVolumes[i].StorageSize = "1Gi"
		}
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator
func (w *AgentWorkspaceWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	ws, ok := obj.(*agentv1alpha1.AgentWorkspace)
	if !ok {
		return nil, fmt.Errorf("expected AgentWorkspace, got %T", obj)
	}
	return w.validate(ws)
}

// ValidateUpdate implements webhook.CustomValidator
func (w *AgentWorkspaceWebhook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	ws, ok := newObj.(*agentv1alpha1.AgentWorkspace)
	if !ok {
		return nil, fmt.Errorf("expected AgentWorkspace, got %T", newObj)
	}
	return w.validate(ws)
}

// ValidateDelete implements webhook.CustomValidator
func (w *AgentWorkspaceWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentWorkspaceWebhook) validate(ws *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
	if ws.Spec.Repository.URL == "" {
		return nil, fmt.Errorf("repository URL is required")
	}

	names := make(map[string]bool)
	for _, vol := range ws.Spec.SharedVolumes {
		if vol.Name == "" {
			return nil, fmt.Errorf("shared volume name is required")
		}
		if vol.MountPath == "" {
			return nil, fmt.Errorf("shared volume mountPath is required for volume %q", vol.Name)
		}
		if names[vol.Name] {
			return nil, fmt.Errorf("duplicate shared volume name: %q", vol.Name)
		}
		names[vol.Name] = true
	}

	return nil, nil
}
