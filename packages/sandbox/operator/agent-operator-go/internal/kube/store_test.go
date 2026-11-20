package kube

import (
	"context"
	"errors"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

var errInjected = errors.New("injected API failure")

func meta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Namespace: namespace, Name: name}
}

func key(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

func newStore(funcs *interceptor.Funcs, objects ...client.Object) (Store, client.Client) {
	builder := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}, &agentv1alpha1.AgentWorkspace{}, &agentv1alpha1.AgentProvider{}).
		WithObjects(objects...)
	if funcs != nil {
		builder = builder.WithInterceptorFuncs(*funcs)
	}
	c := builder.Build()
	return NewStore(c), c
}

func TestStore_ReadsEveryObjectKind(t *testing.T) {
	store, _ := newStore(nil,
		&agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run")},
		&agentv1alpha1.AgentWorkspace{ObjectMeta: meta("ns", "workspace")},
		&agentv1alpha1.AgentProvider{ObjectMeta: meta("ns", "provider")},
		&agentv1alpha1.AgentHarness{ObjectMeta: meta("ns", "harness")},
		&agentv1alpha1.AgentToolchain{ObjectMeta: meta("ns", "toolchain")},
		&corev1.Secret{ObjectMeta: meta("ns", "secret")},
		&batchv1.Job{ObjectMeta: meta("ns", "job")},
		&networkingv1.NetworkPolicy{ObjectMeta: meta("ns", "policy")},
	)
	ctx := context.Background()
	reads := map[string]func() (metav1.Object, error){
		"run":       func() (metav1.Object, error) { return store.GetAgentRun(ctx, key("ns", "run")) },
		"workspace": func() (metav1.Object, error) { return store.GetAgentWorkspace(ctx, key("ns", "workspace")) },
		"provider":  func() (metav1.Object, error) { return store.GetAgentProvider(ctx, key("ns", "provider")) },
		"harness":   func() (metav1.Object, error) { return store.GetAgentHarness(ctx, key("ns", "harness")) },
		"toolchain": func() (metav1.Object, error) { return store.GetAgentToolchain(ctx, key("ns", "toolchain")) },
		"secret":    func() (metav1.Object, error) { return store.GetSecret(ctx, key("ns", "secret")) },
		"job":       func() (metav1.Object, error) { return store.GetJob(ctx, key("ns", "job")) },
		"policy":    func() (metav1.Object, error) { return store.GetNetworkPolicy(ctx, key("ns", "policy")) },
	}
	for name, read := range reads {
		t.Run(name, func(t *testing.T) {
			object, err := read()
			if err != nil {
				t.Fatalf("read error = %v", err)
			}
			if object.GetName() != name {
				t.Errorf("name = %q, want %q", object.GetName(), name)
			}
		})
	}
}

func TestStore_ReportsAnAbsentObjectAsNotFound(t *testing.T) {
	store, _ := newStore(nil)

	_, err := store.GetJob(context.Background(), key("ns", "absent"))

	if !apierrors.IsNotFound(err) {
		t.Errorf("GetJob() error = %v, want NotFound", err)
	}
}

func TestStore_ListsWithinOneNamespace(t *testing.T) {
	store, _ := newStore(nil,
		&agentv1alpha1.AgentProvider{ObjectMeta: meta("ns", "provider")},
		&agentv1alpha1.AgentProvider{ObjectMeta: meta("other", "provider")},
		&agentv1alpha1.AgentPolicy{ObjectMeta: meta("ns", "policy")},
		&agentv1alpha1.AgentPolicy{ObjectMeta: meta("other", "policy")},
	)
	ctx := context.Background()

	providers, err := store.ListAgentProviders(ctx, "ns")
	if err != nil || len(providers) != 1 {
		t.Errorf("ListAgentProviders() = %d, %v, want one", len(providers), err)
	}
	policies, err := store.ListAgentPolicies(ctx, "ns")
	if err != nil || len(policies) != 1 {
		t.Errorf("ListAgentPolicies() = %d, %v, want one", len(policies), err)
	}
}

