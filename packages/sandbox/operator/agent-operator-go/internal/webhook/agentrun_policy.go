package webhook

import (
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentpolicy "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	sharedlaunch "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/launch"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provider"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/validator"
)

func validateTimeout(timeout *metav1.Duration) error {
	if timeout == nil || timeout.Duration <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	return nil
}

func validatePolicyDurations(policy *agentv1alpha1.AgentPolicy) error {
	if timeout := policy.Spec.Defaults.Timeout; timeout != nil && timeout.Duration <= 0 {
		return fmt.Errorf("AgentPolicy default timeout must be greater than zero")
	}
	if timeout := policy.Spec.Enforced.MaxTimeout; timeout != nil && timeout.Duration <= 0 {
		return fmt.Errorf("AgentPolicy maxTimeout must be greater than zero")
	}
	return nil
}

func validateRunSecretAccess(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) error {
	for _, env := range run.Spec.Env {
		if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
			continue
		}
		if err := requireAllowedSecret(env.ValueFrom.SecretKeyRef.Name, env.ValueFrom.SecretKeyRef.Key, policy); err != nil {
			return fmt.Errorf("env %q: %w", env.Name, err)
		}
	}
	if run.Spec.Provider != nil && run.Spec.Provider.AuthSecretRef != nil {
		ref := run.Spec.Provider.AuthSecretRef
		if err := requireAllowedSecret(ref.Name, ref.Key, policy); err != nil {
			return fmt.Errorf("provider authSecretRef: %w", err)
		}
	}
	if run.Spec.Workspace != nil && run.Spec.Workspace.Repository.CredentialsSecretRef != nil {
		ref := run.Spec.Workspace.Repository.CredentialsSecretRef
		if err := requireAllowedSecret(ref.Name, ref.Key, policy); err != nil {
			return fmt.Errorf("repository credentialsSecretRef: %w", err)
		}
	}
	return nil
}

func requireAllowedSecret(name, key string, policy *agentv1alpha1.AgentPolicy) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(key) == "" {
		return fmt.Errorf("secret name and key are required")
	}
	for _, allowed := range policy.Spec.Enforced.AllowedSecretNames {
		if name == allowed {
			return nil
		}
	}
	return fmt.Errorf("secret %q is not allowed by the namespace AgentPolicy", name)
}

// planRun plans the effective run of an admitted AgentRun with the sandbox of
// its namespace, whose AgentPolicy approves the sandboxed runtime classes.
func (w *AgentRunWebhook) planRun(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) (sharedlaunch.Spec, error) {
	route, err := provider.InlineRoute(w.presets, run.Spec.Provider)
	if err != nil {
		return sharedlaunch.Spec{}, fmt.Errorf("agent execution plan: %w", err)
	}
	agentType := agentv1alpha1.AgentTypeClaude
	if run.Spec.Harness != nil && run.Spec.Harness.Type != "" {
		agentType = run.Spec.Harness.Type
	}
	spec, err := launch.Plan(w.planner.Sandbox(approvedRuntimeClasses(policy)), w.planner, run, launch.Execution{
		AgentType: agentType,
		Image:     run.Spec.Image,
		Provider:  route.Provider,
	})
	if err != nil {
		return sharedlaunch.Spec{}, fmt.Errorf("agent execution plan: %w", err)
	}
	return spec, nil
}

// approvedRuntimeClasses are the runtime classes the AgentPolicy approves as
// sandboxed: the enforced one and the allowed ones.
func approvedRuntimeClasses(policy *agentv1alpha1.AgentPolicy) []string {
	enforced := policy.Spec.Enforced
	var approved []string
	if enforced.RuntimeClassName != nil {
		approved = append(approved, *enforced.RuntimeClassName)
	}
	return append(approved, enforced.AllowedRuntimeClassNames...)
}

// sandboxPolicy is what admission demands of every run: a pinned image, host
// tools out of reach and a sandboxed runtime class, and restricted egress when
// the AgentPolicy requires it.
func sandboxPolicy(policy *agentv1alpha1.AgentPolicy) agentpolicy.SandboxPolicy {
	return agentpolicy.SandboxPolicy{
		RequirePinnedProvision:      true,
		RequireHostToolsUnreachable: true,
		RequireKernelIsolation:      true,
		RequireEgressRestricted:     policy.Spec.Enforced.RequireEgressRestricted,
	}
}

// guaranteeMessages state an unmet guarantee in the terms of the AgentRun, so
// the admission error names the field to change. A guarantee with no entry
// reports the error of the shared policy.
var guaranteeMessages = map[agentpolicy.Guarantee]func(run *agentv1alpha1.AgentRun) error{
	agentpolicy.GuaranteePinnedProvision: func(run *agentv1alpha1.AgentRun) error {
		return fmt.Errorf("agent execution image: %w", validator.ValidatePinnedImageReference(run.Spec.Image))
	},
	agentpolicy.GuaranteeKernelIsolation: func(run *agentv1alpha1.AgentRun) error {
		if run.Spec.RuntimeClassName == nil || strings.TrimSpace(*run.Spec.RuntimeClassName) == "" {
			return fmt.Errorf("agent execution requires an explicit sandboxed runtimeClassName or harness/policy default")
		}
		return fmt.Errorf("runtimeClassName %q is not allowed by the namespace AgentPolicy", *run.Spec.RuntimeClassName)
	},
	agentpolicy.GuaranteeEgressRestricted: func(run *agentv1alpha1.AgentRun) error {
		switch {
		case run.Spec.Network == agentv1alpha1.NetworkModeHost:
			return fmt.Errorf("policy requires restricted egress; network=host is unrestricted")
		case run.Spec.NetworkPolicy != nil && run.Spec.NetworkPolicy.Disabled:
			return fmt.Errorf("policy requires restricted egress; NetworkPolicy cannot be disabled")
		default:
			return fmt.Errorf("policy requires restricted egress; custom egress rules cannot be proven restricted")
		}
	},
}

