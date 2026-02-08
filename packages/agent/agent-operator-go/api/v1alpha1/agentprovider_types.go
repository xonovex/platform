package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentProviderSpec defines the desired state of AgentProvider
type AgentProviderSpec struct {
	// DisplayName is a human-readable name for the provider
	DisplayName string `json:"displayName,omitempty"`
	// AgentTypes this provider supports
	AgentTypes []AgentType `json:"agentTypes"`
	// AuthTokenSecretRef references a Secret containing the auth token
	AuthTokenSecretRef *SecretKeyRef `json:"authTokenSecretRef,omitempty"`
	// Environment variables to set when using this provider
	Environment map[string]string `json:"environment,omitempty"`
	// CliArgs are additional CLI arguments for the provider
	CliArgs []string `json:"cliArgs,omitempty"`
}

// AgentProviderStatus defines the observed state of AgentProvider
type AgentProviderStatus struct {
	// Conditions of the AgentProvider
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// Ready indicates if the provider's secret is accessible
	Ready bool `json:"ready,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Display Name",type=string,JSONPath=`.spec.displayName`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentProvider is the Schema for the agentproviders API
type AgentProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentProviderSpec   `json:"spec,omitempty"`
	Status AgentProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentProviderList contains a list of AgentProvider
type AgentProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentProvider{}, &AgentProviderList{})
}
