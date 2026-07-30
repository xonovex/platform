package controller

import (
	"context"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

// runnableAgentRun is an AgentRun as admission leaves it: image and sandboxed
// runtimeClassName already resolved, so the reconcile reaches the Job.
func runnableAgentRun(name string, opts ...testutil.AgentRunOption) *agentv1alpha1.AgentRun {
	return testutil.NewAgentRun("default", name,
		append([]testutil.AgentRunOption{testutil.WithPrompt("do the thing")}, opts...)...)
}

func newRunReconciler(objects ...client.Object) (*AgentRunReconciler, client.Client) {
	scheme := testutil.NewScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(objects...).
		Build()
	return &AgentRunReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: events.NewFakeRecorder(20),
	}, fakeClient
}

func reconcileRun(t *testing.T, reconciler *AgentRunReconciler, name string) ctrl.Result {
	t.Helper()
	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	return result
}

func getRun(t *testing.T, c client.Client, name string) agentv1alpha1.AgentRun {
	t.Helper()
	var run agentv1alpha1.AgentRun
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: name,
	}, &run); err != nil {
		t.Fatalf("Get run: %v", err)
	}
	return run
}

// A phase change carries its explanation as a condition of the same name rather
// than a status field.
func phaseMessage(run agentv1alpha1.AgentRun, phase agentv1alpha1.AgentRunPhase) string {
	for _, condition := range run.Status.Conditions {
		if condition.Type == string(phase) {
			return condition.Message
		}
	}
	return ""
}

// A run with no workspaceRef provisions its own PVC and Job.
func TestAgentRunReconcile_StandaloneProvisionsPVCAndJob(t *testing.T) {
	reconciler, c := newRunReconciler(runnableAgentRun("solo"))

	reconcileRun(t, reconciler, "solo")

	run := getRun(t, c, "solo")
	if run.Status.WorkspacePVC != "solo-workspace" {
		t.Errorf("workspace PVC = %q, want %q", run.Status.WorkspacePVC, "solo-workspace")
	}

	ctx := context.Background()
	var pvc corev1.PersistentVolumeClaim
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: "solo-workspace",
	}, &pvc); err != nil {
		t.Errorf("workspace PVC was not created: %v", err)
	}
	var job batchv1.Job
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: "default", Name: "solo",
	}, &job); err != nil {
		t.Errorf("agent Job was not created: %v", err)
	}
}

func TestAgentRunReconcile_StandaloneHonoursWorkspaceStorageSettings(t *testing.T) {
	run := runnableAgentRun("solo", testutil.WithWorkspace(&agentv1alpha1.WorkspaceSpec{
		Type:         agentv1alpha1.WorkspaceTypeJujutsu,
		StorageClass: "fast",
		StorageSize:  "42Gi",
	}))
	reconciler, c := newRunReconciler(run)

	reconcileRun(t, reconciler, "solo")

	var pvc corev1.PersistentVolumeClaim
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "solo-workspace",
	}, &pvc); err != nil {
		t.Fatalf("workspace PVC was not created: %v", err)
	}
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "fast" {
		t.Errorf("storage class = %v, want %q", pvc.Spec.StorageClassName, "fast")
	}
	if got := pvc.Spec.Resources.Requests.Storage().String(); got != "42Gi" {
		t.Errorf("storage size = %q, want %q", got, "42Gi")
	}
}

// Admission is responsible for resolving the image; the controller refuses to
// guess one.
func TestAgentRunReconcile_FailsWithoutAResolvedImage(t *testing.T) {
	run := runnableAgentRun("solo", testutil.WithImage(""))
	reconciler, c := newRunReconciler(run)

	reconcileRun(t, reconciler, "solo")

	resolved := getRun(t, c, "solo")
	if resolved.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
		t.Errorf("phase = %q, want %q", resolved.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
	}
	message := phaseMessage(resolved, agentv1alpha1.AgentRunPhaseFailed)
	if !strings.Contains(message, "ImageResolutionFailed") {
		t.Errorf("message = %q, want it to name the unresolved image", message)
	}
}