// unmetGuarantee returns the error of the first guarantee of guarantees that
// outcomes report unmet, nil when each is met.
func unmetGuarantee(run *agentv1alpha1.AgentRun, outcomes []agentpolicy.Outcome, guarantees ...agentpolicy.Guarantee) error {
	for _, outcome := range outcomes {
		if outcome.Met() || !slices.Contains(guarantees, outcome.Guarantee) {
			continue
		}
		if message, ok := guaranteeMessages[outcome.Guarantee]; ok {
			return message(run)
		}
		return outcome.Err
	}
	return nil
}

// validateExecutionBoundary checks the boundary every run needs whatever its
// AgentPolicy enforces: an image, the guarantees of the plan, and the security
// context.
func validateExecutionBoundary(run *agentv1alpha1.AgentRun, outcomes []agentpolicy.Outcome) error {
	if err := unmetGuarantee(run, outcomes, agentpolicy.GuaranteePinnedProvision, agentpolicy.GuaranteeHostToolsUnreachable, agentpolicy.GuaranteeKernelIsolation); err != nil {
		return err
	}
	return validateSecurityContextOverrides(run)
}

func enforcePolicy(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy, outcomes []agentpolicy.Outcome) error {
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

	// Reject every override honored by the pod builder that can weaken its
	// hardened container or pod security-context defaults.
	if e.RequireSecurityContext {
		if err := validateSecurityContextOverrides(run); err != nil {
			return err
		}
	}

	// Enforce network policy required
	if e.RequireNetworkPolicy {
		if run.Spec.NetworkPolicy != nil && run.Spec.NetworkPolicy.Disabled {
			return fmt.Errorf("policy requires NetworkPolicy to be enabled")
		}
	}

	if err := unmetGuarantee(run, outcomes, agentpolicy.GuaranteeEgressRestricted); err != nil {
		return err
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
			if imageMatchesAllowedReference(run.Spec.Image, prefix) {
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

func imageMatchesAllowedReference(image, allowed string) bool {
	if allowed == "" {
		return false
	}
	if strings.HasSuffix(allowed, "/") {
		return strings.HasPrefix(image, allowed)
	}
	if strings.Contains(allowed, "@") {
		return image == allowed
	}
	return image == allowed || strings.HasPrefix(image, allowed+"@")
}

func validateSecurityContextOverrides(run *agentv1alpha1.AgentRun) error {
	if sc := run.Spec.SecurityContext; sc != nil {
		if sc.AllowPrivilegeEscalation != nil && *sc.AllowPrivilegeEscalation {
			return fmt.Errorf("policy prohibits AllowPrivilegeEscalation=true")
		}
		if sc.RunAsNonRoot != nil && !*sc.RunAsNonRoot {
			return fmt.Errorf("policy requires RunAsNonRoot=true")
		}
		if sc.ReadOnlyRootFilesystem != nil && !*sc.ReadOnlyRootFilesystem {
			return fmt.Errorf("policy requires ReadOnlyRootFilesystem=true")
		}
		if sc.RunAsUser != nil && *sc.RunAsUser == 0 {
			return fmt.Errorf("policy prohibits RunAsUser=0")
		}
		if sc.RunAsGroup != nil && *sc.RunAsGroup == 0 {
			return fmt.Errorf("policy prohibits RunAsGroup=0")
		}
		if sc.SeccompProfile != nil && sc.SeccompProfile.Type == corev1.SeccompProfileTypeUnconfined {
			return fmt.Errorf("policy prohibits an unconfined container seccomp profile")
		}
		if sc.Capabilities != nil {
			if len(sc.Capabilities.Add) > 0 {
				return fmt.Errorf("policy prohibits adding Linux capabilities")
			}
			if !hasCapability(sc.Capabilities.Drop, "ALL") {
				return fmt.Errorf("policy requires dropping all Linux capabilities")
			}
		}
	}

	if sc := run.Spec.PodSecurityContext; sc != nil {
		if sc.RunAsNonRoot != nil && !*sc.RunAsNonRoot {
			return fmt.Errorf("policy requires pod RunAsNonRoot=true")
		}
		if sc.RunAsUser != nil && *sc.RunAsUser == 0 {
			return fmt.Errorf("policy prohibits pod RunAsUser=0")
		}
		if sc.RunAsGroup != nil && *sc.RunAsGroup == 0 {
			return fmt.Errorf("policy prohibits pod RunAsGroup=0")
		}
		if sc.FSGroup != nil && *sc.FSGroup == 0 {
			return fmt.Errorf("policy prohibits pod FSGroup=0")
		}
		if sc.SeccompProfile != nil && sc.SeccompProfile.Type == corev1.SeccompProfileTypeUnconfined {
			return fmt.Errorf("policy prohibits an unconfined pod seccomp profile")
		}
		if len(sc.Sysctls) > 0 {
			return fmt.Errorf("policy prohibits pod sysctl overrides")
		}
	}

	return nil
}

func hasCapability(capabilities []corev1.Capability, target corev1.Capability) bool {
	for _, capability := range capabilities {
		if strings.EqualFold(string(capability), string(target)) {
			return true
		}
	}
	return false
}
