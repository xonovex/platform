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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

// runnableAgentRun is an AgentRun as admission leaves it: image and sandboxed
// runtimeClassName already resolved, so the reconcile reaches the Job.
func runnableAgentRun(name string, opts ...testutil.AgentRunOption) *agentv1alpha1.AgentRun {
	run := testutil.NewAgentRun("default", name,
		append([]testutil.AgentRunOption{testutil.WithPrompt("do the thing")}, opts...)...)
	run.UID = types.UID(name + "-uid")
	return run
}

func runKey(name string) types.NamespacedName {
	return types.NamespacedName{Namespace: "default", Name: name}
}

func reconcileRun(t *testing.T, h runHarness, name string) ctrl.Result {
	t.Helper()
	result, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey(name)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	return result
}

func reconcileRunError(h runHarness, name string) error {
	_, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: runKey(name)})
	return err
}

func getRun(t *testing.T, h runHarness, name string) *agentv1alpha1.AgentRun {
	t.Helper()
	run, err := h.store.GetAgentRun(context.Background(), runKey(name))
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	return run
}

func getJob(t *testing.T, store *testutil.MemoryStore, name string) *batchv1.Job {
	t.Helper()
	job, err := store.GetJob(context.Background(), runKey(name))
	if err != nil {
		t.Fatalf("get Job %q: %v", name, err)
	}
	return job
}

// phaseMessage returns the message of the condition a phase change adds.
func phaseMessage(run *agentv1alpha1.AgentRun, phase agentv1alpha1.AgentRunPhase) string {
	for _, condition := range run.Status.Conditions {
		if condition.Type == string(phase) {
			return condition.Message
		}
	}
	return ""
}

// startedRun is a run whose Job exists, so a reconcile only observes the Job.
func startedRun(name string, opts ...testutil.AgentRunOption) *agentv1alpha1.AgentRun {
	run := runnableAgentRun(name, opts...)
	run.Status.WorkspacePVC = name + "-workspace"
	run.Status.JobName = name
	run.Status.Phase = agentv1alpha1.AgentRunPhaseInitializing
	return run
}

func runJob(run *agentv1alpha1.AgentRun, status batchv1.JobStatus) *batchv1.Job {
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: run.Name, Namespace: run.Namespace}, Status: status}
	setController(run, agentRunKind, job)
	return job
}

// A run with no workspaceRef provisions its own PVC, the agent ServiceAccount,
// its NetworkPolicy and its Job.
func TestAgentRunReconcile_StandaloneProvisionsItsResources(t *testing.T) {
	h := newRunHarness(runnableAgentRun("solo"))

	reconcileRun(t, h, "solo")

	run := getRun(t, h, "solo")
	if run.Status.WorkspacePVC != "solo-workspace" || run.Status.JobName != "solo" {
		t.Errorf("status = %+v, want PVC solo-workspace and Job solo", run.Status)
	}
	if run.Status.Phase != agentv1alpha1.AgentRunPhaseInitializing {
		t.Errorf("phase = %q, want %q", run.Status.Phase, agentv1alpha1.AgentRunPhaseInitializing)
	}
	ctx := context.Background()
	if _, err := h.store.GetPersistentVolumeClaim(ctx, runKey("solo-workspace")); err != nil {
		t.Errorf("workspace PVC was not created: %v", err)
	}
	if _, err := h.store.GetServiceAccount(ctx, runKey(isoshared.AgentServiceAccountName)); err != nil {
		t.Errorf("agent ServiceAccount was not created: %v", err)
	}
	policy, err := h.store.GetNetworkPolicy(ctx, runKey("solo-netpol"))
	if err != nil {
		t.Fatalf("NetworkPolicy was not created: %v", err)
	}
	if owner := metav1.GetControllerOf(policy); owner == nil || owner.UID != "solo-uid" || owner.Kind != agentRunKind {
		t.Errorf("NetworkPolicy controller = %+v, want the run", owner)
	}
	job := getJob(t, h.store, "solo")
	if owner := metav1.GetControllerOf(job); owner == nil || owner.UID != "solo-uid" {
		t.Errorf("Job controller = %+v, want the run", owner)
	}
	if got := job.Spec.Template.Spec.Containers[0].Command; !slices.Equal(got, []string{"claude"}) {
		t.Errorf("agent command = %v, want the harness command", got)
	}
	if want := []string{"NetworkPolicyCreated", "AgentRunStarted"}; !slices.Equal(h.events.reasons(), want) {
		t.Errorf("events = %v, want %v", h.events.reasons(), want)
	}
}

