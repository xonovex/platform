package controller

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provider"
	provshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/resolver"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// Clock reads the current time.
type Clock interface {
	Now() time.Time
}

// EventRecorder records a Kubernetes event about an object.
type EventRecorder interface {
	Record(object runtime.Object, eventType, reason, message string)
}

// ToolchainResolver resolves the provision leaf of a toolchain spec; nil means
// the type is not registered.
type ToolchainResolver interface {
	ResolveToolchain(spec *agentv1alpha1.ToolchainSpec) provshared.Toolchain
}

// VCSResolver resolves the workspace leaf of a workspace type.
type VCSResolver interface {
	ResolveVCSStrategy(wsType agentv1alpha1.WorkspaceType) (wsshared.VCSStrategy, error)
}

// resourceStore creates the objects a reconciler owns and reads them back to
// verify an object that already exists.
type resourceStore interface {
	GetJob(ctx context.Context, key types.NamespacedName) (*batchv1.Job, error)
	GetNetworkPolicy(ctx context.Context, key types.NamespacedName) (*networkingv1.NetworkPolicy, error)
	CreatePersistentVolumeClaim(ctx context.Context, claim *corev1.PersistentVolumeClaim) error
	CreateServiceAccount(ctx context.Context, account *corev1.ServiceAccount) error
	CreateNetworkPolicy(ctx context.Context, policy *networkingv1.NetworkPolicy) error
	CreateJob(ctx context.Context, job *batchv1.Job) error
}

// RunStore is the cluster port of the AgentRun reconciler.
type RunStore interface {
	resourceStore
	resolver.HarnessReader
	resolver.ToolchainReader
	provider.Reader
	GetAgentRun(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentRun, error)
	GetAgentWorkspace(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentWorkspace, error)
	ListJobPods(ctx context.Context, namespace, jobName string) ([]corev1.Pod, error)
	UpdateAgentRunStatus(ctx context.Context, run *agentv1alpha1.AgentRun) error
}

// WorkspaceStore is the cluster port of the AgentWorkspace reconciler.
type WorkspaceStore interface {
	resourceStore
	GetAgentWorkspace(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentWorkspace, error)
	UpdateAgentWorkspaceStatus(ctx context.Context, workspace *agentv1alpha1.AgentWorkspace) error
}

// ProviderStore is the cluster port of the AgentProvider reconciler.
type ProviderStore interface {
	GetAgentProvider(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentProvider, error)
	GetSecret(ctx context.Context, key types.NamespacedName) (*corev1.Secret, error)
	ListAgentProviders(ctx context.Context, namespace string) ([]agentv1alpha1.AgentProvider, error)
	UpdateAgentProviderStatus(ctx context.Context, provider *agentv1alpha1.AgentProvider) error
}
