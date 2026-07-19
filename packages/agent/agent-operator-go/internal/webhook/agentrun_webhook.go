package webhook

import (
	"context"
	"fmt"
	"strings"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
)

// AgentRunWebhook implements defaulting and validation for AgentRun
type AgentRunWebhook struct {
	Client         client.Client
	DecisionClient GovernanceDecisionClient
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
	policy, err := w.namespacePolicy(ctx, run.Namespace)
	if err != nil {
		return err
	}
	if policy != nil {
		applyPolicyDefaults(run, policy)
	}
	if err := w.applyReferencedDefaults(ctx, run); err != nil {
		return err
	}

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

	policy, err := w.namespacePolicy(ctx, run.Namespace)
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
	if err := validateExecutionBoundary(effectiveRun); err != nil {
		return nil, err
	}
	if err := validateAutonomyOversight(effectiveRun, policy); err != nil {
		return nil, err
	}
	if policy != nil {
		if err := enforcePolicy(effectiveRun, policy); err != nil {
			return nil, err
		}
		governanceWarnings, err := w.enforceGovernance(ctx, effectiveRun, policy)
		if err != nil {
			return nil, err
		}
		warnings = append(warnings, governanceWarnings...)
	}

	return warnings, nil
}

// applyReferencedDefaults resolves harness and toolchain references during
// admission so the stored AgentRun contains the exact image and runtime class
// that policy validation approved. Toolchain images take precedence because the
// controller executes that image.
func (w *AgentRunWebhook) applyReferencedDefaults(ctx context.Context, run *agentv1alpha1.AgentRun) error {
	if (run.Spec.HarnessRef != "" || run.Spec.ToolchainRef != "") && w.Client == nil {
		return fmt.Errorf("referenced harness or toolchain requires a Kubernetes client during admission")
	}

	harness, err := resolver.ResolveHarness(ctx, w.Client, run.Namespace, run.Spec.HarnessRef, run.Spec.Harness)
	if err != nil {
		return fmt.Errorf("resolve harness defaults: %w", err)
	}
	if harness != nil {
		if run.Spec.Image == "" {
			run.Spec.Image = harness.Spec.DefaultImage
		}
		if run.Spec.RuntimeClassName == nil && harness.Spec.DefaultRuntimeClassName != nil {
			runtimeClassName := *harness.Spec.DefaultRuntimeClassName
			run.Spec.RuntimeClassName = &runtimeClassName
		}
	}

	toolchain, err := resolver.ResolveToolchain(ctx, w.Client, run.Namespace, run.Spec.ToolchainRef, run.Spec.Toolchain)
	if err != nil {
		return fmt.Errorf("resolve toolchain defaults: %w", err)
	}
	if toolchain != nil && toolchain.Nix != nil {
		if err := validateNixSpec(toolchain.Nix); err != nil {
			return err
		}
		run.Spec.Image = toolchain.Nix.Image
	}
	return nil
}

func validateExecutionBoundary(run *agentv1alpha1.AgentRun) error {
	if run.Spec.Image == "" {
		return fmt.Errorf("agent execution requires an explicit agent-capable image or harness/toolchain image default")
	}
	if err := validator.ValidatePinnedImageReference(run.Spec.Image); err != nil {
		return fmt.Errorf("agent execution image: %w", err)
	}
	if run.Spec.RuntimeClassName == nil || strings.TrimSpace(*run.Spec.RuntimeClassName) == "" {
		return fmt.Errorf("agent execution requires an explicit sandboxed runtimeClassName or harness/policy default")
	}
	return nil
}