func TestAgentRunReconcile_StandaloneHonoursWorkspaceStorageSettings(t *testing.T) {
	h := newRunHarness(runnableAgentRun("solo", testutil.WithWorkspace(&agentv1alpha1.WorkspaceSpec{
		Type:         agentv1alpha1.WorkspaceTypeJujutsu,
		StorageClass: "fast",
		StorageSize:  "42Gi",
	})))

	reconcileRun(t, h, "solo")

	pvc, err := h.store.GetPersistentVolumeClaim(context.Background(), runKey("solo-workspace"))
	if err != nil {
		t.Fatalf("workspace PVC was not created: %v", err)
	}
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "fast" {
		t.Errorf("storage class = %v, want %q", pvc.Spec.StorageClassName, "fast")
	}
	if got := pvc.Spec.Resources.Requests.Storage().String(); got != "42Gi" {
		t.Errorf("storage size = %q, want %q", got, "42Gi")
	}
	initScript := strings.Join(getJob(t, h.store, "solo").Spec.Template.Spec.InitContainers[0].Args, " ")
	if !strings.Contains(initScript, "post-clone jj") {
		t.Errorf("init script = %q, want the jj strategy", initScript)
	}
}

// Admission resolves the image and the runtime class; the controller refuses
// to guess either, because agent pods run untrusted model output.
func TestAgentRunReconcile_FailsWhatAdmissionDidNotResolve(t *testing.T) {
	tests := []struct {
		name   string
		option testutil.AgentRunOption
		reason string
	}{
		{name: "image", option: testutil.WithImage(""), reason: "ImageResolutionFailed"},
		{name: "runtime class", option: testutil.WithRuntimeClassName(""), reason: "RuntimeClassResolutionFailed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newRunHarness(runnableAgentRun("solo", test.option))

			reconcileRun(t, h, "solo")

			run := getRun(t, h, "solo")
			if run.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
				t.Errorf("phase = %q, want %q", run.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
			}
			if message := phaseMessage(run, agentv1alpha1.AgentRunPhaseFailed); !strings.Contains(message, test.reason) {
				t.Errorf("message = %q, want %q", message, test.reason)
			}
			if condition := run.Status.Conditions[0]; !condition.LastTransitionTime.Time.Equal(testEpoch) {
				t.Errorf("condition time = %v, want the clock time", condition.LastTransitionTime)
			}
			if h.store.JobCount() != 0 || h.store.ClaimCount() != 0 {
				t.Error("a failed resolution created resources")
			}
		})
	}
}

func TestAgentRunReconcile_ResolutionFailureMarksRunFailed(t *testing.T) {
	readyWorkspace := testutil.NewAgentWorkspace("default", "shared",
		testutil.WithWorkspaceRepository("https://github.com/example/repo.git"),
		testutil.WithWorkspacePhase(agentv1alpha1.AgentWorkspacePhaseReady))
	readyWorkspace.Status.WorkspacePVC = "shared-ws"
	tests := []struct {
		name      string
		workspace bool
		option    testutil.AgentRunOption
		reason    string
	}{
		{name: "standalone harness", option: testutil.WithHarnessRef("missing"), reason: "HarnessResolutionFailed"},
		{name: "standalone provider", option: testutil.WithProviderRef("missing"), reason: "ProviderResolutionFailed"},
		{name: "standalone toolchain", option: testutil.WithToolchainRef("missing"), reason: "ToolchainResolutionFailed"},
		{name: "workspace harness", workspace: true, option: testutil.WithHarnessRef("missing"), reason: "HarnessResolutionFailed"},
		{name: "workspace toolchain", workspace: true, option: testutil.WithToolchainRef("missing"), reason: "ToolchainResolutionFailed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := runnableAgentRun("run", test.option)
			objects := []any{run}
			if test.workspace {
				run.Spec.Workspace = nil
				run.Spec.WorkspaceRef = "shared"
				objects = append(objects, readyWorkspace)
			}
			h := newRunHarness(objects...)

			reconcileRun(t, h, "run")

			updated := getRun(t, h, "run")
			if updated.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
				t.Fatalf("phase = %q, want %q", updated.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
			}
			if len(updated.Status.Conditions) != 1 || !strings.Contains(updated.Status.Conditions[0].Message, test.reason) {
				t.Fatalf("conditions = %#v, want one containing %q", updated.Status.Conditions, test.reason)
			}
		})
	}
}

