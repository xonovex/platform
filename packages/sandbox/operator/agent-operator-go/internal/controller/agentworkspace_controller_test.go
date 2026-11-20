package controller

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

func sandboxedWorkspace(opts ...testutil.AgentWorkspaceOption) *agentv1alpha1.AgentWorkspace {
	workspace := testutil.NewAgentWorkspace("default", "repo", append([]testutil.AgentWorkspaceOption{
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceRuntimeClassName("gvisor"),
	}, opts...)...)
	workspace.UID = "repo-uid"
	return workspace
}

func reconcileWorkspace(t *testing.T, h workspaceHarness) ctrl.Result {
	t.Helper()
	result, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey("repo")})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	return result
}

func reconcileWorkspaceError(h workspaceHarness) error {
	_, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey("repo")})
	return err
}

func getWorkspace(t *testing.T, h workspaceHarness) *agentv1alpha1.AgentWorkspace {
	t.Helper()
	workspace, err := h.store.GetAgentWorkspace(context.Background(), runKey("repo"))
	if err != nil {
		t.Fatalf("get workspace: %v", err)
	}
	return workspace
}

// The first pass provisions the workspace PVC, the ServiceAccount, the init
// NetworkPolicy and the init Job, then parks the workspace in Initializing.
func TestAgentWorkspaceReconcile_ProvisionsTheInitialResources(t *testing.T) {
	h := newWorkspaceHarness(sandboxedWorkspace())

	result := reconcileWorkspace(t, h)

	workspace := getWorkspace(t, h)
	if workspace.Status.WorkspacePVC != "repo-ws" || workspace.Status.InitJobName != "repo-init" {
		t.Errorf("status = %+v, want PVC repo-ws and init Job repo-init", workspace.Status)
	}
	if workspace.Status.Phase != agentv1alpha1.AgentWorkspacePhaseInitializing {
		t.Errorf("phase = %q, want %q", workspace.Status.Phase, agentv1alpha1.AgentWorkspacePhaseInitializing)
	}
	ctx := context.Background()
	if _, err := h.store.GetPersistentVolumeClaim(ctx, runKey("repo-ws")); err != nil {
		t.Errorf("workspace PVC was not created: %v", err)
	}
	if _, err := h.store.GetNetworkPolicy(ctx, runKey("repo-init-netpol")); err != nil {
		t.Errorf("init NetworkPolicy was not created: %v", err)
	}
	if _, err := h.store.GetServiceAccount(ctx, runKey(isoshared.AgentServiceAccountName)); err != nil {
		t.Errorf("agent ServiceAccount was not created: %v", err)
	}
	job := getJob(t, h.store, "repo-init")
	if owner := metav1.GetControllerOf(job); owner == nil || owner.UID != "repo-uid" || owner.Kind != agentWorkspaceKind {
		t.Errorf("init Job controller = %+v, want the workspace", owner)
	}
	if job.Spec.Template.Spec.Containers[0].Image != DefaultWorkspaceInitImage {
		t.Errorf("init image = %q, want the default", job.Spec.Template.Spec.Containers[0].Image)
	}
	if result.RequeueAfter != pendingRequeue {
		t.Errorf("RequeueAfter = %v, want %v", result.RequeueAfter, pendingRequeue)
	}
}

// The init Job deadline comes from the injected policy.
func TestAgentWorkspaceReconcile_TakesTheInitDeadlineFromThePolicy(t *testing.T) {
	store := testutil.NewMemoryStore(sandboxedWorkspace())
	policy := isoshared.DefaultJobPolicy()
	policy.WorkspaceInitTimeout = 90 * time.Second
	reconciler := NewAgentWorkspaceReconciler(AgentWorkspaceDependencies{
		Store: store, VCS: stubRegistry{}, Events: &eventLog{}, Clock: newFakeClock(),
		JobPolicy: policy, WorkspaceInitImage: "registry.example/git@sha256:cccc",
	})

	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey("repo")}); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	job := getJob(t, store, "repo-init")
	if deadline := job.Spec.ActiveDeadlineSeconds; deadline == nil || *deadline != 90 {
		t.Errorf("ActiveDeadlineSeconds = %v, want 90", deadline)
	}
	if image := job.Spec.Template.Spec.Containers[0].Image; image != "registry.example/git@sha256:cccc" {
		t.Errorf("init image = %q, want the configured image", image)
	}
}

