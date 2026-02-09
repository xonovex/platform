//go:build e2e

package e2e

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

func createNamespace(t *testing.T, prefix string) string {
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

func TestE2E_JobActuallyScheduled(t *testing.T) {
	ns := createNamespace(t, "e2e-scheduled")

	// Use busybox as the image — the init container (git clone) will fail
	// because busybox lacks git, but the Job/Pod will be scheduled.
	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithImage("busybox:1.37"),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for Job creation
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Wait for the Pod to be created (scheduled on a real node)
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "test-run",
		}); err != nil {
			return false
		}
		if len(podList.Items) == 0 {
			return false
		}
		pod := podList.Items[0]
		// Pod is scheduled once it has a node name
		return pod.Spec.NodeName != ""
	})
}

func TestE2E_PVCBinds(t *testing.T) {
	ns := createNamespace(t, "e2e-pvc-bind")

	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithImage("busybox:1.37"),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Kind uses WaitForFirstConsumer — PVC only binds after a Pod consuming it is scheduled.
	// Wait for a Pod to be scheduled first (ensures the PVC consumer exists).
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "test-run",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})

	// Now the PVC should be Bound (local-path-provisioner provisions on Pod scheduling)
	pvcKey := types.NamespacedName{Name: "test-run-workspace", Namespace: ns}
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var pvc corev1.PersistentVolumeClaim
		if err := k8sClient.Get(ctx, pvcKey, &pvc); err != nil {
			return false
		}
		return pvc.Status.Phase == corev1.ClaimBound
	})
}

func TestE2E_InitContainerFailurePath(t *testing.T) {
	ns := createNamespace(t, "e2e-init-fail")

	// busybox lacks git, so the init container's git clone will fail
	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithImage("busybox:1.37"),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for Job to fail (the init container failure causes Job failure)
	testutil.WaitForCondition(t, ctx, 180*time.Second, func() bool {
		var job batchv1.Job
		jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
		if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
			return false
		}
		for _, cond := range job.Status.Conditions {
			if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	})

	// AgentRun should transition to Failed
	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseFailed, 60*time.Second)
}
