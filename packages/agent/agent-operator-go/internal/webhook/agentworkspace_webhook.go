package webhook

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/builder"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
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
	if err := validator.ValidateRepositoryURL(ws.Spec.Repository.URL); err != nil {
		return nil, err
	}
	if err := validator.ValidateBranch(ws.Spec.Repository.Branch); err != nil {
		return nil, err
	}
	if err := validator.ValidateCommit(ws.Spec.Repository.Commit); err != nil {
		return nil, err
	}
	if ws.Spec.Type != "" {
		if _, err := builder.GetVCSStrategy(ws.Spec.Type); err != nil {
			return nil, fmt.Errorf("invalid workspace type: %s", ws.Spec.Type)
		}
	}

	if ws.Spec.StorageSize != "" {
		if _, err := resource.ParseQuantity(ws.Spec.StorageSize); err != nil {
			return nil, fmt.Errorf("storageSize %q is not a valid resource quantity: %v", ws.Spec.StorageSize, err)
		}
	}

	names := make(map[string]bool)
	mountPaths := make(map[string]bool)
	for _, vol := range ws.Spec.SharedVolumes {
		if vol.Name == "" {
			return nil, fmt.Errorf("shared volume name is required")
		}
		if vol.MountPath == "" {
			return nil, fmt.Errorf("shared volume mountPath is required for volume %q", vol.Name)
		}
		if !strings.HasPrefix(vol.MountPath, "/") {
			return nil, fmt.Errorf("sharedVolumes[%q].mountPath %q must be an absolute path", vol.Name, vol.MountPath)
		}
		if names[vol.Name] {
			return nil, fmt.Errorf("duplicate shared volume name: %q", vol.Name)
		}
		names[vol.Name] = true
		if mountPaths[vol.MountPath] {
			return nil, fmt.Errorf("duplicate mountPath %q in sharedVolumes", vol.MountPath)
		}
		mountPaths[vol.MountPath] = true
		if vol.StorageSize != "" {
			if _, err := resource.ParseQuantity(vol.StorageSize); err != nil {
				return nil, fmt.Errorf("sharedVolumes[%q].storageSize %q is not a valid resource quantity: %v",
					vol.Name, vol.StorageSize, err)
			}
		}
	}

	return nil, nil
}
