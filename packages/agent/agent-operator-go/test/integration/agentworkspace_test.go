//go:build integration

package integration

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestAgentWorkspace_CreatesPVCsAndTransitionsToPending(t *testing.T) {
	ns := createNamespace(t, "ws-pending")
	ws := testutil.NewAgentWorkspace(ns, "test-ws",
		testutil.WithSharedVolumes(
			agentv1alpha1.SharedVolumeSpec{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "1Gi"},
		),
	)

	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	// Wait for Initializing phase (PVCs created, init job created)
	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseInitializing, 30*time.Second)

	// Check workspace PVC exists
	var pvc corev1.PersistentVolumeClaim
	pvcKey := types.NamespacedName{Name: "test-ws-ws", Namespace: ns}
	if err := k8sClient.Get(ctx, pvcKey, &pvc); err != nil {
		t.Fatalf("workspace PVC not created: %v", err)
	}
	if pvc.OwnerReferences[0].Kind != "AgentWorkspace" {
		t.Errorf("workspace PVC owner kind = %q, want AgentWorkspace", pvc.OwnerReferences[0].Kind)
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteMany {
		t.Errorf("expected ReadWriteMany access mode, got %v", pvc.Spec.AccessModes)
	}

	// Check shared volume PVC exists
	var sharedPVC corev1.PersistentVolumeClaim
	sharedPVCKey := types.NamespacedName{Name: "test-ws-claude-config", Namespace: ns}
	if err := k8sClient.Get(ctx, sharedPVCKey, &sharedPVC); err != nil {
		t.Fatalf("shared volume PVC not created: %v", err)
	}
	if sharedPVC.OwnerReferences[0].Kind != "AgentWorkspace" {
		t.Errorf("shared PVC owner kind = %q, want AgentWorkspace", sharedPVC.OwnerReferences[0].Kind)
	}

	// Verify status
	var updated agentv1alpha1.AgentWorkspace
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), &updated); err != nil {
		t.Fatalf("failed to get AgentWorkspace: %v", err)
	}
	if updated.Status.WorkspacePVC != "test-ws-ws" {
		t.Errorf("WorkspacePVC = %q, want test-ws-ws", updated.Status.WorkspacePVC)
	}
	if updated.Status.SharedVolumePVCs["claude-config"] != "test-ws-claude-config" {
		t.Errorf("SharedVolumePVCs[claude-config] = %q, want test-ws-claude-config", updated.Status.SharedVolumePVCs["claude-config"])
	}
}

func TestAgentWorkspace_TransitionsToReadyOnJobComplete(t *testing.T) {
	ns := createNamespace(t, "ws-ready")
	ws := testutil.NewAgentWorkspace(ns, "test-ws")

	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	// Wait for init job to be created
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var w agentv1alpha1.AgentWorkspace
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), &w); err != nil {
			return false
		}
		return w.Status.InitJobName != ""
	})

	// Simulate init job completion
	now := metav1.Now()
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-ws-init", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("init Job not created: %v", err)
	}
	job.Status.StartTime = &now
	job.Status.CompletionTime = &now
	job.Status.Conditions = append(job.Status.Conditions,
		batchv1.JobCondition{
			Type:   batchv1.JobSuccessCriteriaMet,
			Status: corev1.ConditionTrue,
		},
		batchv1.JobCondition{
			Type:   batchv1.JobComplete,
			Status: corev1.ConditionTrue,
		},
	)
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseReady, 30*time.Second)
}

func TestAgentWorkspace_TransitionsToFailedOnJobFailure(t *testing.T) {
	ns := createNamespace(t, "ws-failed")
	ws := testutil.NewAgentWorkspace(ns, "test-ws")

	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var w agentv1alpha1.AgentWorkspace
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), &w); err != nil {
			return false
		}
		return w.Status.InitJobName != ""
	})

	// Simulate init job failure
	now := metav1.Now()
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-ws-init", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("init Job not created: %v", err)
	}
	job.Status.StartTime = &now
	job.Status.Conditions = append(job.Status.Conditions,
		batchv1.JobCondition{
			Type:   batchv1.JobFailureTarget,
			Status: corev1.ConditionTrue,
		},
		batchv1.JobCondition{
			Type:    batchv1.JobFailed,
			Status:  corev1.ConditionTrue,
			Message: "git clone failed",
		},
	)
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseFailed, 30*time.Second)
}

func TestAgentRun_WithWorkspaceRef_WaitsForWorkspaceReady(t *testing.T) {
	ns := createNamespace(t, "ws-run-wait")

	// Create workspace (it will be in Pending/Initializing phase)
	ws := testutil.NewAgentWorkspace(ns, "test-ws")
	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	// Wait for workspace to reach Initializing
	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseInitializing, 30*time.Second)

	// Create AgentRun with workspaceRef
	run := testutil.NewAgentRun(ns, "agent-1",
		testutil.WithWorkspaceRef("test-ws"),
		testutil.WithWorktree("agent-1-work", ""),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// AgentRun should not have a Job yet (workspace not ready)
	time.Sleep(2 * time.Second)
	var updated agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if updated.Status.JobName != "" {
		t.Error("expected no Job before workspace is Ready")
	}

	// Complete the workspace init job
	now := metav1.Now()
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-ws-init", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("init Job not created: %v", err)
	}
	job.Status.StartTime = &now
	job.Status.CompletionTime = &now
	job.Status.Conditions = append(job.Status.Conditions,
		batchv1.JobCondition{
			Type:   batchv1.JobSuccessCriteriaMet,
			Status: corev1.ConditionTrue,
		},
		batchv1.JobCondition{
			Type:   batchv1.JobComplete,
			Status: corev1.ConditionTrue,
		},
	)
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	// Wait for workspace to be Ready
	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseReady, 30*time.Second)

	// Now the AgentRun should create a Job
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Verify the workspace Job has worktree init container
	var wsJob batchv1.Job
	wsJobKey := types.NamespacedName{Name: "agent-1", Namespace: ns}
	if err := k8sClient.Get(ctx, wsJobKey, &wsJob); err != nil {
		t.Fatalf("workspace Job not created: %v", err)
	}
	if len(wsJob.Spec.Template.Spec.InitContainers) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(wsJob.Spec.Template.Spec.InitContainers))
	}
	if wsJob.Spec.Template.Spec.InitContainers[0].Name != "git-worktree" {
		t.Errorf("expected init container name git-worktree, got %s", wsJob.Spec.Template.Spec.InitContainers[0].Name)
	}

	// Verify workspace PVC is used (not an individual one)
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if updated.Status.WorkspacePVC != "test-ws-ws" {
		t.Errorf("WorkspacePVC = %q, want test-ws-ws", updated.Status.WorkspacePVC)
	}
}

func TestAgentRun_WithoutWorkspaceRef_BackwardCompat(t *testing.T) {
	ns := createNamespace(t, "ws-compat")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseInitializing, 30*time.Second)

	// Verify it uses standalone PVC (not workspace)
	var pvc corev1.PersistentVolumeClaim
	pvcKey := types.NamespacedName{Name: "test-run-workspace", Namespace: ns}
	if err := k8sClient.Get(ctx, pvcKey, &pvc); err != nil {
		t.Fatalf("standalone PVC not created: %v", err)
	}
	if pvc.OwnerReferences[0].Kind != "AgentRun" {
		t.Errorf("PVC owner kind = %q, want AgentRun", pvc.OwnerReferences[0].Kind)
	}
}
