//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

func TestAgentRun_CreatesPVCAndTransitionsToInitializing(t *testing.T) {
	ns := createNamespace(t, "pvc-init")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseInitializing, 30*time.Second)

	var pvc corev1.PersistentVolumeClaim
	pvcKey := types.NamespacedName{Name: "test-run-workspace", Namespace: ns}
	if err := k8sClient.Get(ctx, pvcKey, &pvc); err != nil {
		t.Fatalf("PVC not created: %v", err)
	}

	if len(pvc.OwnerReferences) == 0 {
		t.Fatal("PVC has no owner references")
	}
	if pvc.OwnerReferences[0].Kind != "AgentRun" {
		t.Errorf("PVC owner kind = %q, want %q", pvc.OwnerReferences[0].Kind, "AgentRun")
	}

	storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if storageReq.Cmp(resource.MustParse("10Gi")) != 0 {
		t.Errorf("PVC storage = %s, want 10Gi", storageReq.String())
	}

	// Verify AgentRun status
	var updated agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if updated.Status.WorkspacePVC != "test-run-workspace" {
		t.Errorf("WorkspacePVC = %q, want %q", updated.Status.WorkspacePVC, "test-run-workspace")
	}
}

func TestAgentRun_CreatesJobAfterPVC(t *testing.T) {
	ns := createNamespace(t, "job-create")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for Job to be created (status.jobName is set)
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("Job not created: %v", err)
	}

	if *job.Spec.BackoffLimit != 0 {
		t.Errorf("BackoffLimit = %d, want 0", *job.Spec.BackoffLimit)
	}
	if job.Spec.Template.Spec.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("RestartPolicy = %q, want Never", job.Spec.Template.Spec.RestartPolicy)
	}

	// Verify owner reference
	found := false
	for _, ref := range job.OwnerReferences {
		if ref.Kind == "AgentRun" && ref.Name == "test-run" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Job missing AgentRun owner reference")
	}
}

func TestAgentRun_PhaseRunningWhenJobActive(t *testing.T) {
	ns := createNamespace(t, "phase-running")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for Job creation
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Simulate Job becoming active
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	job.Status.Active = 1
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseRunning, 30*time.Second)

	var updated agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if updated.Status.StartTime == nil {
		t.Error("StartTime not set when Running")
	}
}

func TestAgentRun_PhaseSucceededOnJobComplete(t *testing.T) {
	ns := createNamespace(t, "phase-succeeded")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Simulate Job completion (K8s 1.35+ requires startTime, completionTime, and SuccessCriteriaMet)
	now := metav1.Now()
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
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

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseSucceeded, 30*time.Second)

	var updated agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if updated.Status.CompletionTime == nil {
		t.Error("CompletionTime not set on Succeeded")
	}
}

func TestAgentRun_PhaseFailedOnJobFailure(t *testing.T) {
	ns := createNamespace(t, "phase-failed")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Simulate Job failure (K8s 1.35+ requires startTime and FailureTarget condition)
	now := metav1.Now()
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
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
			Message: "BackoffLimitExceeded",
		},
	)
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseFailed, 30*time.Second)

	var updated agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if updated.Status.CompletionTime == nil {
		t.Error("CompletionTime not set on Failed")
	}

	// Verify the failure message is propagated
	found := false
	for _, cond := range updated.Status.Conditions {
		if cond.Type == string(agentv1alpha1.AgentRunPhaseFailed) && cond.Message == "BackoffLimitExceeded" {
			found = true
			break
		}
	}
	if !found {
		t.Error("failure message not propagated to AgentRun conditions")
	}
}

func TestAgentRun_PhaseTimedOut(t *testing.T) {
	ns := createNamespace(t, "phase-timeout")
	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithTimeout(metav1.Duration{Duration: 1 * time.Second}),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Simulate active Job so reconciler enters Running + timeout check
	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	job.Status.Active = 1
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	// Wait for Running first (sets StartTime)
	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseRunning, 30*time.Second)

	// The timeout is 1s; the reconciler requeues every 10s, so we need to wait for the next reconcile.
	// Force a re-reconcile by re-patching the Job status.
	time.Sleep(2 * time.Second)
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	job.Status.Active = 1
	if err := k8sClient.Status().Update(ctx, &job); err != nil {
		t.Fatalf("failed to update Job status: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseTimedOut, 30*time.Second)
}

func TestAgentRun_FailsOnMissingProvider(t *testing.T) {
	ns := createNamespace(t, "missing-provider")
	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithProviderRef("nonexistent-provider"),
	)

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseFailed, 30*time.Second)

	var updated agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}

	found := false
	for _, cond := range updated.Status.Conditions {
		if cond.Type == string(agentv1alpha1.AgentRunPhaseFailed) {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Failed condition on AgentRun")
	}
}

