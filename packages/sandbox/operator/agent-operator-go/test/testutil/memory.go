package testutil

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// MemoryStore is an in-memory cluster with the method set of the operator's
// cluster adapter, so a unit test drives a core through its ports without a
// Kubernetes client. Each read returns a deep copy, as an API server does.
type MemoryStore struct {
	runs         map[types.NamespacedName]*agentv1alpha1.AgentRun
	workspaces   map[types.NamespacedName]*agentv1alpha1.AgentWorkspace
	providers    map[types.NamespacedName]*agentv1alpha1.AgentProvider
	harnesses    map[types.NamespacedName]*agentv1alpha1.AgentHarness
	toolchains   map[types.NamespacedName]*agentv1alpha1.AgentToolchain
	policies     map[types.NamespacedName]*agentv1alpha1.AgentPolicy
	secrets      map[types.NamespacedName]*corev1.Secret
	claims       map[types.NamespacedName]*corev1.PersistentVolumeClaim
	accounts     map[types.NamespacedName]*corev1.ServiceAccount
	pods         map[types.NamespacedName]*corev1.Pod
	jobs         map[types.NamespacedName]*batchv1.Job
	netpols      map[types.NamespacedName]*networkingv1.NetworkPolicy
	failures     map[string]error
	statusWrites int
}

// NewMemoryStore returns a MemoryStore that holds a copy of each object. It
// panics on an object type the store does not hold, which is a test bug.
func NewMemoryStore(objects ...any) *MemoryStore {
	store := &MemoryStore{
		runs:       map[types.NamespacedName]*agentv1alpha1.AgentRun{},
		workspaces: map[types.NamespacedName]*agentv1alpha1.AgentWorkspace{},
		providers:  map[types.NamespacedName]*agentv1alpha1.AgentProvider{},
		harnesses:  map[types.NamespacedName]*agentv1alpha1.AgentHarness{},
		toolchains: map[types.NamespacedName]*agentv1alpha1.AgentToolchain{},
		policies:   map[types.NamespacedName]*agentv1alpha1.AgentPolicy{},
		secrets:    map[types.NamespacedName]*corev1.Secret{},
		claims:     map[types.NamespacedName]*corev1.PersistentVolumeClaim{},
		accounts:   map[types.NamespacedName]*corev1.ServiceAccount{},
		pods:       map[types.NamespacedName]*corev1.Pod{},
		jobs:       map[types.NamespacedName]*batchv1.Job{},
		netpols:    map[types.NamespacedName]*networkingv1.NetworkPolicy{},
		failures:   map[string]error{},
	}
	for _, object := range objects {
		var err error
		switch typed := object.(type) {
		case *agentv1alpha1.AgentRun:
			err = insert(store.runs, "agentruns", typed)
		case *agentv1alpha1.AgentWorkspace:
			err = insert(store.workspaces, "agentworkspaces", typed)
		case *agentv1alpha1.AgentProvider:
			err = insert(store.providers, "agentproviders", typed)
		case *agentv1alpha1.AgentHarness:
			err = insert(store.harnesses, "agentharnesses", typed)
		case *agentv1alpha1.AgentToolchain:
			err = insert(store.toolchains, "agenttoolchains", typed)
		case *agentv1alpha1.AgentPolicy:
			err = insert(store.policies, "agentpolicies", typed)
		case *corev1.Secret:
			err = insert(store.secrets, "secrets", typed)
		case *corev1.PersistentVolumeClaim:
			err = insert(store.claims, "persistentvolumeclaims", typed)
		case *corev1.ServiceAccount:
			err = insert(store.accounts, "serviceaccounts", typed)
		case *corev1.Pod:
			err = insert(store.pods, "pods", typed)
		case *batchv1.Job:
			err = insert(store.jobs, "jobs", typed)
		case *networkingv1.NetworkPolicy:
			err = insert(store.netpols, "networkpolicies", typed)
		default:
			panic(fmt.Sprintf("MemoryStore does not hold %T", object))
		}
		if err != nil {
			panic(err)
		}
	}
	return store
}

// Fail makes every later call of the named method return err.
func (s *MemoryStore) Fail(method string, err error) { s.failures[method] = err }

// StatusWrites returns the number of status writes the store applied.
func (s *MemoryStore) StatusWrites() int { return s.statusWrites }

type storedObject[T any] interface {
	*T
	DeepCopy() *T
	GetName() string
	GetNamespace() string
}

func insert[T any, P storedObject[T]](objects map[types.NamespacedName]*T, resource string, object P) error {
	key := types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()}
	if _, exists := objects[key]; exists {
		return apierrors.NewAlreadyExists(schema.GroupResource{Resource: resource}, key.Name)
	}
	objects[key] = object.DeepCopy()
	return nil
}