// Agent pods run untrusted model output, so a run without a sandboxed runtime
// class fails rather than falling back to the default runtime.
func TestAgentRunReconcile_FailsWithoutASandboxedRuntimeClass(t *testing.T) {
	run := runnableAgentRun("solo", testutil.WithRuntimeClassName(""))
	reconciler, c := newRunReconciler(run)

	reconcileRun(t, reconciler, "solo")

	resolved := getRun(t, c, "solo")
	if resolved.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
		t.Errorf("phase = %q, want %q", resolved.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
	}
	message := phaseMessage(resolved, agentv1alpha1.AgentRunPhaseFailed)
	if !strings.Contains(message, "RuntimeClassResolutionFailed") {
		t.Errorf("message = %q, want it to name the unresolved runtime class", message)
	}
}

func TestAgentRunReconcile_AppliesHarnessDefaults(t *testing.T) {
	harness := testutil.NewAgentHarness("default", "claude",
		testutil.WithHarnessType(agentv1alpha1.AgentTypeOpencode),
		testutil.WithDefaultImage("ghcr.io/xonovex/opencode:pinned"))
	run := runnableAgentRun("solo",
		testutil.WithHarnessRef("claude"),
		testutil.WithImage(""))
	reconciler, c := newRunReconciler(harness, run)

	reconcileRun(t, reconciler, "solo")

	var job batchv1.Job
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "solo",
	}, &job); err != nil {
		t.Fatalf("agent Job was not created: %v", err)
	}
	containers := job.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		t.Fatal("agent Job has no containers")
	}
	if containers[0].Image != "ghcr.io/xonovex/opencode:pinned" {
		t.Errorf("image = %q, want the harness default", containers[0].Image)
	}
}

// The harness supplies a default, not an override: a run naming its own image
// keeps it.
func TestAgentRunReconcile_RunImageOverridesTheHarnessDefault(t *testing.T) {
	harness := testutil.NewAgentHarness("default", "claude",
		testutil.WithDefaultImage("ghcr.io/xonovex/opencode:pinned"))
	run := runnableAgentRun("solo",
		testutil.WithHarnessRef("claude"),
		testutil.WithImage("ghcr.io/xonovex/explicit:chosen"))
	reconciler, c := newRunReconciler(harness, run)

	reconcileRun(t, reconciler, "solo")

	var job batchv1.Job
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "solo",
	}, &job); err != nil {
		t.Fatalf("agent Job was not created: %v", err)
	}
	containers := job.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		t.Fatal("agent Job has no containers")
	}
	if containers[0].Image != "ghcr.io/xonovex/explicit:chosen" {
		t.Errorf("image = %q, want the run's own image", containers[0].Image)
	}
}

// A run referencing a workspace waits rather than provisioning its own storage.
func TestAgentRunReconcile_WaitsForAWorkspaceThatIsNotReady(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "shared",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspacePhase(agentv1alpha1.AgentWorkspacePhaseInitializing))
	run := runnableAgentRun("attached", testutil.WithWorkspaceRef("shared"))
	reconciler, c := newRunReconciler(workspace, run)

	result := reconcileRun(t, reconciler, "attached")

	if result.RequeueAfter == 0 {
		t.Error("Reconcile() must requeue while the workspace is initializing")
	}
	resolved := getRun(t, c, "attached")
	if resolved.Status.WorkspacePVC != "" {
		t.Errorf("run adopted PVC %q before the workspace was ready",
			resolved.Status.WorkspacePVC)
	}
}