func TestAgentRun_SkipsTerminalPhases(t *testing.T) {
	ns := createNamespace(t, "skip-terminal")
	run := testutil.NewAgentRun(ns, "test-run")

	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for initial reconciliation to create PVC/Job
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	// Manually set phase to Succeeded (retry on conflict since reconciler may race)
	testutil.WaitForCondition(t, ctx, 10*time.Second, func() bool {
		var current agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
			return false
		}
		current.Status.Phase = agentv1alpha1.AgentRunPhaseSucceeded
		return k8sClient.Status().Update(ctx, &current) == nil
	})

	// Trigger a re-reconcile by annotating the object (retry on conflict)
	testutil.WaitForCondition(t, ctx, 10*time.Second, func() bool {
		var current agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
			return false
		}
		if current.Annotations == nil {
			current.Annotations = map[string]string{}
		}
		current.Annotations["test/trigger"] = "reconcile"
		return k8sClient.Update(ctx, &current) == nil
	})

	// Give reconciler time to process
	time.Sleep(2 * time.Second)

	// Verify phase is still Succeeded (reconciler did not change it)
	var current agentv1alpha1.AgentRun
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &current); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if current.Status.Phase != agentv1alpha1.AgentRunPhaseSucceeded {
		t.Errorf("phase = %q, want Succeeded (reconciler should skip terminal phases)", current.Status.Phase)
	}
}

func TestAgentRun_AppliesAgentConfigDefaults(t *testing.T) {
	ns := createNamespace(t, "config-defaults")

	config := testutil.NewAgentConfig(ns, "agent-config",
		testutil.WithStorageSize("20Gi"),
		testutil.WithStorageClass("fast"),
	)
	if err := k8sClient.Create(ctx, config); err != nil {
		t.Fatalf("failed to create AgentConfig: %v", err)
	}

	run := testutil.NewAgentRun(ns, "test-run")
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, client.ObjectKeyFromObject(run), agentv1alpha1.AgentRunPhaseInitializing, 30*time.Second)

	var pvc corev1.PersistentVolumeClaim
	pvcKey := types.NamespacedName{Name: "test-run-workspace", Namespace: ns}
	if err := k8sClient.Get(ctx, pvcKey, &pvc); err != nil {
		t.Fatalf("PVC not created: %v", err)
	}

	storageReq := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if storageReq.Cmp(resource.MustParse("20Gi")) != 0 {
		t.Errorf("PVC storage = %s, want 20Gi", storageReq.String())
	}

	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "fast" {
		sc := "<nil>"
		if pvc.Spec.StorageClassName != nil {
			sc = *pvc.Spec.StorageClassName
		}
		t.Errorf("PVC storageClassName = %s, want fast", sc)
	}
}

func TestAgentRun_ProviderEnvVarsInjectedIntoJob(t *testing.T) {
	ns := createNamespace(t, "provider-env")

	secret := testutil.NewSecret(ns, "provider-secret", map[string][]byte{
		"api-key": []byte("test-key-123"),
	})
	if err := k8sClient.Create(ctx, secret); err != nil {
		t.Fatalf("failed to create Secret: %v", err)
	}

	provider := testutil.NewAgentProvider(ns, "test-provider",
		testutil.WithAgentTypes(agentv1alpha1.AgentTypeClaude),
		testutil.WithAuthTokenSecretRef("provider-secret", "api-key"),
		testutil.WithEnvironment(map[string]string{
			"ANTHROPIC_BASE_URL": "http://proxy:8080",
			"CUSTOM_VAR":         "custom-value",
		}),
	)
	if err := k8sClient.Create(ctx, provider); err != nil {
		t.Fatalf("failed to create AgentProvider: %v", err)
	}

	run := testutil.NewAgentRun(ns, "test-run",
		testutil.WithProviderRef("test-provider"),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	// Wait for Job creation
	testutil.WaitForCondition(t, ctx, 30*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	jobKey := types.NamespacedName{Name: "test-run", Namespace: ns}
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("Job not created: %v", err)
	}

	envMap := map[string]string{}
	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		envMap[env.Name] = env.Value
	}

	expected := map[string]string{
		"ANTHROPIC_BASE_URL":   "http://proxy:8080",
		"CUSTOM_VAR":           "custom-value",
		"ANTHROPIC_AUTH_TOKEN": "test-key-123",
	}

	for k, v := range expected {
		if envMap[k] != v {
			t.Errorf("env %s = %q, want %q", k, envMap[k], v)
		}
	}
}

func TestAgentRun_UniqueNamesPerNamespace(t *testing.T) {
	ns1 := createNamespace(t, "unique-ns1")
	ns2 := createNamespace(t, "unique-ns2")

	for i, ns := range []string{ns1, ns2} {
		run := testutil.NewAgentRun(ns, fmt.Sprintf("test-run-%d", i))
		if err := k8sClient.Create(ctx, run); err != nil {
			t.Fatalf("failed to create AgentRun in %s: %v", ns, err)
		}
	}

	// Verify both runs reach Initializing independently
	for i, ns := range []string{ns1, ns2} {
		key := types.NamespacedName{Name: fmt.Sprintf("test-run-%d", i), Namespace: ns}
		testutil.WaitForAgentRunPhase(t, ctx, k8sClient, key, agentv1alpha1.AgentRunPhaseInitializing, 30*time.Second)
	}
}