// A harness type the registry does not know stops the run before any Job.
func TestAgentRunReconcile_FailsWhenTheJobCannotBeBuilt(t *testing.T) {
	h := newRunHarness(runnableAgentRun("solo", testutil.WithHarness(&agentv1alpha1.AgentSpec{Type: "mystery"})))

	reconcileRun(t, h, "solo")

	run := getRun(t, h, "solo")
	if message := phaseMessage(run, agentv1alpha1.AgentRunPhaseFailed); !strings.Contains(message, "JobBuildFailed") {
		t.Errorf("message = %q, want JobBuildFailed", message)
	}
	if h.store.JobCount() != 0 {
		t.Error("a Job was created for an unknown harness")
	}
}

func TestAgentRunReconcile_AppliesHarnessDefaults(t *testing.T) {
	harness := testutil.NewAgentHarness("default", "claude",
		testutil.WithHarnessType(agentv1alpha1.AgentTypeOpencode),
		testutil.WithDefaultImage("ghcr.io/xonovex/opencode:pinned"))
	h := newRunHarness(harness, runnableAgentRun("solo", testutil.WithHarnessRef("claude"), testutil.WithImage("")))

	reconcileRun(t, h, "solo")

	container := getJob(t, h.store, "solo").Spec.Template.Spec.Containers[0]
	if container.Image != "ghcr.io/xonovex/opencode:pinned" {
		t.Errorf("image = %q, want the harness default", container.Image)
	}
	if !slices.Equal(container.Command, []string{"opencode"}) {
		t.Errorf("command = %v, want the harness type", container.Command)
	}
}

// The harness supplies a default, not an override: a run naming its own image
// keeps it.
func TestAgentRunReconcile_RunImageOverridesTheHarnessDefault(t *testing.T) {
	harness := testutil.NewAgentHarness("default", "claude", testutil.WithDefaultImage("ghcr.io/xonovex/opencode:pinned"))
	h := newRunHarness(harness, runnableAgentRun("solo",
		testutil.WithHarnessRef("claude"), testutil.WithImage("ghcr.io/xonovex/explicit:chosen")))

	reconcileRun(t, h, "solo")

	if image := getJob(t, h.store, "solo").Spec.Template.Spec.Containers[0].Image; image != "ghcr.io/xonovex/explicit:chosen" {
		t.Errorf("image = %q, want the run's own image", image)
	}
}

// A toolchain is a pre-built image, so it replaces the run image.
func TestAgentRunReconcile_TheToolchainImageReplacesTheRunImage(t *testing.T) {
	h := newRunHarness(runnableAgentRun("solo", testutil.WithToolchain(&agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Image: "ghcr.io/xonovex/toolchain@sha256:bbbb"},
	})))

	reconcileRun(t, h, "solo")

	if image := getJob(t, h.store, "solo").Spec.Template.Spec.Containers[0].Image; image != "ghcr.io/xonovex/toolchain@sha256:bbbb" {
		t.Errorf("image = %q, want the toolchain image", image)
	}
}

func TestAgentRunReconcile_ADisabledNetworkPolicyCreatesNone(t *testing.T) {
	run := runnableAgentRun("solo")
	run.Spec.NetworkPolicy = &agentv1alpha1.AgentNetworkPolicy{Disabled: true}
	h := newRunHarness(run)

	reconcileRun(t, h, "solo")

	if _, err := h.store.GetNetworkPolicy(context.Background(), runKey("solo-netpol")); !apierrors.IsNotFound(err) {
		t.Errorf("GetNetworkPolicy() error = %v, want NotFound", err)
	}
}

// A run referencing a workspace waits rather than provisioning its own storage.
func TestAgentRunReconcile_WaitsForAWorkspaceThatIsNotReady(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "shared",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspacePhase(agentv1alpha1.AgentWorkspacePhaseInitializing))
	h := newRunHarness(workspace, runnableAgentRun("attached", testutil.WithWorkspaceRef("shared")))

	result := reconcileRun(t, h, "attached")

	if result.RequeueAfter != pendingRequeue {
		t.Errorf("RequeueAfter = %v, want %v while the workspace initializes", result.RequeueAfter, pendingRequeue)
	}
	if pvc := getRun(t, h, "attached").Status.WorkspacePVC; pvc != "" {
		t.Errorf("run adopted PVC %q before the workspace was ready", pvc)
	}
}

