package resolver

import (
	"time"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const (
	DefaultImage   = "node:trixie-slim"
	DefaultTimeout = time.Hour
)

// ResolvedDefaults holds the resolved configuration values
type ResolvedDefaults struct {
	Image         string
	Timeout       time.Duration
	NetworkPolicy *agentv1alpha1.AgentNetworkPolicy
	TTL           *int32
}

// ApplyHarnessDefaults resolves image, timeout, runtimeClassName from the harness,
// mutating the run in place for runtimeClassName, and returning image and timeout.
func ApplyHarnessDefaults(run *agentv1alpha1.AgentRun, harness *agentv1alpha1.AgentHarness) ResolvedDefaults {
	image := DefaultImage
	timeout := DefaultTimeout

	if harness != nil {
		if harness.Spec.DefaultImage != "" {
			image = harness.Spec.DefaultImage
		}
		if harness.Spec.DefaultTimeout != nil {
			timeout = harness.Spec.DefaultTimeout.Duration
		}
		if run.Spec.RuntimeClassName == nil && harness.Spec.DefaultRuntimeClassName != nil {
			run.Spec.RuntimeClassName = harness.Spec.DefaultRuntimeClassName
		}
		if run.Spec.SecurityContext == nil && harness.Spec.DefaultSecurityContext != nil {
			run.Spec.SecurityContext = harness.Spec.DefaultSecurityContext
		}
		if run.Spec.PodSecurityContext == nil && harness.Spec.DefaultPodSecurityContext != nil {
			run.Spec.PodSecurityContext = harness.Spec.DefaultPodSecurityContext
		}
	}

	if run.Spec.Image != "" {
		image = run.Spec.Image
	}
	if run.Spec.Timeout != nil {
		timeout = run.Spec.Timeout.Duration
	}

	netpol := run.Spec.NetworkPolicy
	if netpol == nil && harness != nil {
		netpol = harness.Spec.DefaultNetworkPolicy
	}

	ttl := run.Spec.TTLSecondsAfterFinished
	if ttl == nil && harness != nil {
		ttl = harness.Spec.DefaultTTLSecondsAfterFinished
	}

	return ResolvedDefaults{
		Image:         image,
		Timeout:       timeout,
		NetworkPolicy: netpol,
		TTL:           ttl,
	}
}
