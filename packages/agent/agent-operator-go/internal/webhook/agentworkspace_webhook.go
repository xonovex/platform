package webhook

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
)

// AgentWorkspaceWebhook implements defaulting and validation for AgentWorkspace
type AgentWorkspaceWebhook struct {
	Client client.Client
}

var _ admission.Defaulter[*agentv1alpha1.AgentWorkspace] = &AgentWorkspaceWebhook{}
var _ admission.Validator[*agentv1alpha1.AgentWorkspace] = &AgentWorkspaceWebhook{}

// SetupWebhookWithManager sets up the webhook with the Manager
func (w *AgentWorkspaceWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr, &agentv1alpha1.AgentWorkspace{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements admission.Defaulter
func (w *AgentWorkspaceWebhook) Default(ctx context.Context, ws *agentv1alpha1.AgentWorkspace) error {
	if ws.Spec.StorageSize == "" {
		ws.Spec.StorageSize = "10Gi"
	}

	for i := range ws.Spec.SharedVolumes {
		if ws.Spec.SharedVolumes[i].StorageSize == "" {
			ws.Spec.SharedVolumes[i].StorageSize = "1Gi"
		}
	}

	policy, err := namespacePolicy(ctx, w.Client, ws.Namespace)
	if err != nil {
		return err
	}
	if policy != nil && ws.Spec.RuntimeClassName == nil && policy.Spec.Defaults.RuntimeClassName != nil {
		runtimeClassName := *policy.Spec.Defaults.RuntimeClassName
		ws.Spec.RuntimeClassName = &runtimeClassName
	}

	return nil
}

// ValidateCreate implements admission.Validator
func (w *AgentWorkspaceWebhook) ValidateCreate(ctx context.Context, ws *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
	return w.validate(ctx, ws)
}

// ValidateUpdate implements admission.Validator
func (w *AgentWorkspaceWebhook) ValidateUpdate(ctx context.Context, oldObj *agentv1alpha1.AgentWorkspace, newObj *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
	if !equality.Semantic.DeepEqual(oldObj.Spec, newObj.Spec) {
		return nil, fmt.Errorf("AgentWorkspace spec is immutable after creation; create a new workspace for changed repository or execution inputs")
	}
	return w.validate(ctx, newObj)
}

// ValidateDelete implements admission.Validator
func (w *AgentWorkspaceWebhook) ValidateDelete(_ context.Context, _ *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentWorkspaceWebhook) validate(ctx context.Context, ws *agentv1alpha1.AgentWorkspace) (admission.Warnings, error) {
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
		if _, err := plugins.GetVCSStrategy(ws.Spec.Type); err != nil {
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

	policy, err := namespacePolicy(ctx, w.Client, ws.Namespace)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("namespace requires exactly one AgentPolicy before workspace initialization")
	}
	runtimeClassName := ws.Spec.RuntimeClassName
	if runtimeClassName == nil && policy.Spec.Defaults.RuntimeClassName != nil {
		runtimeClassName = policy.Spec.Defaults.RuntimeClassName
	}
	if runtimeClassName == nil || strings.TrimSpace(*runtimeClassName) == "" {
		return nil, fmt.Errorf("workspace initialization requires an explicit sandboxed runtimeClassName or policy default")
	}
	if !runtimeClassApproved(*runtimeClassName, policy) {
		return nil, fmt.Errorf("runtimeClassName %q is not allowed by the namespace AgentPolicy", *runtimeClassName)
	}
	if ref := ws.Spec.Repository.CredentialsSecretRef; ref != nil {
		if err := requireAllowedSecret(ref.Name, ref.Key, policy); err != nil {
			return nil, fmt.Errorf("repository credentialsSecretRef: %w", err)
		}
	}

	return nil, nil
}
