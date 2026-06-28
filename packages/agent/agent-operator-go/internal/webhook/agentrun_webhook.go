package webhook

import (
	"context"
	"fmt"
	"strings"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
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
func (w *AgentRunWebhook) Default(_ context.Context, run *agentv1alpha1.AgentRun) error {
	if run.Spec.Timeout == nil {
		defaultTimeout := metav1.Duration{Duration: time.Hour}
		run.Spec.Timeout = &defaultTimeout
	}

	return nil
}

// ValidateCreate implements admission.Validator
func (w *AgentRunWebhook) ValidateCreate(ctx context.Context, run *agentv1alpha1.AgentRun) (admission.Warnings, error) {
	return w.validate(ctx, run)
}

// ValidateUpdate implements admission.Validator
func (w *AgentRunWebhook) ValidateUpdate(ctx context.Context, _ *agentv1alpha1.AgentRun, newObj *agentv1alpha1.AgentRun) (admission.Warnings, error) {
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

	// Validate NetworkPolicy egress rules
	var warnings admission.Warnings
	if run.Spec.NetworkPolicy != nil && !run.Spec.NetworkPolicy.Disabled {
		for _, rule := range run.Spec.NetworkPolicy.Egress {
			if len(rule.To) == 0 {
				warnings = append(warnings, "NetworkPolicy egress rule with empty 'to' allows all destinations")
			}
		}
	}

	// Look up AgentPolicy in the namespace
	if w.Client != nil {
		var policyList agentv1alpha1.AgentPolicyList
		if err := w.Client.List(ctx, &policyList, client.InNamespace(run.Namespace)); err != nil {
			warnings = append(warnings, "AgentPolicy lookup failed: "+err.Error())
			return warnings, nil
		}
		if len(policyList.Items) > 0 {
			if err := enforcePolicy(run, &policyList.Items[0]); err != nil {
				return nil, err
			}
		}
	}

	return warnings, nil
}

// validateNixSpec validates the nix toolchain: a pinned rev, exactly one source
// (packages XOR project flake), and a pre-built pinned image. The provisioning is
// build-time, so the image must be supplied — fail closed otherwise.
func validateNixSpec(nix *agentv1alpha1.NixSpec) error {
	if nix == nil {
		return nil
	}
	if nix.NixpkgsRev == "" {
		return fmt.Errorf("nix toolchain requires nixpkgsRev (the reproducibility pin)")
	}
	hasPackages := len(nix.Packages) > 0
	hasFlake := nix.FlakeRef != ""
	if hasPackages && hasFlake {
		return fmt.Errorf("nix toolchain: packages and flakeRef are mutually exclusive")
	}
	if !hasPackages && !hasFlake {
		return fmt.Errorf("nix toolchain requires a source: packages or flakeRef")
	}
	if nix.Image == "" {
		return fmt.Errorf("nix toolchain requires a pre-built pinned image (build-time provisioning)")
	}
	return nil
}

func enforcePolicy(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) error {
	e := policy.Spec.Enforced

	// Enforce runtimeClassName
	if e.RuntimeClassName != nil {
		rc := run.Spec.RuntimeClassName
		if rc == nil || *rc != *e.RuntimeClassName {
			return fmt.Errorf("policy requires runtimeClassName %q", *e.RuntimeClassName)
		}
	}

	// Enforce allowed runtime class names
	if len(e.AllowedRuntimeClassNames) > 0 {
		allowed := false
		for _, name := range e.AllowedRuntimeClassNames {
			if run.Spec.RuntimeClassName != nil && *run.Spec.RuntimeClassName == name {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("runtimeClassName must be one of %v", e.AllowedRuntimeClassNames)
		}
	}

	// Enforce no privilege escalation weakening
	if e.RequireSecurityContext && run.Spec.SecurityContext != nil {
		sc := run.Spec.SecurityContext
		if sc.AllowPrivilegeEscalation != nil && *sc.AllowPrivilegeEscalation {
			return fmt.Errorf("policy prohibits AllowPrivilegeEscalation=true")
		}
		if sc.RunAsNonRoot != nil && !*sc.RunAsNonRoot {
			return fmt.Errorf("policy requires RunAsNonRoot=true")
		}
	}

	// Enforce network policy required
	if e.RequireNetworkPolicy {
		if run.Spec.NetworkPolicy != nil && run.Spec.NetworkPolicy.Disabled {
			return fmt.Errorf("policy requires NetworkPolicy to be enabled")
		}
	}

	// Enforce max timeout
	if e.MaxTimeout != nil && run.Spec.Timeout != nil {
		if run.Spec.Timeout.Duration > e.MaxTimeout.Duration {
			return fmt.Errorf("timeout %v exceeds policy maximum %v", run.Spec.Timeout.Duration, e.MaxTimeout.Duration)
		}
	}

	// Enforce allowed images
	if len(e.AllowedImages) > 0 && run.Spec.Image != "" {
		allowed := false
		for _, prefix := range e.AllowedImages {
			if strings.HasPrefix(run.Spec.Image, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("image %q is not in the allowed images list", run.Spec.Image)
		}
	}

	return nil
}