func TestAgentWorkspaceReconcile_CreatesSharedVolumePVCs(t *testing.T) {
	h := newWorkspaceHarness(sandboxedWorkspace(testutil.WithSharedVolumes(agentv1alpha1.SharedVolumeSpec{
		Name: "cache", MountPath: "/cache", StorageSize: "1Gi",
	})))

	reconcileWorkspace(t, h)

	if got := getWorkspace(t, h).Status.SharedVolumePVCs["cache"]; got != "repo-cache" {
		t.Errorf("shared volume PVC = %q, want %q", got, "repo-cache")
	}
	if _, err := h.store.GetPersistentVolumeClaim(context.Background(), runKey("repo-cache")); err != nil {
		t.Errorf("shared volume PVC was not created: %v", err)
	}
}

func TestAgentWorkspaceReconcile_EmitsInitStartedNamingTheRepository(t *testing.T) {
	h := newWorkspaceHarness(sandboxedWorkspace())

	reconcileWorkspace(t, h)

	if len(h.events.events) != 1 || h.events.events[0].reason != "WorkspaceInitStarted" ||
		!strings.Contains(h.events.events[0].message, "github.com/org/repo") {
		t.Errorf("events = %+v, want one WorkspaceInitStarted naming the repository", h.events.events)
	}
}

// Reconcile runs repeatedly; a second pass must not duplicate resources or move
// the workspace backwards.
func TestAgentWorkspaceReconcile_IsIdempotentBeforeTheJobFinishes(t *testing.T) {
	h := newWorkspaceHarness(sandboxedWorkspace())

	reconcileWorkspace(t, h)
	result := reconcileWorkspace(t, h)

	workspace := getWorkspace(t, h)
	if workspace.Status.Phase != agentv1alpha1.AgentWorkspacePhaseInitializing {
		t.Errorf("phase = %q, want it to stay Initializing", workspace.Status.Phase)
	}
	if h.store.JobCount() != 1 || h.store.ClaimCount() != 1 {
		t.Errorf("Jobs = %d, PVCs = %d, want one of each", h.store.JobCount(), h.store.ClaimCount())
	}
	if result.RequeueAfter == 0 {
		t.Error("Reconcile() must requeue while the init Job is unfinished")
	}
}

func TestAgentWorkspaceReconcile_MapsTheInitJobStateToThePhase(t *testing.T) {
	tests := []struct {
		name      string
		condition batchv1.JobCondition
		want      agentv1alpha1.AgentWorkspacePhase
		event     string
	}{
		{name: "completed", condition: batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}, want: agentv1alpha1.AgentWorkspacePhaseReady, event: "WorkspaceReady"},
		{name: "failed", condition: batchv1.JobCondition{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "clone failed"}, want: agentv1alpha1.AgentWorkspacePhaseFailed, event: "WorkspaceFailed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newWorkspaceHarness(sandboxedWorkspace())
			reconcileWorkspace(t, h)
			if err := h.store.SetJobStatus(runKey("repo-init"), batchv1.JobStatus{Conditions: []batchv1.JobCondition{test.condition}}); err != nil {
				t.Fatalf("SetJobStatus() error = %v", err)
			}

			result := reconcileWorkspace(t, h)

			if phase := getWorkspace(t, h).Status.Phase; phase != test.want {
				t.Errorf("phase = %q, want %q", phase, test.want)
			}
			if reasons := h.events.reasons(); reasons[len(reasons)-1] != test.event {
				t.Errorf("events = %v, want %s last", reasons, test.event)
			}
			if result.RequeueAfter != 0 {
				t.Errorf("RequeueAfter = %v, want none in a terminal phase", result.RequeueAfter)
			}
		})
	}
}

// Ready and Failed are terminal, so a further reconcile leaves the workspace be.
func TestAgentWorkspaceReconcile_SkipsTerminalPhases(t *testing.T) {
	for _, phase := range []agentv1alpha1.AgentWorkspacePhase{
		agentv1alpha1.AgentWorkspacePhaseReady,
		agentv1alpha1.AgentWorkspacePhaseFailed,
	} {
		t.Run(string(phase), func(t *testing.T) {
			h := newWorkspaceHarness(sandboxedWorkspace(testutil.WithWorkspacePhase(phase)))

			result := reconcileWorkspace(t, h)

			if result.RequeueAfter != 0 || h.store.StatusWrites() != 0 || h.store.ClaimCount() != 0 {
				t.Error("a terminal workspace was reconciled")
			}
		})
	}
}

func TestAgentWorkspaceReconcile_IgnoresAnAbsentWorkspace(t *testing.T) {
	if result := reconcileWorkspace(t, newWorkspaceHarness()); result.RequeueAfter != 0 {
		t.Error("Reconcile() must not requeue for a deleted workspace")
	}
}

