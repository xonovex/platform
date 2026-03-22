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
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/builder"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
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
	// Mutual exclusivity
	if run.Spec.HarnessRef != "" && run.Spec.Harness != nil {
		return nil, fmt.Errorf("cannot specify both harnessRef and inline harness")
	}
	if run.Spec.ProviderRef != "" && run.Spec.Provider != nil {
		return nil, fmt.Errorf("cannot specify both providerRef and inline provider")
	}
	if run.Spec.WorkspaceRef != "" && run.Spec.Workspace != nil {
		return nil, fmt.Errorf("cannot specify both workspaceRef and inline workspace")
	}
	if run.Spec.ToolchainRef != "" && run.Spec.Toolchain != nil {
		return nil, fmt.Errorf("cannot specify both toolchainRef and inline toolchain")
	}

	// Validate inline types
	if run.Spec.Harness != nil && run.Spec.Harness.Type != "" {
		if _, err := builder.GetHarnessCommand(run.Spec.Harness.Type); err != nil {
			return nil, fmt.Errorf("invalid agent type: %s", run.Spec.Harness.Type)
		}
	}
	if run.Spec.Workspace != nil && run.Spec.Workspace.Type != "" {
		if _, err := builder.GetVCSStrategy(run.Spec.Workspace.Type); err != nil {
			return nil, fmt.Errorf("invalid workspace type: %s", run.Spec.Workspace.Type)
		}
	}
	if run.Spec.Toolchain != nil && run.Spec.Toolchain.Type != "" {
		validTypes := map[agentv1alpha1.ToolchainType]bool{agentv1alpha1.ToolchainTypeNix: true}
		if !validTypes[run.Spec.Toolchain.Type] {
			return nil, fmt.Errorf("invalid toolchain type: %s", run.Spec.Toolchain.Type)
		}
	}

	// Validate inline workspace repository fields
	if run.Spec.Workspace != nil {
		repo := run.Spec.Workspace.Repository
		if err := validator.ValidateRepositoryURL(repo.URL); err != nil {
			return nil, err
		}
		if err := validator.ValidateBranch(repo.Branch); err != nil {
			return nil, err
		}
		if err := validator.ValidateCommit(repo.Commit); err != nil {
			return nil, err
		}
	}

	// Standalone: require workspace with repository URL
	if run.Spec.WorkspaceRef == "" {
		if run.Spec.Workspace == nil || run.Spec.Workspace.Repository.URL == "" {
			return nil, fmt.Errorf("workspace with repository URL is required (or use workspaceRef)")
		}
	}

	return nil, nil
}
