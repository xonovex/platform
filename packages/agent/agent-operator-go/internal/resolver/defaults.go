package resolver

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const DefaultTimeout = time.Hour

// ResolvedDefaults holds the resolved configuration values
type ResolvedDefaults struct {
	Image         string
	Timeout       time.Duration
	NetworkPolicy *agentv1alpha1.AgentNetworkPolicy
	TTL           *int32
}

// ApplyHarnessDefaults applies every harness default to the run and returns the
// exact values consumed by the Job builder. Applying it more than once is
// idempotent; run-level environment variables override harness defaults by name.
func ApplyHarnessDefaults(run *agentv1alpha1.AgentRun, harness *agentv1alpha1.AgentHarness) ResolvedDefaults {
	image := ""
	timeout := DefaultTimeout

	if harness != nil {
		if harness.Spec.DefaultImage != "" {
			image = harness.Spec.DefaultImage
		}
		if run.Spec.Image == "" && harness.Spec.DefaultImage != "" {
			run.Spec.Image = harness.Spec.DefaultImage
		}
		if harness.Spec.DefaultTimeout != nil {
			timeout = harness.Spec.DefaultTimeout.Duration
		}
		if run.Spec.Timeout == nil && harness.Spec.DefaultTimeout != nil {
			run.Spec.Timeout = harness.Spec.DefaultTimeout.DeepCopy()
		}
		if len(run.Spec.Resources.Requests) == 0 && len(run.Spec.Resources.Limits) == 0 &&
			(len(harness.Spec.DefaultResources.Requests) > 0 || len(harness.Spec.DefaultResources.Limits) > 0) {
			harness.Spec.DefaultResources.DeepCopyInto(&run.Spec.Resources)
		}
		run.Spec.Env = mergeEnvironment(harness.Spec.Env, run.Spec.Env)
		if run.Spec.RuntimeClassName == nil && harness.Spec.DefaultRuntimeClassName != nil {
			runtimeClassName := *harness.Spec.DefaultRuntimeClassName
			run.Spec.RuntimeClassName = &runtimeClassName
		}
		if run.Spec.SecurityContext == nil && harness.Spec.DefaultSecurityContext != nil {
			run.Spec.SecurityContext = harness.Spec.DefaultSecurityContext.DeepCopy()
		}
		if run.Spec.PodSecurityContext == nil && harness.Spec.DefaultPodSecurityContext != nil {
			run.Spec.PodSecurityContext = harness.Spec.DefaultPodSecurityContext.DeepCopy()
		}
		if run.Spec.NetworkPolicy == nil && harness.Spec.DefaultNetworkPolicy != nil {
			run.Spec.NetworkPolicy = harness.Spec.DefaultNetworkPolicy.DeepCopy()
		}
		if run.Spec.TTLSecondsAfterFinished == nil && harness.Spec.DefaultTTLSecondsAfterFinished != nil {
			ttl := *harness.Spec.DefaultTTLSecondsAfterFinished
			run.Spec.TTLSecondsAfterFinished = &ttl
		}
	}

	if run.Spec.Image != "" {
		image = run.Spec.Image
	}
	if run.Spec.Timeout != nil {
		timeout = run.Spec.Timeout.Duration
	}

	return ResolvedDefaults{
		Image:         image,
		Timeout:       timeout,
		NetworkPolicy: run.Spec.NetworkPolicy,
		TTL:           run.Spec.TTLSecondsAfterFinished,
	}
}

func mergeEnvironment(defaults, overrides []corev1.EnvVar) []corev1.EnvVar {
	merged := make([]corev1.EnvVar, 0, len(defaults)+len(overrides))
	indices := make(map[string]int, len(defaults)+len(overrides))
	for _, env := range append(append([]corev1.EnvVar{}, defaults...), overrides...) {
		if index, exists := indices[env.Name]; exists {
			merged[index] = *env.DeepCopy()
			continue
		}
		indices[env.Name] = len(merged)
		merged = append(merged, *env.DeepCopy())
	}
	return merged
}
