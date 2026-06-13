package webhook

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/shell"
)

// AgentToolchainWebhook implements validation for AgentToolchain
type AgentToolchainWebhook struct{}

var _ admission.Validator[*agentv1alpha1.AgentToolchain] = &AgentToolchainWebhook{}

func (w *AgentToolchainWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &agentv1alpha1.AgentToolchain{}).
		WithValidator(w).
		Complete()
}

func (w *AgentToolchainWebhook) ValidateCreate(_ context.Context, tc *agentv1alpha1.AgentToolchain) (admission.Warnings, error) {
	return w.validate(tc)
}

func (w *AgentToolchainWebhook) ValidateUpdate(_ context.Context, _ *agentv1alpha1.AgentToolchain, newObj *agentv1alpha1.AgentToolchain) (admission.Warnings, error) {
	return w.validate(newObj)
}

func (w *AgentToolchainWebhook) ValidateDelete(_ context.Context, _ *agentv1alpha1.AgentToolchain) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentToolchainWebhook) validate(tc *agentv1alpha1.AgentToolchain) (admission.Warnings, error) {
	validTypes := map[agentv1alpha1.ToolchainType]bool{
		agentv1alpha1.ToolchainTypeNix: true,
	}
	if tc.Spec.Type != "" && !validTypes[tc.Spec.Type] {
		return nil, fmt.Errorf("invalid toolchain type: %s", tc.Spec.Type)
	}

	if tc.Spec.Nix != nil {
		for i, pkg := range tc.Spec.Nix.Packages {
			if shell.ContainsMetachars(pkg) {
				return nil, fmt.Errorf("nix.packages[%d] %q contains shell metacharacters", i, pkg)
			}
		}
	}

	return nil, nil
}