func load[T any, P storedObject[T]](objects map[types.NamespacedName]*T, resource string, key types.NamespacedName) (*T, error) {
	object, exists := objects[key]
	if !exists {
		return nil, apierrors.NewNotFound(schema.GroupResource{Resource: resource}, key.Name)
	}
	return P(object).DeepCopy(), nil
}

func (s *MemoryStore) failure(method string) error { return s.failures[method] }

// GetAgentRun returns a copy of the stored AgentRun.
func (s *MemoryStore) GetAgentRun(_ context.Context, key types.NamespacedName) (*agentv1alpha1.AgentRun, error) {
	if err := s.failure("GetAgentRun"); err != nil {
		return nil, err
	}
	return load(s.runs, "agentruns", key)
}

// GetAgentWorkspace returns a copy of the stored AgentWorkspace.
func (s *MemoryStore) GetAgentWorkspace(_ context.Context, key types.NamespacedName) (*agentv1alpha1.AgentWorkspace, error) {
	if err := s.failure("GetAgentWorkspace"); err != nil {
		return nil, err
	}
	return load(s.workspaces, "agentworkspaces", key)
}

// GetAgentProvider returns a copy of the stored AgentProvider.
func (s *MemoryStore) GetAgentProvider(_ context.Context, key types.NamespacedName) (*agentv1alpha1.AgentProvider, error) {
	if err := s.failure("GetAgentProvider"); err != nil {
		return nil, err
	}
	return load(s.providers, "agentproviders", key)
}

// GetAgentHarness returns a copy of the stored AgentHarness.
func (s *MemoryStore) GetAgentHarness(_ context.Context, key types.NamespacedName) (*agentv1alpha1.AgentHarness, error) {
	if err := s.failure("GetAgentHarness"); err != nil {
		return nil, err
	}
	return load(s.harnesses, "agentharnesses", key)
}

// GetAgentToolchain returns a copy of the stored AgentToolchain.
func (s *MemoryStore) GetAgentToolchain(_ context.Context, key types.NamespacedName) (*agentv1alpha1.AgentToolchain, error) {
	if err := s.failure("GetAgentToolchain"); err != nil {
		return nil, err
	}
	return load(s.toolchains, "agenttoolchains", key)
}

// GetSecret returns a copy of the stored Secret.
func (s *MemoryStore) GetSecret(_ context.Context, key types.NamespacedName) (*corev1.Secret, error) {
	if err := s.failure("GetSecret"); err != nil {
		return nil, err
	}
	return load(s.secrets, "secrets", key)
}

// GetJob returns a copy of the stored Job.
func (s *MemoryStore) GetJob(_ context.Context, key types.NamespacedName) (*batchv1.Job, error) {
	if err := s.failure("GetJob"); err != nil {
		return nil, err
	}
	return load(s.jobs, "jobs", key)
}

// GetNetworkPolicy returns a copy of the stored NetworkPolicy.
func (s *MemoryStore) GetNetworkPolicy(_ context.Context, key types.NamespacedName) (*networkingv1.NetworkPolicy, error) {
	if err := s.failure("GetNetworkPolicy"); err != nil {
		return nil, err
	}
	return load(s.netpols, "networkpolicies", key)
}

// GetPersistentVolumeClaim returns a copy of the stored PersistentVolumeClaim.
func (s *MemoryStore) GetPersistentVolumeClaim(_ context.Context, key types.NamespacedName) (*corev1.PersistentVolumeClaim, error) {
	return load(s.claims, "persistentvolumeclaims", key)
}

// GetServiceAccount returns a copy of the stored ServiceAccount.
func (s *MemoryStore) GetServiceAccount(_ context.Context, key types.NamespacedName) (*corev1.ServiceAccount, error) {
	return load(s.accounts, "serviceaccounts", key)
}

// ListAgentProviders returns copies of the AgentProviders of a namespace.
func (s *MemoryStore) ListAgentProviders(_ context.Context, namespace string) ([]agentv1alpha1.AgentProvider, error) {
	if err := s.failure("ListAgentProviders"); err != nil {
		return nil, err
	}
	var providers []agentv1alpha1.AgentProvider
	for key, provider := range s.providers {
		if key.Namespace == namespace {
			providers = append(providers, *provider.DeepCopy())
		}
	}
	return providers, nil
}

// ListAgentPolicies returns copies of the AgentPolicies of a namespace.
func (s *MemoryStore) ListAgentPolicies(_ context.Context, namespace string) ([]agentv1alpha1.AgentPolicy, error) {
	if err := s.failure("ListAgentPolicies"); err != nil {
		return nil, err
	}
	var policies []agentv1alpha1.AgentPolicy
	for key, policy := range s.policies {
		if key.Namespace == namespace {
			policies = append(policies, *policy.DeepCopy())
		}
	}
	return policies, nil
}

