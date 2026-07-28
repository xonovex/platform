package controller

import (
	"context"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func newWorkspaceReconciler(
	recorder events.EventRecorder,
	objects ...client.Object,
) (*AgentWorkspaceReconciler, client.Client) {
	scheme := testutil.NewScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentWorkspace{}).
		WithObjects(objects...).
		Build()
	return &AgentWorkspaceReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: recorder,
	}, fakeClient
}

func reconcileWorkspace(
	t *testing.T,
	reconciler *AgentWorkspaceReconciler,
	namespace, name string,
) ctrl.Result {
	t.Helper()
	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: namespace, Name: name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	return result
}

func getWorkspace(t *testing.T, c client.Client, namespace, name string) agentv1alpha1.AgentWorkspace {
	t.Helper()
	var workspace agentv1alpha1.AgentWorkspace
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: namespace, Name: name,
	}, &workspace); err != nil {
		t.Fatalf("Get workspace: %v", err)
	}
	return workspace
}

// The first pass provisions the workspace PVC, the ServiceAccount, the init
// NetworkPolicy and the init Job, then parks the workspace in Initializing.
func TestAgentWorkspaceReconcile_ProvisionsTheInitialResources(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

	reconcileWorkspace(t, reconciler, "default", "repo")

	resolved := getWorkspace(t, c, "default", "repo")
	if resolved.Status.WorkspacePVC != "repo-ws" {
		t.Errorf("workspace PVC = %q, want %q", resolved.Status.WorkspacePVC, "repo-ws")
	}
	if resolved.Status.InitJobName != "repo-init" {
		t.Errorf("init job = %q, want %q", resolved.Status.InitJobName, "repo-init")
	}
	if resolved.Status.Phase != agentv1alpha1.AgentWorkspacePhaseInitializing {
		t.Errorf("phase = %q, want %q",
			resolved.Status.Phase, agentv1alpha1.AgentWorkspacePhaseInitializing)
	}

	ctx := context.Background()
	var pvc corev1.PersistentVolumeClaim
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: "repo-ws",
	}, &pvc); err != nil {
		t.Errorf("workspace PVC was not created: %v", err)
	}
	var job batchv1.Job
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: "repo-init",
	}, &job); err != nil {
		t.Errorf("init Job was not created: %v", err)
	}
	var policy networkingv1.NetworkPolicy
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: "repo-init-netpol",
	}, &policy); err != nil {
		t.Errorf("init NetworkPolicy was not created: %v", err)
	}
	var serviceAccount corev1.ServiceAccount
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: isoshared.AgentServiceAccountName,
	}, &serviceAccount); err != nil {
		t.Errorf("agent ServiceAccount was not created: %v", err)
	}
}

func TestAgentWorkspaceReconcile_CreatesSharedVolumePVCs(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"),
		testutil.WithSharedVolumes(agentv1alpha1.SharedVolumeSpec{
			Name: "cache", MountPath: "/cache", StorageSize: "1Gi",
		}))
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

	reconcileWorkspace(t, reconciler, "default", "repo")

	resolved := getWorkspace(t, c, "default", "repo")
	if got := resolved.Status.SharedVolumePVCs["cache"]; got != "repo-cache" {
		t.Errorf("shared volume PVC = %q, want %q", got, "repo-cache")
	}
	var pvc corev1.PersistentVolumeClaim
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "repo-cache",
	}, &pvc); err != nil {
		t.Errorf("shared volume PVC was not created: %v", err)
	}
}

func TestAgentWorkspaceReconcile_EmitsInitStartedNamingTheRepository(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	recorder := events.NewFakeRecorder(10)
	reconciler, _ := newWorkspaceReconciler(recorder, workspace)

	reconcileWorkspace(t, reconciler, "default", "repo")

	select {
	case event := <-recorder.Events:
		if !strings.Contains(event, "WorkspaceInitStarted") {
			t.Errorf("event = %q, want a WorkspaceInitStarted event", event)
		}
	default:
		t.Fatal("expected an event when the init Job is created")
	}
}

// Reconcile runs repeatedly; a second pass must not duplicate resources or move
// the workspace backwards.
func TestAgentWorkspaceReconcile_IsIdempotentBeforeTheJobFinishes(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

	reconcileWorkspace(t, reconciler, "default", "repo")
	first := getWorkspace(t, c, "default", "repo")
	result := reconcileWorkspace(t, reconciler, "default", "repo")
	second := getWorkspace(t, c, "default", "repo")

	if first.Status.InitJobName != second.Status.InitJobName {
		t.Errorf("init job changed across reconciles: %q then %q",
			first.Status.InitJobName, second.Status.InitJobName)
	}
	if second.Status.Phase != agentv1alpha1.AgentWorkspacePhaseInitializing {
		t.Errorf("phase = %q, want it to stay %q",
			second.Status.Phase, agentv1alpha1.AgentWorkspacePhaseInitializing)
	}
	if result.RequeueAfter == 0 {
		t.Error("Reconcile() must requeue while the init Job is unfinished")
	}
}

func TestAgentWorkspaceReconcile_ReportsReadyWhenTheInitJobCompletes(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)
	reconcileWorkspace(t, reconciler, "default", "repo")

	ctx := context.Background()
	var job batchv1.Job
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: "repo-init",
	}, &job); err != nil {
		t.Fatalf("get init Job: %v", err)
	}
	job.Status.Conditions = []batchv1.JobCondition{
		{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
	}
	if err := c.Status().Update(ctx, &job); err != nil {
		t.Fatalf("update init Job status: %v", err)
	}

	reconcileWorkspace(t, reconciler, "default", "repo")

	resolved := getWorkspace(t, c, "default", "repo")
	if resolved.Status.Phase != agentv1alpha1.AgentWorkspacePhaseReady {
		t.Errorf("phase = %q, want %q",
			resolved.Status.Phase, agentv1alpha1.AgentWorkspacePhaseReady)
	}
}