func TestStore_ListsThePodsOfOneJob(t *testing.T) {
	store, _ := newStore(nil,
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "mine", Labels: map[string]string{"job-name": "job"}}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "ns", Name: "theirs", Labels: map[string]string{"job-name": "other"}}},
	)

	pods, err := store.ListJobPods(context.Background(), "ns", "job")

	if err != nil || len(pods) != 1 || pods[0].Name != "mine" {
		t.Errorf("ListJobPods() = %v, %v, want only the pod of the Job", pods, err)
	}
}

func TestStore_SurfacesListFailures(t *testing.T) {
	store, _ := newStore(&interceptor.Funcs{List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
		return errInjected
	}})
	ctx := context.Background()

	if _, err := store.ListAgentProviders(ctx, "ns"); !errors.Is(err, errInjected) {
		t.Errorf("ListAgentProviders() error = %v", err)
	}
	if _, err := store.ListAgentPolicies(ctx, "ns"); !errors.Is(err, errInjected) {
		t.Errorf("ListAgentPolicies() error = %v", err)
	}
	if _, err := store.ListJobPods(ctx, "ns", "job"); !errors.Is(err, errInjected) {
		t.Errorf("ListJobPods() error = %v", err)
	}
}

func TestStore_CreatesEveryWrittenKind(t *testing.T) {
	store, c := newStore(nil)
	ctx := context.Background()
	claim := &corev1.PersistentVolumeClaim{ObjectMeta: meta("ns", "claim")}
	account := &corev1.ServiceAccount{ObjectMeta: meta("ns", "account")}
	policy := &networkingv1.NetworkPolicy{ObjectMeta: meta("ns", "policy")}
	job := &batchv1.Job{ObjectMeta: meta("ns", "job")}

	for _, err := range []error{
		store.CreatePersistentVolumeClaim(ctx, claim),
		store.CreateServiceAccount(ctx, account),
		store.CreateNetworkPolicy(ctx, policy),
		store.CreateJob(ctx, job),
	} {
		if err != nil {
			t.Fatalf("create error = %v", err)
		}
	}

	for _, object := range []client.Object{claim, account, policy, job} {
		if err := c.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
			t.Errorf("get created %T: %v", object, err)
		}
	}
}

func TestStore_WritesTheStatusOntoTheLatestObject(t *testing.T) {
	stored := &agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run"), Spec: agentv1alpha1.AgentRunSpec{Prompt: "latest"}}
	store, c := newStore(nil, stored,
		&agentv1alpha1.AgentWorkspace{ObjectMeta: meta("ns", "workspace")})
	ctx := context.Background()
	stale := &agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run"), Spec: agentv1alpha1.AgentRunSpec{Prompt: "stale"}}
	stale.Status.Phase = agentv1alpha1.AgentRunPhaseRunning
	workspace := &agentv1alpha1.AgentWorkspace{ObjectMeta: meta("ns", "workspace")}
	workspace.Status.Phase = agentv1alpha1.AgentWorkspacePhaseReady

	if err := store.UpdateAgentRunStatus(ctx, stale); err != nil {
		t.Fatalf("UpdateAgentRunStatus() error = %v", err)
	}
	if err := store.UpdateAgentWorkspaceStatus(ctx, workspace); err != nil {
		t.Fatalf("UpdateAgentWorkspaceStatus() error = %v", err)
	}

	var run agentv1alpha1.AgentRun
	if err := c.Get(ctx, key("ns", "run"), &run); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status.Phase != agentv1alpha1.AgentRunPhaseRunning || run.Spec.Prompt != "latest" {
		t.Errorf("stored run = %q/%q, want the new status on the latest spec", run.Status.Phase, run.Spec.Prompt)
	}
	if stale.Spec.Prompt != "stale" || stale.ResourceVersion != run.ResourceVersion {
		t.Errorf("caller copy = %q at %q, want its own spec at the stored version", stale.Spec.Prompt, stale.ResourceVersion)
	}
	if workspace.Status.Phase != agentv1alpha1.AgentWorkspacePhaseReady || workspace.ResourceVersion == "" {
		t.Errorf("workspace = %+v, want the stored workspace", workspace)
	}
}

