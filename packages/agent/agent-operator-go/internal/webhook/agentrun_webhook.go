package webhook

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AgentRunWebhook implements defaulting and validation for AgentRun
type AgentRunWebhook struct{}

var _ webhook.CustomDefaulter = &AgentRunWebhook{}
var _ webhook.CustomValidator = &AgentRunWebhook{}

// SetupWebhookWithManager sets up the webhook with the Manager
func (w *AgentRunWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&agentv1alpha1.AgentRun{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements webhook.CustomDefaulter
func (w *AgentRunWebhook) Default(_ context.Context, obj runtime.Object) error {
	run, ok := obj.(*agentv1alpha1.AgentRun)
	if !ok {
		return fmt.Errorf("expected AgentRun, got %T", obj)
	}

	if run.Spec.Agent == "" {
		run.Spec.Agent = agentv1alpha1.AgentTypeClaude
	}

	if run.Spec.Timeout == nil {
		defaultTimeout := metav1.Duration{Duration: time.Hour}
		run.Spec.Timeout = &defaultTimeout
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator
func (w *AgentRunWebhook) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	run, ok := obj.(*agentv1alpha1.AgentRun)
	if !ok {
		return nil, fmt.Errorf("expected AgentRun, got %T", obj)
	}

	return w.validate(run)
}

// ValidateUpdate implements webhook.CustomValidator
func (w *AgentRunWebhook) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	run, ok := newObj.(*agentv1alpha1.AgentRun)
	if !ok {
		return nil, fmt.Errorf("expected AgentRun, got %T", newObj)
	}

	return w.validate(run)
}

// ValidateDelete implements webhook.CustomValidator
func (w *AgentRunWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentRunWebhook) validate(run *agentv1alpha1.AgentRun) (admission.Warnings, error) {
	if run.Spec.Agent != agentv1alpha1.AgentTypeClaude && run.Spec.Agent != agentv1alpha1.AgentTypeOpencode {
		return nil, fmt.Errorf("invalid agent type: %s, must be one of: claude, opencode", run.Spec.Agent)
	}

	if run.Spec.WorkspaceRef != "" {
		// Workspace mode: worktree is required, repository must be empty
		if run.Spec.Worktree == nil {
			return nil, fmt.Errorf("worktree is required when workspaceRef is set")
		}
		if run.Spec.Repository.URL != "" {
			return nil, fmt.Errorf("repository must not be set when workspaceRef is set")
		}
	} else {
		// Standalone mode: repository URL is required
		if run.Spec.Repository.URL == "" {
			return nil, fmt.Errorf("repository URL is required")
		}
	}

	if run.Spec.ProviderRef != "" && run.Spec.Provider != nil {
		return nil, fmt.Errorf("cannot specify both providerRef and inline provider")
	}

	return nil, nil
}
