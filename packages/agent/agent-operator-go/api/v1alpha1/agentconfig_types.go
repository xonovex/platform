package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentConfigSpec defines namespace-level defaults for AgentRuns
type AgentConfigSpec struct {
	// DefaultAgent is the default agent type (default: claude)
	DefaultAgent AgentType `json:"defaultAgent,omitempty"`
	// DefaultProviders maps agent types to default provider names
	DefaultProviders map[AgentType]string `json:"defaultProviders,omitempty"`
	// DefaultImage is the default container image
	DefaultImage string `json:"defaultImage,omitempty"`
	// DefaultResources are the default resource requirements
	DefaultResources corev1.ResourceRequirements `json:"defaultResources,omitempty"`
	// DefaultTimeout is the default timeout for agent runs
	DefaultTimeout *metav1.Duration `json:"defaultTimeout,omitempty"`
	// StorageClass for workspace PVCs
	StorageClass string `json:"storageClass,omitempty"`
	// StorageSize for workspace PVCs (default: 10Gi)
	StorageSize string `json:"storageSize,omitempty"`
	// Env are additional environment variables applied to all runs
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// AgentConfigStatus defines the observed state of AgentConfig
type AgentConfigStatus struct {
	// Conditions of the AgentConfig
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Default Agent",type=string,JSONPath=`.spec.defaultAgent`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentConfig is the Schema for the agentconfigs API
type AgentConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentConfigSpec   `json:"spec,omitempty"`
	Status AgentConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentConfigList contains a list of AgentConfig
type AgentConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentConfig{}, &AgentConfigList{})
}
