// Package kube is the controller-runtime adapter of the operator ports. Store
// reads and writes cluster objects through a client.Client, and EventRecorder
// records events through a client-go event recorder. The cores name only their
// own ports; the composition root passes these adapters to them.
package kube

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// Store is the cluster adapter of every reader, writer and status port.
type Store struct {
	client client.Client
}

// NewStore returns a Store over a controller-runtime client.
func NewStore(c client.Client) Store {
	return Store{client: c}
}

type object[T any] interface {
	*T
	client.Object
}

func get[T any, P object[T]](ctx context.Context, c client.Client, key types.NamespacedName) (*T, error) {
	var value T
	if err := c.Get(ctx, key, P(&value)); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetAgentRun reads an AgentRun.
func (s Store) GetAgentRun(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentRun, error) {
	return get[agentv1alpha1.AgentRun](ctx, s.client, key)
}

// GetAgentWorkspace reads an AgentWorkspace.
func (s Store) GetAgentWorkspace(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentWorkspace, error) {
	return get[agentv1alpha1.AgentWorkspace](ctx, s.client, key)
}

// GetAgentProvider reads an AgentProvider.
func (s Store) GetAgentProvider(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentProvider, error) {
	return get[agentv1alpha1.AgentProvider](ctx, s.client, key)
}

// GetAgentHarness reads an AgentHarness.
func (s Store) GetAgentHarness(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentHarness, error) {
	return get[agentv1alpha1.AgentHarness](ctx, s.client, key)
}

// GetAgentToolchain reads an AgentToolchain.
func (s Store) GetAgentToolchain(ctx context.Context, key types.NamespacedName) (*agentv1alpha1.AgentToolchain, error) {
	return get[agentv1alpha1.AgentToolchain](ctx, s.client, key)
}

// GetSecret reads a Secret.
func (s Store) GetSecret(ctx context.Context, key types.NamespacedName) (*corev1.Secret, error) {
	return get[corev1.Secret](ctx, s.client, key)
}

// GetJob reads a Job.
func (s Store) GetJob(ctx context.Context, key types.NamespacedName) (*batchv1.Job, error) {
	return get[batchv1.Job](ctx, s.client, key)
}

// GetNetworkPolicy reads a NetworkPolicy.
func (s Store) GetNetworkPolicy(ctx context.Context, key types.NamespacedName) (*networkingv1.NetworkPolicy, error) {
	return get[networkingv1.NetworkPolicy](ctx, s.client, key)
}

// ListAgentProviders lists the AgentProviders of a namespace.
func (s Store) ListAgentProviders(ctx context.Context, namespace string) ([]agentv1alpha1.AgentProvider, error) {
	var list agentv1alpha1.AgentProviderList
	if err := s.client.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// ListAgentPolicies lists the AgentPolicies of a namespace.
func (s Store) ListAgentPolicies(ctx context.Context, namespace string) ([]agentv1alpha1.AgentPolicy, error) {
	var list agentv1alpha1.AgentPolicyList
	if err := s.client.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// ListJobPods lists the pods of a Job through the job-name label the Job
// controller sets.
func (s Store) ListJobPods(ctx context.Context, namespace, jobName string) ([]corev1.Pod, error) {
	var list corev1.PodList
	if err := s.client.List(ctx, &list, client.InNamespace(namespace), client.MatchingLabels{"job-name": jobName}); err != nil {
		return nil, err
	}
	return list.Items, nil
}

// CreatePersistentVolumeClaim creates a PersistentVolumeClaim.
func (s Store) CreatePersistentVolumeClaim(ctx context.Context, claim *corev1.PersistentVolumeClaim) error {
	return s.client.Create(ctx, claim)
}

// CreateServiceAccount creates a ServiceAccount.
func (s Store) CreateServiceAccount(ctx context.Context, account *corev1.ServiceAccount) error {
	return s.client.Create(ctx, account)
}

// CreateNetworkPolicy creates a NetworkPolicy.
func (s Store) CreateNetworkPolicy(ctx context.Context, policy *networkingv1.NetworkPolicy) error {
	return s.client.Create(ctx, policy)
}

// CreateJob creates a Job.
func (s Store) CreateJob(ctx context.Context, job *batchv1.Job) error {
	return s.client.Create(ctx, job)
}

// UpdateAgentRunStatus writes the status of run onto the latest stored run and
// copies the stored status and resource version back into run. The spec of run
// stays as the caller resolved it, because the controller applies the harness
// defaults to it in memory and builds the Job from it after this write. An
// absent run is not an error.
func (s Store) UpdateAgentRunStatus(ctx context.Context, run *agentv1alpha1.AgentRun) error {
	desired := run.Status.DeepCopy()
	resolved := run.Spec.DeepCopy()
	err := updateStatus(ctx, s.client, client.ObjectKeyFromObject(run), run, func(current *agentv1alpha1.AgentRun) bool {
		current.Status = *desired.DeepCopy()
		return true
	})
	run.Spec = *resolved
	return err
}

// UpdateAgentWorkspaceStatus writes the status of workspace onto the latest
// stored workspace and copies the result back into workspace. An absent
// workspace is not an error.
func (s Store) UpdateAgentWorkspaceStatus(ctx context.Context, workspace *agentv1alpha1.AgentWorkspace) error {
	desired := workspace.Status.DeepCopy()
	return updateStatus(ctx, s.client, client.ObjectKeyFromObject(workspace), workspace, func(current *agentv1alpha1.AgentWorkspace) bool {
		current.Status = *desired.DeepCopy()
		return true
	})
}

// UpdateAgentProviderStatus writes the status of provider onto the latest
// stored provider unless it is already equal, and copies the result back into
// provider. An absent provider is not an error.
func (s Store) UpdateAgentProviderStatus(ctx context.Context, provider *agentv1alpha1.AgentProvider) error {
	desired := provider.Status.DeepCopy()
	return updateStatus(ctx, s.client, client.ObjectKeyFromObject(provider), provider, func(current *agentv1alpha1.AgentProvider) bool {
		if equality.Semantic.DeepEqual(current.Status, *desired) {
			return false
		}
		current.Status = *desired.DeepCopy()
		return true
	})
}

// updateStatus re-reads the object on every conflict retry, so a status write
// never overwrites a newer spec or metadata with the reconciler's stale copy.
func updateStatus[T any, P object[T]](ctx context.Context, c client.Client, key types.NamespacedName, target P, apply func(P) bool) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current T
		if err := c.Get(ctx, key, P(&current)); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if apply(P(&current)) {
			if err := c.Status().Update(ctx, P(&current)); err != nil {
				return err
			}
		}
		*target = current
		return nil
	})
}
