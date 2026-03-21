package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentToolchainStatus defines the observed state of AgentToolchain
type AgentToolchainStatus struct {
	// Conditions of the AgentToolchain
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentToolchain is the Schema for the agenttoolchains API
type AgentToolchain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ToolchainSpec        `json:"spec,omitempty"`
	Status AgentToolchainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentToolchainList contains a list of AgentToolchain
type AgentToolchainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentToolchain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentToolchain{}, &AgentToolchainList{})
}
