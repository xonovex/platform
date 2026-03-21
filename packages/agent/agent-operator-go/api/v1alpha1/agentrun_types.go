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

// WorkspaceType represents the version control system used for workspace management
type WorkspaceType string

const (
	WorkspaceTypeGit     WorkspaceType = "git"
	WorkspaceTypeJujutsu WorkspaceType = "jj"
)

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentTypeClaude   AgentType = "claude"
	AgentTypeOpencode AgentType = "opencode"
)

// ProviderType represents the type of AI provider
type ProviderType string

// ToolchainType represents the type of toolchain
type ToolchainType string

const (
	ToolchainTypeNix ToolchainType = "nix"
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
	// Type of the provider (e.g., "anthropic", "openai")
	Type ProviderType `json:"type,omitempty"`
	// AuthSecretRef references a Secret containing the auth token
	AuthSecretRef *SecretKeyRef `json:"authSecretRef,omitempty"`
	// Environment variables to set
	Environment map[string]string `json:"environment,omitempty"`
	// CliArgs are additional CLI arguments for the provider
	CliArgs []string `json:"cliArgs,omitempty"`
}

// NixSpec configures Nix package provisioning for agent containers.
type NixSpec struct {
	// Packages are nixpkgs attribute names to install (e.g. "nodejs_22", "python3", "ripgrep")
	Packages []string `json:"packages,omitempty"`
	// Image is the Nix container image for the init container (default: "nixos/nix:latest")
	Image string `json:"image,omitempty"`
}

// ToolchainSpec defines toolchain configuration
type ToolchainSpec struct {
	// Type of toolchain
	Type ToolchainType `json:"type"`
	// Nix configures Nix package provisioning
	Nix *NixSpec `json:"nix,omitempty"`
}

// GitWorkspaceConfig holds git-specific workspace configuration
type GitWorkspaceConfig struct {
	// Worktree configuration for git worktrees
	Worktree *WorktreeSpec `json:"worktree,omitempty"`
}

// JujutsuWorkspaceConfig holds jj-specific workspace configuration
type JujutsuWorkspaceConfig struct {
	// Revision for jj workspace
	Revision string `json:"revision,omitempty"`
}

// WorkspaceSpec defines inline workspace configuration
type WorkspaceSpec struct {
	// Type of workspace (git or jj)
	Type WorkspaceType `json:"type,omitempty"`
	// Repository to clone
	Repository RepositorySpec `json:"repository,omitempty"`
	// StorageClass for workspace PVCs
	StorageClass string `json:"storageClass,omitempty"`
	// StorageSize for workspace PVCs (default: 10Gi)
	StorageSize string `json:"storageSize,omitempty"`
	// Git holds git-specific configuration
	Git *GitWorkspaceConfig `json:"git,omitempty"`
	// Jj holds jj-specific configuration
	Jj *JujutsuWorkspaceConfig `json:"jj,omitempty"`
}

// AgentSpec defines agent/harness configuration
type AgentSpec struct {
	// Type of agent
	Type AgentType `json:"type"`
	// DefaultProvider is the default provider name
	DefaultProvider string `json:"defaultProvider,omitempty"`
	// DefaultImage is the default container image
	DefaultImage string `json:"defaultImage,omitempty"`
	// DefaultResources are the default resource requirements
	DefaultResources corev1.ResourceRequirements `json:"defaultResources,omitempty"`
	// DefaultTimeout is the default timeout for agent runs
	DefaultTimeout *metav1.Duration `json:"defaultTimeout,omitempty"`
	// DefaultRuntimeClassName sets the default pod runtimeClassName
	DefaultRuntimeClassName *string `json:"defaultRuntimeClassName,omitempty"`
	// Env are additional environment variables
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// AgentRunSpec defines the desired state of AgentRun
type AgentRunSpec struct {
	// HarnessRef references an AgentHarness in the namespace
	HarnessRef string `json:"harnessRef,omitempty"`
	// Harness is an inline harness configuration
	Harness *AgentSpec `json:"harness,omitempty"`
	// ProviderRef references an AgentProvider in the namespace
	ProviderRef string `json:"providerRef,omitempty"`
	// Provider is an inline provider configuration
	Provider *ProviderSpec `json:"provider,omitempty"`
	// WorkspaceRef references an AgentWorkspace for shared workspace support
	WorkspaceRef string `json:"workspaceRef,omitempty"`
	// Workspace is an inline workspace configuration
	Workspace *WorkspaceSpec `json:"workspace,omitempty"`
	// ToolchainRef references an AgentToolchain in the namespace
	ToolchainRef string `json:"toolchainRef,omitempty"`
	// Toolchain is an inline toolchain configuration
	Toolchain *ToolchainSpec `json:"toolchain,omitempty"`
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
	// RuntimeClassName sets the pod runtimeClassName for VM-based isolation
	RuntimeClassName *string `json:"runtimeClassName,omitempty"`
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
