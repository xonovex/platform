package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentRunPhase represents the phase of an AgentRun
type AgentRunPhase string

const (
	AgentRunPhasePending      AgentRunPhase = "Pending"
	AgentRunPhaseInitializing AgentRunPhase = "Initializing"
	AgentRunPhaseRunning      AgentRunPhase = "Running"
	AgentRunPhaseSucceeded    AgentRunPhase = "Succeeded"
	AgentRunPhaseFailed       AgentRunPhase = "Failed"
	AgentRunPhaseTimedOut     AgentRunPhase = "TimedOut"
)

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentTypeClaude   AgentType = "claude"
	AgentTypeOpencode AgentType = "opencode"
)

// SecretKeyRef references a key in a Kubernetes Secret
type SecretKeyRef struct {
	// Name of the Secret
	Name string `json:"name"`
	// Key within the Secret
	Key string `json:"key"`
}

// RepositorySpec defines the git repository to clone
type RepositorySpec struct {
	// URL is the git repository URL
	URL string `json:"url"`
	// Branch to checkout
	Branch string `json:"branch,omitempty"`
	// Commit to checkout (overrides branch)
	Commit string `json:"commit,omitempty"`
	// CredentialsSecretRef references a Secret containing git credentials
	CredentialsSecretRef *SecretKeyRef `json:"credentialsSecretRef,omitempty"`
}

// WorktreeSpec defines worktree configuration
type WorktreeSpec struct {
	// Branch name for the worktree
	Branch string `json:"branch"`
	// SourceBranch to create the worktree from
	SourceBranch string `json:"sourceBranch,omitempty"`
}

// ProviderSpec defines inline provider configuration
type ProviderSpec struct {
	// Name of the provider (e.g., "gemini", "glm")
	Name string `json:"name"`
	// AuthSecretRef references a Secret containing the auth token
	AuthSecretRef *SecretKeyRef `json:"authSecretRef,omitempty"`
	// Environment variables to set
	Environment map[string]string `json:"environment,omitempty"`
	// CliArgs are additional CLI arguments for the provider
	CliArgs []string `json:"cliArgs,omitempty"`
}

// AgentRunSpec defines the desired state of AgentRun
type AgentRunSpec struct {
	// Agent type to run
	Agent AgentType `json:"agent"`
	// ProviderRef references an AgentProvider in the namespace
	ProviderRef string `json:"providerRef,omitempty"`
	// Provider is an inline provider configuration
	Provider *ProviderSpec `json:"provider,omitempty"`
	// Repository to clone
	Repository RepositorySpec `json:"repository"`
	// Worktree configuration
	Worktree *WorktreeSpec `json:"worktree,omitempty"`
	// Prompt for headless task execution
	Prompt string `json:"prompt,omitempty"`
	// Resources for the agent container
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Timeout for the agent run (default: 1h)
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// Env are additional environment variables
	Env []corev1.EnvVar `json:"env,omitempty"`
	// Image is the container image for the agent
	Image string `json:"image,omitempty"`
	// NodeSelector for pod scheduling
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations for pod scheduling
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// AgentRunStatus defines the observed state of AgentRun
type AgentRunStatus struct {
	// Phase of the AgentRun
	Phase AgentRunPhase `json:"phase,omitempty"`
	// Conditions of the AgentRun
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// JobName is the name of the created Job
	JobName string `json:"jobName,omitempty"`
	// PodName is the name of the running Pod
	PodName string `json:"podName,omitempty"`
	// StartTime is when the run started
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// CompletionTime is when the run completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// ExitCode of the agent process
	ExitCode *int32 `json:"exitCode,omitempty"`
	// WorkspacePVC is the name of the workspace PersistentVolumeClaim
	WorkspacePVC string `json:"workspacePVC,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Agent",type=string,JSONPath=`.spec.agent`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgentRun is the Schema for the agentruns API
type AgentRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentRunSpec   `json:"spec,omitempty"`
	Status AgentRunStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentRunList contains a list of AgentRun
type AgentRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentRun{}, &AgentRunList{})
}