// ListJobPods returns copies of the pods a Job owns through its job-name label.
func (s *MemoryStore) ListJobPods(_ context.Context, namespace, jobName string) ([]corev1.Pod, error) {
	if err := s.failure("ListJobPods"); err != nil {
		return nil, err
	}
	var pods []corev1.Pod
	for key, pod := range s.pods {
		if key.Namespace == namespace && pod.Labels["job-name"] == jobName {
			pods = append(pods, *pod.DeepCopy())
		}
	}
	return pods, nil
}

// CreatePersistentVolumeClaim stores a copy of the claim.
func (s *MemoryStore) CreatePersistentVolumeClaim(_ context.Context, claim *corev1.PersistentVolumeClaim) error {
	if err := s.failure("CreatePersistentVolumeClaim"); err != nil {
		return err
	}
	return insert(s.claims, "persistentvolumeclaims", claim)
}

// CreateServiceAccount stores a copy of the ServiceAccount.
func (s *MemoryStore) CreateServiceAccount(_ context.Context, account *corev1.ServiceAccount) error {
	if err := s.failure("CreateServiceAccount"); err != nil {
		return err
	}
	return insert(s.accounts, "serviceaccounts", account)
}

// CreateNetworkPolicy stores a copy of the NetworkPolicy.
func (s *MemoryStore) CreateNetworkPolicy(_ context.Context, policy *networkingv1.NetworkPolicy) error {
	if err := s.failure("CreateNetworkPolicy"); err != nil {
		return err
	}
	return insert(s.netpols, "networkpolicies", policy)
}

// CreateJob stores a copy of the Job.
func (s *MemoryStore) CreateJob(_ context.Context, job *batchv1.Job) error {
	if err := s.failure("CreateJob"); err != nil {
		return err
	}
	return insert(s.jobs, "jobs", job)
}

// UpdateAgentRunStatus replaces the stored status and copies the stored run
// back into run. An absent run is not an error.
func (s *MemoryStore) UpdateAgentRunStatus(_ context.Context, run *agentv1alpha1.AgentRun) error {
	if err := s.failure("UpdateAgentRunStatus"); err != nil {
		return err
	}
	stored, exists := s.runs[types.NamespacedName{Namespace: run.Namespace, Name: run.Name}]
	if !exists {
		return nil
	}
	stored.Status = *run.Status.DeepCopy()
	s.statusWrites++
	*run = *stored.DeepCopy()
	return nil
}

// UpdateAgentWorkspaceStatus replaces the stored status and copies the stored
// workspace back into workspace. An absent workspace is not an error.
func (s *MemoryStore) UpdateAgentWorkspaceStatus(_ context.Context, workspace *agentv1alpha1.AgentWorkspace) error {
	if err := s.failure("UpdateAgentWorkspaceStatus"); err != nil {
		return err
	}
	stored, exists := s.workspaces[types.NamespacedName{Namespace: workspace.Namespace, Name: workspace.Name}]
	if !exists {
		return nil
	}
	stored.Status = *workspace.Status.DeepCopy()
	s.statusWrites++
	*workspace = *stored.DeepCopy()
	return nil
}

// UpdateAgentProviderStatus replaces the stored status unless it is equal and
// copies the stored provider back into provider. An absent provider is not an
// error.
func (s *MemoryStore) UpdateAgentProviderStatus(_ context.Context, provider *agentv1alpha1.AgentProvider) error {
	if err := s.failure("UpdateAgentProviderStatus"); err != nil {
		return err
	}
	stored, exists := s.providers[types.NamespacedName{Namespace: provider.Namespace, Name: provider.Name}]
	if !exists {
		return nil
	}
	if !equality.Semantic.DeepEqual(stored.Status, provider.Status) {
		stored.Status = *provider.Status.DeepCopy()
		s.statusWrites++
	}
	*provider = *stored.DeepCopy()
	return nil
}

// SetJobStatus replaces the status of a stored Job, as the Job controller does.
func (s *MemoryStore) SetJobStatus(key types.NamespacedName, status batchv1.JobStatus) error {
	job, exists := s.jobs[key]
	if !exists {
		return apierrors.NewNotFound(schema.GroupResource{Resource: "jobs"}, key.Name)
	}
	job.Status = *status.DeepCopy()
	return nil
}

// JobCount returns the number of stored Jobs.
func (s *MemoryStore) JobCount() int { return len(s.jobs) }

// ClaimCount returns the number of stored PersistentVolumeClaims.
func (s *MemoryStore) ClaimCount() int { return len(s.claims) }