func validateAutonomyOversight(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) error {
	triggerKind := run.Annotations[agentv1alpha1.TriggeredByKindAnnotation]
	if triggerKind != "" {
		if triggerKind != "AgentSchedule" && triggerKind != "AgentTrigger" {
			return fmt.Errorf("triggered run has unsupported trigger kind %q", triggerKind)
		}
		if run.Annotations[agentv1alpha1.TriggeredByNameAnnotation] == "" {
			return fmt.Errorf("triggered run requires a trigger name")
		}
		if run.Spec.AccountableOwner == "" {
			return fmt.Errorf("triggered run requires an accountableOwner")
		}
		if err := validateRunProvenance(run.Spec.Provenance); err != nil {
			return fmt.Errorf("triggered run %w", err)
		}
	}

	if run.Spec.Autonomy == nil || run.Spec.Autonomy.Level != agentv1alpha1.AutonomyLevelUnattended {
		return nil
	}
	if run.Spec.AccountableOwner == "" {
		return fmt.Errorf("A3 run requires an accountableOwner")
	}
	if err := validateRunProvenance(run.Spec.Provenance); err != nil {
		return fmt.Errorf("A3 run %w", err)
	}
	if len(run.Spec.Autonomy.ProtectedTargets) == 0 {
		return fmt.Errorf("A3 run requires at least one protected target")
	}
	route := run.Spec.Autonomy.EscalationRoute
	if route == nil || route.Recipient == "" {
		return fmt.Errorf("A3 run requires an accountable escalation recipient")
	}
	if route.Window.Duration <= 0 {
		return fmt.Errorf("A3 escalation window must be positive")
	}
	if route.SafeDefault != agentv1alpha1.EscalationSafeDefaultPause && route.SafeDefault != agentv1alpha1.EscalationSafeDefaultAbandon {
		return fmt.Errorf("A3 escalation safeDefault must be pause or abandon")
	}
	if policy == nil || !policy.Spec.Enforced.RequireGovernanceVerdict {
		return fmt.Errorf("A3 run requires a non-bypassable governance verdict at admission")
	}
	return nil
}

func validateRunProvenance(provenance *agentv1alpha1.AgentRunProvenance) error {
	if provenance == nil || provenance.Model == "" || provenance.Provider == "" || provenance.PromptReference == "" {
		return fmt.Errorf("requires provenance with model, provider, and promptReference")
	}
	if provenance.Tools == nil || provenance.GrantedPermissions == nil {
		return fmt.Errorf("requires provenance with declared tools and grantedPermissions")
	}
	return nil
}

func (w *AgentRunWebhook) namespacePolicy(ctx context.Context, namespace string) (*agentv1alpha1.AgentPolicy, error) {
	if w.Client == nil {
		return nil, nil
	}

	var policyList agentv1alpha1.AgentPolicyList
	if err := w.Client.List(ctx, &policyList, client.InNamespace(namespace)); err != nil {
		return nil, fmt.Errorf("list AgentPolicy in namespace %q: %w", namespace, err)
	}

	switch len(policyList.Items) {
	case 0:
		return nil, nil
	case 1:
		return policyList.Items[0].DeepCopy(), nil
	default:
		return nil, fmt.Errorf("namespace %q has %d AgentPolicies; exactly one is supported", namespace, len(policyList.Items))
	}
}

func applyPolicyDefaults(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) {
	defaults := policy.Spec.Defaults
	if run.Spec.Image == "" && defaults.Image != "" {
		run.Spec.Image = defaults.Image
	}
	if run.Spec.Timeout == nil && defaults.Timeout != nil {
		run.Spec.Timeout = defaults.Timeout.DeepCopy()
	}
	if run.Spec.RuntimeClassName == nil && defaults.RuntimeClassName != nil {
		runtimeClassName := *defaults.RuntimeClassName
		run.Spec.RuntimeClassName = &runtimeClassName
	}
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
	if err := validator.ValidatePinnedImageReference(nix.Image); err != nil {
		return fmt.Errorf("nix toolchain image must use an immutable @sha256: digest")
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
	if e.MaxTimeout != nil {
		if run.Spec.Timeout == nil {
			return fmt.Errorf("policy requires an explicit or default timeout")
		}
		if run.Spec.Timeout.Duration > e.MaxTimeout.Duration {
			return fmt.Errorf("timeout %v exceeds policy maximum %v", run.Spec.Timeout.Duration, e.MaxTimeout.Duration)
		}
	}

	// Enforce resource ceilings. A missing limit is unbounded from the policy's
	// perspective, so every resource named by MaxResources requires a limit.
	if e.MaxResources != nil {
		for name, maximum := range *e.MaxResources {
			limit, found := run.Spec.Resources.Limits[name]
			if !found {
				return fmt.Errorf("policy requires a resource limit for %s", name)
			}
			if limit.Cmp(maximum) > 0 {
				return fmt.Errorf("resource limit %s=%s exceeds policy maximum %s", name, limit.String(), maximum.String())
			}
			if request, found := run.Spec.Resources.Requests[name]; found && request.Cmp(maximum) > 0 {
				return fmt.Errorf("resource request %s=%s exceeds policy maximum %s", name, request.String(), maximum.String())
			}
		}
	}

	// Enforce allowed images
	if len(e.AllowedImages) > 0 {
		if run.Spec.Image == "" {
			return fmt.Errorf("policy requires an explicit or default image")
		}
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
