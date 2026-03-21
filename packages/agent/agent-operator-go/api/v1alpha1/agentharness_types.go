package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentHarnessStatus defines the observed state of AgentHarness
type AgentHarnessStatus struct {
	// Conditions of the AgentHarness
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentHarness is the Schema for the agentharnesses API
type AgentHarness struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentSpec          `json:"spec,omitempty"`
	Status AgentHarnessStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentHarnessList contains a list of AgentHarness
type AgentHarnessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentHarness `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentHarness{}, &AgentHarnessList{})
}