func TestAgentWorkspaceReconcile_SkipsAWorkspaceBeingDeleted(t *testing.T) {
	workspace := sandboxedWorkspace()
	deletionTimestamp := metav1.NewTime(testEpoch)
	workspace.DeletionTimestamp = &deletionTimestamp
	h := newWorkspaceHarness(workspace)

	reconcileWorkspace(t, h)

	if h.store.StatusWrites() != 0 || h.store.ClaimCount() != 0 {
		t.Error("a workspace being deleted was reconciled")
	}
}

// The egress policy is derived from the repository host, so an unparseable URL
// stops the reconcile before any pod can reach the network.
func TestAgentWorkspaceReconcile_RefusesAnInvalidRepositoryURL(t *testing.T) {
	h := newWorkspaceHarness(sandboxedWorkspace(testutil.WithWorkspaceRepository("not a url")))

	err := reconcileWorkspaceError(h)

	if err == nil || !strings.Contains(err.Error(), "NetworkPolicy") {
		t.Fatalf("Reconcile() error = %v, want it to name the NetworkPolicy it could not build", err)
	}
	if h.store.JobCount() != 0 {
		t.Error("an init Job was created despite the unusable egress policy")
	}
}

// Workspace init clones untrusted code, so it is refused outright rather than
// falling back to the default runtime when no sandboxed runtimeClassName is set.
func TestAgentWorkspaceReconcile_FailsWithoutASandboxedRuntimeClass(t *testing.T) {
	workspace := sandboxedWorkspace()
	workspace.Spec.RuntimeClassName = nil
	h := newWorkspaceHarness(workspace)

	reconcileWorkspace(t, h)

	updated := getWorkspace(t, h)
	if updated.Status.Phase != agentv1alpha1.AgentWorkspacePhaseFailed {
		t.Errorf("phase = %q, want Failed", updated.Status.Phase)
	}
	if h.store.JobCount() != 0 {
		t.Error("an init Job was created without a sandboxed runtimeClassName")
	}
}

func TestAgentWorkspaceReconcile_FailsOnAnUnknownWorkspaceType(t *testing.T) {
	h := newWorkspaceHarness(sandboxedWorkspace(testutil.WithWorkspaceType("svn")))

	reconcileWorkspace(t, h)

	updated := getWorkspace(t, h)
	if len(updated.Status.Conditions) != 1 || !strings.Contains(updated.Status.Conditions[0].Message, "WorkspaceInitBuildFailed") {
		t.Errorf("conditions = %+v, want WorkspaceInitBuildFailed", updated.Status.Conditions)
	}
}

// An existing policy that the workspace does not own is a name collision the
// controller must refuse to adopt.
func TestAgentWorkspaceReconcile_RejectsAnUnownedNetworkPolicy(t *testing.T) {
	foreign := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "repo-init-netpol", Namespace: "default"}}
	h := newWorkspaceHarness(sandboxedWorkspace(), foreign)

	err := reconcileWorkspaceError(h)

	if err == nil || !strings.Contains(err.Error(), "not controlled by AgentWorkspace") {
		t.Fatalf("Reconcile() error = %v, want an ownership error", err)
	}
}

func TestAgentWorkspaceReconcile_SurfacesStoreFailures(t *testing.T) {
	for _, method := range []string{
		"GetAgentWorkspace", "CreatePersistentVolumeClaim", "UpdateAgentWorkspaceStatus",
		"CreateServiceAccount", "CreateNetworkPolicy", "CreateJob", "GetJob",
	} {
		t.Run(method, func(t *testing.T) {
			h := newWorkspaceHarness(sandboxedWorkspace())
			h.store.Fail(method, errInjected)

			if err := reconcileWorkspaceError(h); !errors.Is(err, errInjected) {
				t.Errorf("Reconcile() error = %v, want the injected failure", err)
			}
		})
	}
}

func TestWorkspaceClaims_LeavesTheStoredStatusUnchanged(t *testing.T) {
	workspace := sandboxedWorkspace(testutil.WithSharedVolumes(agentv1alpha1.SharedVolumeSpec{Name: "cache", MountPath: "/cache"}))
	workspace.Status.SharedVolumePVCs = map[string]string{}

	claims, status := workspaceClaims(workspace)

	if len(workspace.Status.SharedVolumePVCs) != 0 || workspace.Status.WorkspacePVC != "" {
		t.Errorf("stored status = %+v, want it untouched", workspace.Status)
	}
	names := []string{claims[0].Name, claims[1].Name}
	if !slices.Equal(names, []string{"repo-ws", "repo-cache"}) || status.SharedVolumePVCs["cache"] != "repo-cache" {
		t.Errorf("claims = %v, status = %+v", names, status)
	}
}