func TestAgentRunReconcile_AdoptsAReadyWorkspacePVC(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "shared",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspacePhase(agentv1alpha1.AgentWorkspacePhaseReady))
	workspace.Status.WorkspacePVC = "shared-ws"
	run := runnableAgentRun("attached", testutil.WithWorkspaceRef("shared"))
	reconciler, c := newRunReconciler(workspace, run)

	reconcileRun(t, reconciler, "attached")

	resolved := getRun(t, c, "attached")
	if resolved.Status.WorkspacePVC != "shared-ws" {
		t.Errorf("workspace PVC = %q, want the workspace's %q",
			resolved.Status.WorkspacePVC, "shared-ws")
	}
	var job batchv1.Job
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "default", Name: "attached",
	}, &job); err != nil {
		t.Errorf("agent Job was not created: %v", err)
	}
}

func TestAgentRunReconcile_FailsOnAnAbsentWorkspaceRef(t *testing.T) {
	run := runnableAgentRun("attached", testutil.WithWorkspaceRef("absent"))
	reconciler, c := newRunReconciler(run)

	reconcileRun(t, reconciler, "attached")

	resolved := getRun(t, c, "attached")
	if resolved.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
		t.Errorf("phase = %q, want %q", resolved.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
	}
	message := phaseMessage(resolved, agentv1alpha1.AgentRunPhaseFailed)
	if !strings.Contains(message, "WorkspaceResolutionFailed") {
		t.Errorf("message = %q, want it to name the workspace failure", message)
	}
}

func TestAgentRunReconcile_IgnoresAnAbsentRun(t *testing.T) {
	reconciler, _ := newRunReconciler()

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "absent"},
	})

	if err != nil {
		t.Fatalf("Reconcile() error = %v, want nil for a deleted run", err)
	}
	if result.RequeueAfter != 0 {
		t.Error("Reconcile() must not requeue for a deleted run")
	}
}

func TestAgentRunReconcile_SkipsARunBeingDeleted(t *testing.T) {
	run := runnableAgentRun("solo")
	deletionTimestamp := metav1.Now()
	run.DeletionTimestamp = &deletionTimestamp
	run.Finalizers = []string{"xonovex.com/test"}
	reconciler, c := newRunReconciler(run)

	reconcileRun(t, reconciler, "solo")

	resolved := getRun(t, c, "solo")
	if resolved.Status.WorkspacePVC != "" {
		t.Errorf("a run being deleted provisioned PVC %q", resolved.Status.WorkspacePVC)
	}
}

// Succeeded, Failed and TimedOut are terminal: a resync must not restart work.
func TestAgentRunReconcile_SkipsTerminalPhases(t *testing.T) {
	for _, phase := range []agentv1alpha1.AgentRunPhase{
		agentv1alpha1.AgentRunPhaseSucceeded,
		agentv1alpha1.AgentRunPhaseFailed,
		agentv1alpha1.AgentRunPhaseTimedOut,
	} {
		t.Run(string(phase), func(t *testing.T) {
			run := runnableAgentRun("solo", testutil.WithPhase(phase))
			reconciler, c := newRunReconciler(run)

			result := reconcileRun(t, reconciler, "solo")

			if result.RequeueAfter != 0 {
				t.Errorf("RequeueAfter = %v, want no requeue in a terminal phase",
					result.RequeueAfter)
			}
			resolved := getRun(t, c, "solo")
			if resolved.Status.WorkspacePVC != "" {
				t.Errorf("terminal run provisioned PVC %q", resolved.Status.WorkspacePVC)
			}
		})
	}
}

func TestAgentRunReconcile_IsIdempotentBeforeTheJobFinishes(t *testing.T) {
	reconciler, c := newRunReconciler(runnableAgentRun("solo"))

	reconcileRun(t, reconciler, "solo")
	first := getRun(t, c, "solo")
	reconcileRun(t, reconciler, "solo")
	second := getRun(t, c, "solo")

	if first.Status.WorkspacePVC != second.Status.WorkspacePVC {
		t.Errorf("workspace PVC changed across reconciles: %q then %q",
			first.Status.WorkspacePVC, second.Status.WorkspacePVC)
	}
	var jobs batchv1.JobList
	if err := c.List(context.Background(), &jobs); err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Items) != 1 {
		t.Errorf("job count = %d, want exactly one across two reconciles", len(jobs.Items))
	}
}