func TestAgentRunReconcile_MountsAReadyWorkspace(t *testing.T) {
	workspace := testutil.NewAgentWorkspace("default", "shared",
		testutil.WithWorkspaceRepository("https://github.com/org/repo.git"),
		testutil.WithWorkspaceType(agentv1alpha1.WorkspaceTypeJujutsu),
		testutil.WithWorkspacePhase(agentv1alpha1.AgentWorkspacePhaseReady))
	workspace.Status.WorkspacePVC = "shared-ws"
	h := newRunHarness(workspace, runnableAgentRun("attached", testutil.WithWorkspaceRef("shared")))

	reconcileRun(t, h, "attached")

	if pvc := getRun(t, h, "attached").Status.WorkspacePVC; pvc != "shared-ws" {
		t.Errorf("workspace PVC = %q, want the workspace's %q", pvc, "shared-ws")
	}
	if h.store.ClaimCount() != 0 {
		t.Error("a run on a shared workspace created its own PVC")
	}
	initContainers := getJob(t, h.store, "attached").Spec.Template.Spec.InitContainers
	if len(initContainers) == 0 || initContainers[0].Name != "jj-worktree" {
		t.Errorf("init containers = %v, want the jj worktree", initContainers)
	}
}

func TestAgentRunReconcile_FailsOnAnAbsentWorkspaceRef(t *testing.T) {
	h := newRunHarness(runnableAgentRun("attached", testutil.WithWorkspaceRef("absent")))

	reconcileRun(t, h, "attached")

	run := getRun(t, h, "attached")
	if run.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
		t.Errorf("phase = %q, want %q", run.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
	}
	if message := phaseMessage(run, agentv1alpha1.AgentRunPhaseFailed); !strings.Contains(message, "WorkspaceResolutionFailed") {
		t.Errorf("message = %q, want it to name the workspace failure", message)
	}
}

func TestAgentRunReconcile_IgnoresAnAbsentRun(t *testing.T) {
	result := reconcileRun(t, newRunHarness(), "absent")

	if result.RequeueAfter != 0 {
		t.Error("Reconcile() must not requeue for a deleted run")
	}
}

func TestAgentRunReconcile_SurfacesARunReadFailure(t *testing.T) {
	h := newRunHarness()
	h.store.Fail("GetAgentRun", errInjected)

	if err := reconcileRunError(h, "solo"); !errors.Is(err, errInjected) {
		t.Errorf("Reconcile() error = %v, want the injected failure", err)
	}
}

func TestAgentRunReconcile_SkipsARunBeingDeleted(t *testing.T) {
	run := runnableAgentRun("solo")
	deletionTimestamp := metav1.NewTime(testEpoch)
	run.DeletionTimestamp = &deletionTimestamp
	h := newRunHarness(run)

	reconcileRun(t, h, "solo")

	if h.store.StatusWrites() != 0 || h.store.ClaimCount() != 0 {
		t.Error("a run being deleted was reconciled")
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
			h := newRunHarness(runnableAgentRun("solo", testutil.WithPhase(phase)))

			result := reconcileRun(t, h, "solo")

			if result.RequeueAfter != 0 {
				t.Errorf("RequeueAfter = %v, want no requeue in a terminal phase", result.RequeueAfter)
			}
			if h.store.StatusWrites() != 0 || h.store.ClaimCount() != 0 {
				t.Error("a terminal run was reconciled")
			}
		})
	}
}

func TestAgentRunReconcile_IsIdempotentBeforeTheJobFinishes(t *testing.T) {
	h := newRunHarness(runnableAgentRun("solo"))

	reconcileRun(t, h, "solo")
	first := getRun(t, h, "solo")
	result := reconcileRun(t, h, "solo")
	second := getRun(t, h, "solo")

	if first.Status.WorkspacePVC != second.Status.WorkspacePVC || first.Status.JobName != second.Status.JobName {
		t.Errorf("status changed across reconciles: %+v then %+v", first.Status, second.Status)
	}
	if h.store.JobCount() != 1 || h.store.ClaimCount() != 1 {
		t.Errorf("Jobs = %d, PVCs = %d, want one of each", h.store.JobCount(), h.store.ClaimCount())
	}
	if result.RequeueAfter != runningRequeue {
		t.Errorf("RequeueAfter = %v, want %v while the Job runs", result.RequeueAfter, runningRequeue)
	}
}

