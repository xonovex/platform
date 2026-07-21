package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	agentvalidation "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/validation"
)

// AgentRunWebhook implements defaulting and validation for AgentRun
type AgentRunWebhook struct {
	Client client.Client
}

var _ admission.Defaulter[*agentv1alpha1.AgentRun] = &AgentRunWebhook{}
var _ admission.Validator[*agentv1alpha1.AgentRun] = &AgentRunWebhook{}

// SetupWebhookWithManager sets up the webhook with the Manager
func (w *AgentRunWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr, &agentv1alpha1.AgentRun{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default implements admission.Defaulter
func (w *AgentRunWebhook) Default(ctx context.Context, run *agentv1alpha1.AgentRun) error {
	policy, err := namespacePolicy(ctx, w.Client, run.Namespace)
	if err != nil {
		return err
	}
	if policy != nil {
		applyPolicyDefaults(run, policy)
	}
	if err := w.applyReferencedDefaults(ctx, run); err != nil {
		return err
	}
	applyBuiltInDefaults(run)
	return nil
}

// ValidateCreate implements admission.Validator
func (w *AgentRunWebhook) ValidateCreate(ctx context.Context, run *agentv1alpha1.AgentRun) (admission.Warnings, error) {
	return w.validate(ctx, run)
}

// ValidateUpdate implements admission.Validator
func (w *AgentRunWebhook) ValidateUpdate(ctx context.Context, oldObj *agentv1alpha1.AgentRun, newObj *agentv1alpha1.AgentRun) (admission.Warnings, error) {
	if !equality.Semantic.DeepEqual(oldObj.Spec, newObj.Spec) {
		return nil, fmt.Errorf("AgentRun spec is immutable after creation; create a new run for changed execution inputs")
	}
	return w.validate(ctx, newObj)
}

// ValidateDelete implements admission.Validator
func (w *AgentRunWebhook) ValidateDelete(_ context.Context, _ *agentv1alpha1.AgentRun) (admission.Warnings, error) {
	return nil, nil
}

func (w *AgentRunWebhook) validate(ctx context.Context, run *agentv1alpha1.AgentRun) (admission.Warnings, error) {
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
		if _, err := plugins.GetHarnessCommand(run.Spec.Harness.Type); err != nil {
			return nil, fmt.Errorf("invalid agent type: %s", run.Spec.Harness.Type)
		}
	}
	if run.Spec.Workspace != nil && run.Spec.Workspace.Type != "" {
		if _, err := plugins.GetVCSStrategy(run.Spec.Workspace.Type); err != nil {
			return nil, fmt.Errorf("invalid workspace type: %s", run.Spec.Workspace.Type)
		}
	}
	if run.Spec.Toolchain != nil && run.Spec.Toolchain.Type != "" {
		validTypes := map[agentv1alpha1.ToolchainType]bool{agentv1alpha1.ToolchainTypeNix: true}
		if !validTypes[run.Spec.Toolchain.Type] {
			return nil, fmt.Errorf("invalid toolchain type: %s", run.Spec.Toolchain.Type)
		}
		if err := validateNixSpec(run.Spec.Toolchain.Nix); err != nil {
			return nil, err
		}
	}
	if run.Spec.Provider != nil {
		if err := validateProviderConfig(
			run.Spec.Provider.PresetRef,
			run.Spec.Provider.AgentType,
			run.Spec.Provider.AuthSecretRef,
			run.Spec.Provider.AuthTokenEnv,
			run.Spec.Provider.Environment,
			run.Spec.Provider.CliArgs,
		); err != nil {
			return nil, fmt.Errorf("invalid inline provider: %w", err)
		}
	}

	// Validate inline workspace repository fields
	if run.Spec.Workspace != nil {
		repo := run.Spec.Workspace.Repository
		if run.Spec.Workspace.StorageSize != "" {
			if _, err := resource.ParseQuantity(run.Spec.Workspace.StorageSize); err != nil {
				return nil, fmt.Errorf("workspace storageSize %q is not a valid resource quantity: %v", run.Spec.Workspace.StorageSize, err)
			}
		}
		if err := agentvalidation.ValidateRepositoryURL(repo.URL); err != nil {
			return nil, err
		}
		if err := agentvalidation.ValidateBranch(repo.Branch); err != nil {
			return nil, err
		}
		if err := agentvalidation.ValidateCommit(repo.Commit); err != nil {
			return nil, err
		}
	}

	// Standalone: require workspace with repository URL
	if run.Spec.WorkspaceRef == "" {
		if run.Spec.Workspace == nil || run.Spec.Workspace.Repository.URL == "" {
			return nil, fmt.Errorf("workspace with repository URL is required (or use workspaceRef)")
		}
	}

	// Validate NetworkPolicy egress rules
	var warnings admission.Warnings
	if run.Spec.Network == agentv1alpha1.NetworkModeProxy {
		return nil, fmt.Errorf("network=proxy is unavailable until an enforceable FQDN-aware backend is configured")
	}
	if run.Spec.NetworkPolicy != nil && !run.Spec.NetworkPolicy.Disabled {
		for _, rule := range run.Spec.NetworkPolicy.Egress {
			if len(rule.To) == 0 {
				warnings = append(warnings, "NetworkPolicy egress rule with empty 'to' allows all destinations")
			}
		}
	}

	policy, err := namespacePolicy(ctx, w.Client, run.Namespace)
	if err != nil {
		return nil, err
	}
	effectiveRun := run.DeepCopy()
	if policy != nil {
		applyPolicyDefaults(effectiveRun, policy)
	}
	if err := w.applyReferencedDefaults(ctx, effectiveRun); err != nil {
		return nil, err
	}
	applyBuiltInDefaults(effectiveRun)
	if err := validateTimeout(effectiveRun.Spec.Timeout); err != nil {
		return nil, err
	}
	if err := validateExecutionBoundary(effectiveRun, policy); err != nil {
		return nil, err
	}
	if policy != nil {
		if err := validatePolicyDurations(policy); err != nil {
			return nil, err
		}
		if err := validateRunSecretAccess(effectiveRun, policy); err != nil {
			return nil, err
		}
		if err := enforcePolicy(effectiveRun, policy); err != nil {
			return nil, err
		}
	}

	return warnings, nil
}
