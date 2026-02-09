package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentWorkspacePhase represents the phase of an AgentWorkspace
type AgentWorkspacePhase string

const (
	AgentWorkspacePhasePending      AgentWorkspacePhase = "Pending"
	AgentWorkspacePhaseInitializing AgentWorkspacePhase = "Initializing"
	AgentWorkspacePhaseReady        AgentWorkspacePhase = "Ready"
	AgentWorkspacePhaseFailed       AgentWorkspacePhase = "Failed"
)

// SharedVolumeSpec defines a shared volume for agent config/state directories
type SharedVolumeSpec struct {
	// Name is the PVC name suffix and volume name
	Name string `json:"name"`
	// MountPath is where to mount in agent containers
	MountPath string `json:"mountPath"`
	// StorageSize for the shared volume PVC (default: 1Gi)
	StorageSize string `json:"storageSize,omitempty"`
}

// AgentWorkspaceSpec defines the desired state of AgentWorkspace
type AgentWorkspaceSpec struct {
	// Repository to clone into the shared workspace
	Repository RepositorySpec `json:"repository"`
	// StorageClass for the workspace PVC (must support ReadWriteMany)
	StorageClass string `json:"storageClass,omitempty"`
	// StorageSize for the workspace PVC (default: 10Gi)
	StorageSize string `json:"storageSize,omitempty"`
	// SharedVolumes are optional shared config/state directories for agents
	SharedVolumes []SharedVolumeSpec `json:"sharedVolumes,omitempty"`
}

// AgentWorkspaceStatus defines the observed state of AgentWorkspace
type AgentWorkspaceStatus struct {
	// Phase of the AgentWorkspace
	Phase AgentWorkspacePhase `json:"phase,omitempty"`
	// Conditions of the AgentWorkspace
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// WorkspacePVC is the name of the workspace PersistentVolumeClaim
	WorkspacePVC string `json:"workspacePVC,omitempty"`
	// SharedVolumePVCs maps volume name to PVC name
	SharedVolumePVCs map[string]string `json:"sharedVolumePVCs,omitempty"`
	// InitJobName is the name of the init Job that cloned the repository
	InitJobName string `json:"initJobName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentWorkspace is the Schema for the agentworkspaces API
type AgentWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentWorkspaceSpec   `json:"spec,omitempty"`
	Status AgentWorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentWorkspaceList contains a list of AgentWorkspace
type AgentWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentWorkspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentWorkspace{}, &AgentWorkspaceList{})
}
