//go:build e2e

package e2e

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestE2E_JobGCOnAgentRunDelete(t *testing.T) {
	ns := createNamespace(t, "e2e-job-gc")

	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithImage("busybox:1.37"),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for Job creation
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var job batchv1.Job
		return k8sClient.Get(ctx, jobKey, &job) == nil
	})

	// Delete the AgentRun
	var current agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if err := k8sClient.Delete(ctx, &current); err != nil {
		t.Fatalf("failed to delete AgentRun: %v", err)
	}

	// Wait for Job to be garbage collected
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var job batchv1.Job
		err := k8sClient.Get(ctx, jobKey, &job)
		return errors.IsNotFound(err)
	})
}

func TestE2E_PVCGCOnAgentRunDelete(t *testing.T) {
	ns := createNamespace(t, "e2e-pvc-gc")

	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithImage("busybox:1.37"),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for PVC creation
	pvcKey := types.NamespacedName{Name: "test-run-workspace", Namespace: ns}
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var pvc corev1.PersistentVolumeClaim
		return k8sClient.Get(ctx, pvcKey, &pvc) == nil
	})

	// Delete the AgentRun
	var current agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if err := k8sClient.Delete(ctx, &current); err != nil {
		t.Fatalf("failed to delete AgentRun: %v", err)
	}

	// Wait for PVC to be garbage collected
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var pvc corev1.PersistentVolumeClaim
		err := k8sClient.Get(ctx, pvcKey, &pvc)
		return errors.IsNotFound(err)
	})
}