// A Job of the same name that the run does not control is a collision the
// controller must refuse to adopt.
func TestAgentRunReconcile_RejectsAnUnownedExistingJob(t *testing.T) {
	foreign := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "solo", Namespace: "default"}}
	h := newRunHarness(runnableAgentRun("solo"), foreign)

	err := reconcileRunError(h, "solo")

	if err == nil || !strings.Contains(err.Error(), "not controlled by AgentRun") {
		t.Fatalf("Reconcile() error = %v, want an ownership error", err)
	}
}

func TestAgentRunReconcile_RejectsAnUnownedExistingNetworkPolicy(t *testing.T) {
	foreign := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "solo-netpol", Namespace: "default"}}
	h := newRunHarness(runnableAgentRun("solo"), foreign)

	err := reconcileRunError(h, "solo")

	if err == nil || !strings.Contains(err.Error(), "not controlled by AgentRun") {
		t.Fatalf("Reconcile() error = %v, want an ownership error", err)
	}
	if h.store.JobCount() != 0 {
		t.Error("a Job was created behind a foreign NetworkPolicy")
	}
}

// Each API failure that is neither NotFound nor AlreadyExists reaches the
// caller, so controller-runtime retries with backoff.
func TestAgentRunReconcile_SurfacesStoreFailures(t *testing.T) {
	foreignPolicy := &networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "solo-netpol", Namespace: "default"}}
	tests := []struct {
		method  string
		run     *agentv1alpha1.AgentRun
		objects []any
		phrase  string
	}{
		{method: "CreatePersistentVolumeClaim", run: runnableAgentRun("solo"), phrase: "create PersistentVolumeClaim"},
		{method: "UpdateAgentRunStatus", run: runnableAgentRun("solo")},
		{method: "CreateNetworkPolicy", run: runnableAgentRun("solo"), phrase: "create NetworkPolicy"},
		{method: "GetNetworkPolicy", run: runnableAgentRun("solo"), objects: []any{foreignPolicy}, phrase: "get existing NetworkPolicy"},
		{method: "CreateServiceAccount", run: runnableAgentRun("solo"), phrase: "create agent ServiceAccount"},
		{method: "CreateJob", run: runnableAgentRun("solo"), phrase: "create Job"},
		{method: "GetJob", run: startedRun("solo")},
		{method: "ListJobPods", run: startedRun("solo"), objects: []any{runJob(startedRun("solo"), batchv1.JobStatus{})}, phrase: "list pods for Job"},
	}
	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			h := newRunHarness(append([]any{test.run}, test.objects...)...)
			h.store.Fail(test.method, errInjected)

			err := reconcileRunError(h, "solo")

			if !errors.Is(err, errInjected) {
				t.Fatalf("Reconcile() error = %v, want the injected failure", err)
			}
			if !strings.Contains(err.Error(), test.phrase) {
				t.Errorf("error = %v, want it to name %q", err, test.phrase)
			}
		})
	}
}

// The Job read lags its create in an informer cache, so a missing Job waits.
func TestAgentRunReconcile_WaitsForTheJobToBeVisible(t *testing.T) {
	h := newRunHarness(startedRun("solo"))

	result := reconcileRun(t, h, "solo")

	if result.RequeueAfter != pendingRequeue {
		t.Errorf("RequeueAfter = %v, want %v", result.RequeueAfter, pendingRequeue)
	}
}

