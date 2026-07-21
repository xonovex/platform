package webhook

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/validator"
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

func validateExecutionBoundary(run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) error {
	if policy == nil {
		return fmt.Errorf("namespace requires exactly one AgentPolicy before agent execution")
	}
	if run.Spec.Image == "" {
		return fmt.Errorf("agent execution requires an explicit agent-capable image or harness/toolchain image default")
	}
	if err := validator.ValidatePinnedImageReference(run.Spec.Image); err != nil {
		return fmt.Errorf("agent execution image: %w", err)
	}
	if run.Spec.RuntimeClassName == nil || strings.TrimSpace(*run.Spec.RuntimeClassName) == "" {
		return fmt.Errorf("agent execution requires an explicit sandboxed runtimeClassName or harness/policy default")
	}
	if !runtimeClassApproved(*run.Spec.RuntimeClassName, policy) {
		return fmt.Errorf("runtimeClassName %q is not allowed by the namespace AgentPolicy", *run.Spec.RuntimeClassName)
	}
	if err := validateSecurityContextOverrides(run); err != nil {
		return err
	}
	return nil
}

func runtimeClassApproved(runtimeClassName string, policy *agentv1alpha1.AgentPolicy) bool {
	if policy == nil {
		return false
	}
	enforced := policy.Spec.Enforced
	if enforced.RuntimeClassName != nil && *enforced.RuntimeClassName == runtimeClassName {
		return true
	}
	for _, allowed := range enforced.AllowedRuntimeClassNames {
		if allowed == runtimeClassName {
			return true
		}
	}
	return false
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

	if e.RequireEgressRestricted {
		switch run.Spec.Network {
		case agentv1alpha1.NetworkModeHost:
			return fmt.Errorf("policy requires restricted egress; network=host is unrestricted")
		case agentv1alpha1.NetworkModeProxy:
			return fmt.Errorf("policy requires restricted egress; network=proxy has no enforceable backend")
		}
		if run.Spec.NetworkPolicy != nil {
			if run.Spec.NetworkPolicy.Disabled {
				return fmt.Errorf("policy requires restricted egress; NetworkPolicy cannot be disabled")
			}
			if len(run.Spec.NetworkPolicy.Egress) > 0 {
				return fmt.Errorf("policy requires restricted egress; custom egress rules cannot be proven restricted")
			}
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
