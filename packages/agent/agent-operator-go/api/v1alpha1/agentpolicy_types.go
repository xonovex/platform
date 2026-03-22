package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ensure unused imports don't cause errors
var _ = resource.Quantity{}

// AgentPolicyEnforced defines constraints that AgentRuns in this namespace cannot override.
type AgentPolicyEnforced struct {
	// RuntimeClassName, if set, requires all AgentRuns to use this runtimeClassName.
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`

	// RequireSecurityContext, if true, prevents AgentRuns from setting
	// SecurityContext fields that weaken the hardened defaults
	// (e.g. RunAsNonRoot=false, AllowPrivilegeEscalation=true).
	RequireSecurityContext bool `json:"requireSecurityContext,omitempty"`

	// RequireNetworkPolicy, if true, requires AgentRuns to have a NetworkPolicy
	// (i.e. spec.networkPolicy must not be Disabled).
	RequireNetworkPolicy bool `json:"requireNetworkPolicy,omitempty"`

	// MaxTimeout is the maximum allowed timeout for AgentRuns.
	MaxTimeout *metav1.Duration `json:"maxTimeout,omitempty"`

	// MaxResources defines the upper bound for any single container's resource limits.
	MaxResources *corev1.ResourceList `json:"maxResources,omitempty"`

	// AllowedImages is a list of allowed container image prefixes.
	// If set, AgentRun.Spec.Image must match one of these prefixes.
	AllowedImages []string `json:"allowedImages,omitempty"`

	// AllowedRuntimeClassNames lists permitted runtimeClassNames.
	// If non-empty, AgentRun.Spec.RuntimeClassName must be in this list.
	AllowedRuntimeClassNames []string `json:"allowedRuntimeClassNames,omitempty"`
}

// AgentPolicyDefaults defines overridable defaults applied when AgentRun fields are absent.
type AgentPolicyDefaults struct {
	// Image is the default container image when AgentRun.Spec.Image is not set.
	Image string `json:"image,omitempty"`

	// Timeout is the default timeout when AgentRun.Spec.Timeout is not set.
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// RuntimeClassName is the default runtimeClassName when not set on AgentRun or AgentHarness.
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`
}

// AgentPolicySpec defines the desired state of AgentPolicy.
type AgentPolicySpec struct {
	// Enforced constraints — AgentRuns that violate these are rejected by the webhook.
	Enforced AgentPolicyEnforced `json:"enforced,omitempty"`

	// Defaults — applied when AgentRun fields are not set.
	Defaults AgentPolicyDefaults `json:"defaults,omitempty"`
}

// AgentPolicyStatus defines the observed state of AgentPolicy.
type AgentPolicyStatus struct {
	// Conditions of the AgentPolicy.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentPolicy defines enforced security constraints and defaults for AgentRuns in a namespace.
type AgentPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentPolicySpec   `json:"spec,omitempty"`
	Status AgentPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentPolicyList contains a list of AgentPolicy.
type AgentPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentPolicy{}, &AgentPolicyList{})
}