func TestStore_IgnoresAStatusWriteForADeletedObject(t *testing.T) {
	store, _ := newStore(nil)
	ctx := context.Background()

	if err := store.UpdateAgentRunStatus(ctx, &agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "absent")}); err != nil {
		t.Errorf("UpdateAgentRunStatus() error = %v", err)
	}
	if err := store.UpdateAgentWorkspaceStatus(ctx, &agentv1alpha1.AgentWorkspace{ObjectMeta: meta("ns", "absent")}); err != nil {
		t.Errorf("UpdateAgentWorkspaceStatus() error = %v", err)
	}
	if err := store.UpdateAgentProviderStatus(ctx, &agentv1alpha1.AgentProvider{ObjectMeta: meta("ns", "absent")}); err != nil {
		t.Errorf("UpdateAgentProviderStatus() error = %v", err)
	}
}

func TestStore_RetriesAStatusWriteOnConflict(t *testing.T) {
	conflicts := 1
	store, c := newStore(&interceptor.Funcs{SubResourceUpdate: func(ctx context.Context, c client.Client, subResource string, object client.Object, opts ...client.SubResourceUpdateOption) error {
		if conflicts > 0 {
			conflicts--
			return apierrors.NewConflict(schema.GroupResource{Resource: "agentruns"}, object.GetName(), errInjected)
		}
		return c.SubResource(subResource).Update(ctx, object, opts...)
	}}, &agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run")})
	run := &agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run")}
	run.Status.Phase = agentv1alpha1.AgentRunPhaseSucceeded

	if err := store.UpdateAgentRunStatus(context.Background(), run); err != nil {
		t.Fatalf("UpdateAgentRunStatus() error = %v", err)
	}

	var stored agentv1alpha1.AgentRun
	if err := c.Get(context.Background(), key("ns", "run"), &stored); err != nil || stored.Status.Phase != agentv1alpha1.AgentRunPhaseSucceeded {
		t.Errorf("stored phase = %q, %v, want Succeeded after the retry", stored.Status.Phase, err)
	}
}

func TestStore_SurfacesAStatusReadFailure(t *testing.T) {
	store, _ := newStore(&interceptor.Funcs{Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
		return errInjected
	}})

	if err := store.UpdateAgentRunStatus(context.Background(), &agentv1alpha1.AgentRun{ObjectMeta: meta("ns", "run")}); !errors.Is(err, errInjected) {
		t.Errorf("UpdateAgentRunStatus() error = %v, want the injected failure", err)
	}
}

func TestStore_SkipsAnEqualProviderStatus(t *testing.T) {
	stored := &agentv1alpha1.AgentProvider{ObjectMeta: meta("ns", "provider")}
	stored.Status.Ready = true
	store, c := newStore(nil, stored)
	ctx := context.Background()
	var before agentv1alpha1.AgentProvider
	if err := c.Get(ctx, key("ns", "provider"), &before); err != nil {
		t.Fatalf("get provider: %v", err)
	}
	provider := before.DeepCopy()

	if err := store.UpdateAgentProviderStatus(ctx, provider); err != nil {
		t.Fatalf("UpdateAgentProviderStatus() error = %v", err)
	}
	if provider.ResourceVersion != before.ResourceVersion {
		t.Errorf("ResourceVersion = %q, want %q for an equal status", provider.ResourceVersion, before.ResourceVersion)
	}

	provider.Status.Ready = false
	if err := store.UpdateAgentProviderStatus(ctx, provider); err != nil {
		t.Fatalf("UpdateAgentProviderStatus() error = %v", err)
	}
	if provider.Status.Ready || provider.ResourceVersion == before.ResourceVersion {
		t.Errorf("provider = ready %v at %q, want a written status", provider.Status.Ready, provider.ResourceVersion)
	}
}
