//go:build e2e

package e2e

import (
	"fmt"
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

func createE2ENamespace(t *testing.T, prefix string) string {
	t.Helper()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix + "-",
		},
	}
	if err := k8sClient.Create(ctx, ns); err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}
	t.Cleanup(func() {
		_ = k8sClient.Delete(ctx, ns)
	})
	return ns.Name
}

func TestE2E_MultiAgentWorkspace(t *testing.T) {
	ns := createE2ENamespace(t, "e2e-ws")

	// 1. Create workspace
	ws := testutil.NewAgentWorkspace(ns, "shared-ws",
		testutil.WithWorkspaceStorageSize("1Gi"),
		testutil.WithSharedVolumes(
			agentv1alpha1.SharedVolumeSpec{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "256Mi"},
		),
	)
	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	// Wait for init job to be created
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var w agentv1alpha1.AgentWorkspace
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), &w); err != nil {
			return false
		}
		return w.Status.InitJobName != ""
	})

	// Simulate workspace init job completion (in e2e the git clone may fail,
	// so we manually complete it for the workspace test)
	now := metav1.Now()
	var initJob batchv1.Job
	initJobKey := types.NamespacedName{Name: "shared-ws-init", Namespace: ns}
	if err := k8sClient.Get(ctx, initJobKey, &initJob); err != nil {
		t.Fatalf("init Job not created: %v", err)
	}
	initJob.Status.StartTime = &now
	initJob.Status.CompletionTime = &now
	initJob.Status.Conditions = append(initJob.Status.Conditions,
		batchv1.JobCondition{
			Type:   batchv1.JobSuccessCriteriaMet,
			Status: corev1.ConditionTrue,
		},
		batchv1.JobCondition{
			Type:   batchv1.JobComplete,
			Status: corev1.ConditionTrue,
		},
	)
	if err := k8sClient.Status().Update(ctx, &initJob); err != nil {
		t.Fatalf("failed to update init Job status: %v", err)
	}

	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseReady, 60*time.Second)

	// 2. Create 2 concurrent AgentRuns with workspaceRef
	for i := 1; i <= 2; i++ {
		run := testutil.NewAgentRun(ns, fmt.Sprintf("agent-%d", i),
			testutil.WithWorkspaceRef("shared-ws"),
			testutil.WithWorktree(fmt.Sprintf("agent-%d-work", i), ""),
		)
		if err := k8sClient.Create(ctx, run); err != nil {
			t.Fatalf("failed to create AgentRun agent-%d: %v", i, err)
		}
	}

	// 3. Wait for both AgentRuns to create their Jobs
	for i := 1; i <= 2; i++ {
		name := fmt.Sprintf("agent-%d", i)
		key := types.NamespacedName{Name: name, Namespace: ns}
		testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
			var r agentv1alpha1.AgentRun
			if err := k8sClient.Get(ctx, key, &r); err != nil {
				return false
			}
			return r.Status.JobName != ""
		})
	}

	// 4. Simulate both Jobs completing
	for i := 1; i <= 2; i++ {
		name := fmt.Sprintf("agent-%d", i)
		jobKey := types.NamespacedName{Name: name, Namespace: ns}
		var job batchv1.Job
		if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
			t.Fatalf("Job %s not created: %v", name, err)
		}

		jobNow := metav1.Now()
		job.Status.StartTime = &jobNow
		job.Status.CompletionTime = &jobNow
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
			t.Fatalf("failed to update Job %s status: %v", name, err)
		}
	}

	// 5. Verify both reach Succeeded
	for i := 1; i <= 2; i++ {
		name := fmt.Sprintf("agent-%d", i)
		key := types.NamespacedName{Name: name, Namespace: ns}
		testutil.WaitForAgentRunPhase(t, ctx, k8sClient, key, agentv1alpha1.AgentRunPhaseSucceeded, 60*time.Second)
	}

	// 6. Verify both used the shared workspace PVC
	for i := 1; i <= 2; i++ {
		name := fmt.Sprintf("agent-%d", i)
		var run agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, &run); err != nil {
			t.Fatalf("failed to get AgentRun %s: %v", name, err)
		}
		if run.Status.WorkspacePVC != "shared-ws-ws" {
			t.Errorf("agent-%d WorkspacePVC = %q, want shared-ws-ws", i, run.Status.WorkspacePVC)
		}
	}
}