// Ready and Failed are terminal, so a further reconcile leaves the workspace be.
func TestAgentWorkspaceReconcile_SkipsTerminalPhases(t *testing.T) {
	for _, phase := range []agentv1alpha1.AgentWorkspacePhase{
		agentv1alpha1.AgentWorkspacePhaseReady,
		agentv1alpha1.AgentWorkspacePhaseFailed,
	} {
		t.Run(string(phase), func(t *testing.T) {
			workspace := testutil.NewAgentWorkspace("default", "repo",
				testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
				testutil.WithWorkspaceRuntimeClassName("gvisor"),
				testutil.WithWorkspacePhase(phase))
			reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

			result := reconcileWorkspace(t, reconciler, "default", "repo")

			if result.RequeueAfter != 0 {
				t.Errorf("RequeueAfter = %v, want no requeue in a terminal phase",
					result.RequeueAfter)
			}
			resolved := getWorkspace(t, c, "default", "repo")
			if resolved.Status.InitJobName != "" {
				t.Errorf("terminal workspace gained init job %q",
					resolved.Status.InitJobName)
			}
		})
	}
}

func TestAgentWorkspaceReconcile_IgnoresAnAbsentWorkspace(t *testing.T) {
	reconciler, _ := newWorkspaceReconciler(events.NewFakeRecorder(10))

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "absent"},
	})

	if err != nil {
		t.Fatalf("Reconcile() error = %v, want nil for a deleted workspace", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("Reconcile() must not requeue for a deleted workspace")
	}
}

func TestAgentWorkspaceReconcile_SkipsAWorkspaceBeingDeleted(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	deletionTimestamp := metav1.Now()
	workspace.DeletionTimestamp = &deletionTimestamp
	workspace.Finalizers = []string{"xonovex.com/test"}
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

	reconcileWorkspace(t, reconciler, "default", "repo")

	resolved := getWorkspace(t, c, "default", "repo")
	if resolved.Status.WorkspacePVC != "" {
		t.Errorf("a workspace being deleted provisioned PVC %q",
			resolved.Status.WorkspacePVC)
	}
}

// The egress policy is derived from the repository host, so an unparseable URL
// stops the reconcile before any pod can reach the network.
func TestAgentWorkspaceReconcile_RefusesAnInvalidRepositoryURL(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("not a url"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "repo"},
	})

	if err == nil {
		t.Fatal("Reconcile() must fail when the repository URL cannot be parsed")
	}
	if !strings.Contains(err.Error(), "NetworkPolicy") {
		t.Errorf("error = %v, want it to name the NetworkPolicy it could not build", err)
	}
	resolved := getWorkspace(t, c, "default", "repo")
	if resolved.Status.InitJobName != "" {
		t.Errorf("init Job %q was created despite the unusable egress policy",
			resolved.Status.InitJobName)
	}
}

// Workspace init clones untrusted code, so it is refused outright rather than
// falling back to the default runtime when no sandboxed runtimeClassName is set.
func TestAgentWorkspaceReconcile_FailsWithoutASandboxedRuntimeClass(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"))
	reconciler, c := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)

	reconcileWorkspace(t, reconciler, "default", "repo")

	resolved := getWorkspace(t, c, "default", "repo")
	if resolved.Status.Phase != agentv1alpha1.AgentWorkspacePhaseFailed {
		t.Errorf("phase = %q, want %q",
			resolved.Status.Phase, agentv1alpha1.AgentWorkspacePhaseFailed)
	}
	if resolved.Status.InitJobName != "" {
		t.Errorf("init Job %q was created without a sandboxed runtimeClassName",
			resolved.Status.InitJobName)
	}
}

func TestEnsureWorkspaceInitNetworkPolicy_IsIdempotent(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	reconciler, _ := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace)
	ctx := context.Background()

	if err := reconciler.ensureWorkspaceInitNetworkPolicy(ctx, workspace); err != nil {
		t.Fatalf("first ensureWorkspaceInitNetworkPolicy() error = %v", err)
	}
	if err := reconciler.ensureWorkspaceInitNetworkPolicy(ctx, workspace); err != nil {
		t.Fatalf("second ensureWorkspaceInitNetworkPolicy() error = %v", err)
	}
}

// An existing policy that the workspace does not own is a name collision the
// controller must refuse to adopt.
func TestEnsureWorkspaceInitNetworkPolicy_RejectsAnUnownedPolicy(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "repo",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"))
	foreign := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{
		Name: "repo-init-netpol", Namespace: "default",
	}}
	reconciler, _ := newWorkspaceReconciler(events.NewFakeRecorder(10), workspace, foreign)

	err := reconciler.ensureWorkspaceInitNetworkPolicy(context.Background(), workspace)

	if err == nil {
		t.Fatal("ensureWorkspaceInitNetworkPolicy() must reject an unowned policy")
	}
	if !strings.Contains(err.Error(), "not controlled by") {
		t.Errorf("error = %v, want an ownership error", err)
	}
}

func TestUpdateWorkspaceStatus_IgnoresADeletedWorkspace(t *testing.T) {
	reconciler, _ := newWorkspaceReconciler(events.NewFakeRecorder(10))
	absent := testutil.NewAgentWorkspace("default", "absent")
	absent.Status.Phase = agentv1alpha1.AgentWorkspacePhaseReady

	if err := reconciler.updateWorkspaceStatus(context.Background(), absent); err != nil {
		t.Errorf("updateWorkspaceStatus() error = %v, want nil for a deleted workspace", err)
	}
}