func TestAgentRunReconcile_MapsTheJobStateToThePhase(t *testing.T) {
	timeout := testutil.WithTimeout(metav1.Duration{Duration: time.Second})
	tests := []struct {
		name      string
		run       *agentv1alpha1.AgentRun
		job       batchv1.JobStatus
		elapsed   time.Duration
		want      agentv1alpha1.AgentRunPhase
		event     string
		completed bool
	}{
		{
			name:      "completed",
			run:       startedRun("solo"),
			job:       batchv1.JobStatus{Conditions: []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}},
			want:      agentv1alpha1.AgentRunPhaseSucceeded,
			event:     "AgentRunSucceeded",
			completed: true,
		},
		{
			name:      "failed",
			run:       startedRun("solo"),
			job:       batchv1.JobStatus{Conditions: []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "backoff"}}},
			want:      agentv1alpha1.AgentRunPhaseFailed,
			event:     "AgentRunFailed",
			completed: true,
		},
		{
			name: "started",
			run:  startedRun("solo"),
			job:  batchv1.JobStatus{Active: 1},
			want: agentv1alpha1.AgentRunPhaseRunning,
		},
		{
			name:      "timed out",
			run:       runningRun("solo", timeout),
			job:       batchv1.JobStatus{Active: 1},
			elapsed:   2 * time.Second,
			want:      agentv1alpha1.AgentRunPhaseTimedOut,
			event:     "AgentRunTimedOut",
			completed: true,
		},
		{
			name: "within the timeout",
			run:  runningRun("solo", timeout),
			job:  batchv1.JobStatus{Active: 1},
			want: agentv1alpha1.AgentRunPhaseRunning,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := newRunHarness(test.run, runJob(test.run, test.job))
			h.clock.advance(test.elapsed)

			reconcileRun(t, h, "solo")

			run := getRun(t, h, "solo")
			if run.Status.Phase != test.want {
				t.Fatalf("phase = %q, want %q", run.Status.Phase, test.want)
			}
			if test.completed != (run.Status.CompletionTime != nil) {
				t.Errorf("CompletionTime = %v, want set %v", run.Status.CompletionTime, test.completed)
			}
			if test.completed && !run.Status.CompletionTime.Time.Equal(h.clock.Now()) {
				t.Errorf("CompletionTime = %v, want the clock time", run.Status.CompletionTime)
			}
			if reasons := h.events.reasons(); test.event != "" && reasons[len(reasons)-1] != test.event {
				t.Errorf("events = %v, want %s last", reasons, test.event)
			}
		})
	}
}

// runningRun is a started run in the Running phase since the clock epoch.
func runningRun(name string, opts ...testutil.AgentRunOption) *agentv1alpha1.AgentRun {
	run := startedRun(name, opts...)
	run.Status.Phase = agentv1alpha1.AgentRunPhaseRunning
	started := metav1.NewTime(testEpoch)
	run.Status.StartTime = &started
	return run
}

func TestAgentRunReconcile_RecordsTheRunningStartTime(t *testing.T) {
	run := startedRun("solo")
	h := newRunHarness(run, runJob(run, batchv1.JobStatus{Active: 1}))

	reconcileRun(t, h, "solo")

	if start := getRun(t, h, "solo").Status.StartTime; start == nil || !start.Time.Equal(testEpoch) {
		t.Errorf("StartTime = %v, want the clock time", start)
	}
}

func TestAgentRunReconcile_RecordsThePodAndTheAgentExitCode(t *testing.T) {
	run := startedRun("solo")
	exitCode := int32(17)
	h := newRunHarness(run, runJob(run, batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
		Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "failed",
	}}}), agentPod("solo-pod", testEpoch, &exitCode))

	reconcileRun(t, h, "solo")

	updated := getRun(t, h, "solo")
	if updated.Status.PodName != "solo-pod" {
		t.Errorf("PodName = %q, want solo-pod", updated.Status.PodName)
	}
	if updated.Status.ExitCode == nil || *updated.Status.ExitCode != exitCode {
		t.Errorf("ExitCode = %v, want %d", updated.Status.ExitCode, exitCode)
	}
}

// A resync over an unchanged running Job writes no status.
func TestAgentRunReconcile_WritesNoStatusWithoutAChange(t *testing.T) {
	run := runningRun("solo")
	run.Status.PodName = "solo-pod"
	h := newRunHarness(run, runJob(run, batchv1.JobStatus{Active: 1}), agentPod("solo-pod", testEpoch, nil))

	result := reconcileRun(t, h, "solo")

	if h.store.StatusWrites() != 0 {
		t.Errorf("status writes = %d, want none", h.store.StatusWrites())
	}
	if result.RequeueAfter != runningRequeue {
		t.Errorf("RequeueAfter = %v, want %v", result.RequeueAfter, runningRequeue)
	}
}

func agentPod(name string, created time.Time, exitCode *int32) *corev1.Pod {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:              name,
		Namespace:         "default",
		Labels:            map[string]string{"job-name": "solo"},
		CreationTimestamp: metav1.NewTime(created),
	}}
	if exitCode != nil {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
			Name:  "agent",
			State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: *exitCode}},
		}}
	}
	return pod
}
