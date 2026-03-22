//go:build e2e_tee

package e2e_tee

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
		ObjectMeta: metav1.ObjectMeta{GenerateName: prefix + "-"},
	}
	if err := k8sClient.Create(ctx, ns); err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(ctx, ns) })
	return ns.Name
}

// TestE2E_TEE_ConfidentialComputingAMDSEVSNP verifies that an AgentRun with
// confidentialComputing.tee=amd-sev-snp gets kata-cc runtimeClassName and
// AKS confidential compute node affinity on the Job.
func TestE2E_TEE_ConfidentialComputingAMDSEVSNP(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-sev")

	run := testutil.NewAgentRun(ns, "tee-sev-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE: agentv1alpha1.TEETypeAMDSEVSNP,
		}),
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

	// Verify Job PodSpec
	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-sev-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	podSpec := job.Spec.Template.Spec

	// RuntimeClassName should be kata-cc (auto-selected for amd-sev-snp)
	if podSpec.RuntimeClassName == nil || *podSpec.RuntimeClassName != "kata-cc" {
		got := "<nil>"
		if podSpec.RuntimeClassName != nil {
			got = *podSpec.RuntimeClassName
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-cc", got)
	}

	// Node affinity should target AKS confidential computing nodes
	if podSpec.Affinity == nil || podSpec.Affinity.NodeAffinity == nil {
		t.Fatal("expected Affinity with NodeAffinity to be set")
	}
	nodeAffinity := podSpec.Affinity.NodeAffinity
	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.Fatal("expected RequiredDuringSchedulingIgnoredDuringExecution")
	}
	terms := nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) != 1 {
		t.Fatalf("expected 1 NodeSelectorTerm, got %d", len(terms))
	}
	exprs := terms[0].MatchExpressions
	if len(exprs) != 1 {
		t.Fatalf("expected 1 MatchExpression, got %d", len(exprs))
	}
	if exprs[0].Key != "kubernetes.azure.com/confidential-computing" {
		t.Errorf("MatchExpression key = %q, want kubernetes.azure.com/confidential-computing", exprs[0].Key)
	}
	if exprs[0].Operator != corev1.NodeSelectorOpIn {
		t.Errorf("MatchExpression operator = %v, want In", exprs[0].Operator)
	}
	if len(exprs[0].Values) != 1 || exprs[0].Values[0] != "true" {
		t.Errorf("MatchExpression values = %v, want [true]", exprs[0].Values)
	}

	t.Log("Job correctly configured with kata-cc runtimeClassName and AKS CC node affinity")
}

// TestE2E_TEE_ConfidentialComputingIntelTDX verifies that an AgentRun with
// confidentialComputing.tee=intel-tdx gets kata-tdx runtimeClassName.
func TestE2E_TEE_ConfidentialComputingIntelTDX(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-tdx")

	run := testutil.NewAgentRun(ns, "tee-tdx-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE: agentv1alpha1.TEETypeIntelTDX,
		}),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-tdx-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "kata-tdx" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-tdx", got)
	}

	t.Log("Job correctly configured with kata-tdx runtimeClassName")
}

// TestE2E_TEE_CCTakesPrecedenceOverRuntimeClassName verifies that
// confidentialComputing.tee overrides a directly-set runtimeClassName.
func TestE2E_TEE_CCTakesPrecedenceOverRuntimeClassName(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-precedence")

	run := testutil.NewAgentRun(ns, "tee-precedence-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithRuntimeClassName("some-other-runtime"),
		testutil.WithConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE: agentv1alpha1.TEETypeAMDSEVSNP,
		}),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-precedence-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "kata-cc" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-cc (CC takes precedence)", got)
	}

	t.Log("CC correctly takes precedence over direct runtimeClassName")
}

// TestE2E_TEE_OverrideRuntimeClassName verifies that
// confidentialComputing.overrideRuntimeClassName overrides the default for the TEE type.
func TestE2E_TEE_OverrideRuntimeClassName(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-override")

	override := "my-custom-kata-cc"
	run := testutil.NewAgentRun(ns, "tee-override-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE:                     agentv1alpha1.TEETypeAMDSEVSNP,
			OverrideRuntimeClassName: &override,
		}),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-override-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "my-custom-kata-cc" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want my-custom-kata-cc", got)
	}

	t.Log("OverrideRuntimeClassName correctly applied")
}

// TestE2E_TEE_DisableNodeAffinity verifies that disableNodeAffinity=true
// skips adding the AKS node affinity while still setting runtimeClassName.
func TestE2E_TEE_DisableNodeAffinity(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-noaffinity")

	run := testutil.NewAgentRun(ns, "tee-noaffinity-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE:                 agentv1alpha1.TEETypeAMDSEVSNP,
			DisableNodeAffinity: true,
		}),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-noaffinity-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	podSpec := job.Spec.Template.Spec

	// RuntimeClassName should still be set
	if podSpec.RuntimeClassName == nil || *podSpec.RuntimeClassName != "kata-cc" {
		got := "<nil>"
		if podSpec.RuntimeClassName != nil {
			got = *podSpec.RuntimeClassName
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-cc", got)
	}

	// Node affinity should NOT be set
	if podSpec.Affinity != nil {
		t.Errorf("expected no Affinity when disableNodeAffinity=true, got %+v", podSpec.Affinity)
	}

	t.Log("DisableNodeAffinity correctly skips node affinity while keeping runtimeClassName")
}

// TestE2E_TEE_DefaultCCFromHarness verifies that ConfidentialComputing defaults
// are inherited from an AgentHarness when not set on the AgentRun.
func TestE2E_TEE_DefaultCCFromHarness(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-harness")

	harness := testutil.NewAgentHarness(ns, "tee-harness",
		testutil.WithDefaultConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE: agentv1alpha1.TEETypeAMDSEVSNP,
		}),
	)
	if err := k8sClient.Create(ctx, harness); err != nil {
		t.Fatalf("failed to create AgentHarness: %v", err)
	}

	// AgentRun without explicit CC — should inherit from harness
	run := testutil.NewAgentRun(ns, "tee-harness-run",
		testutil.WithHarnessRef("tee-harness"),
		testutil.WithImage("busybox:1.37"),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-harness-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	podSpec := job.Spec.Template.Spec

	// RuntimeClassName should be kata-cc (inherited from harness CC)
	if podSpec.RuntimeClassName == nil || *podSpec.RuntimeClassName != "kata-cc" {
		got := "<nil>"
		if podSpec.RuntimeClassName != nil {
			got = *podSpec.RuntimeClassName
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-cc (inherited from harness CC)", got)
	}

	// Node affinity should be set
	if podSpec.Affinity == nil || podSpec.Affinity.NodeAffinity == nil {
		t.Fatal("expected NodeAffinity inherited from harness CC")
	}

	t.Log("CC defaults correctly inherited from AgentHarness")
}

// TestE2E_TEE_WorkspaceJobWithCC verifies that workspace-based AgentRun Jobs
// get CC runtimeClassName and node affinity.
func TestE2E_TEE_WorkspaceJobWithCC(t *testing.T) {
	ns := createNamespace(t, "e2e-tee-ws")

	// Create workspace
	ws := testutil.NewAgentWorkspace(ns, "tee-ws",
		testutil.WithWorkspaceStorageSize("1Gi"),
	)
	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	// Wait for init job
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var w agentv1alpha1.AgentWorkspace
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), &w); err != nil {
			return false
		}
		return w.Status.InitJobName != ""
	})

	// Manually complete the workspace init Job
	var initJob batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-ws-init", Namespace: ns}, &initJob); err != nil {
		t.Fatalf("init Job not created: %v", err)
	}
	now := metav1.Now()
	initJob.Status.Active = 0
	if initJob.Status.StartTime == nil {
		initJob.Status.StartTime = &now
	}
	initJob.Status.CompletionTime = &now
	initJob.Status.Succeeded = 1
	initJob.Status.Conditions = append(initJob.Status.Conditions,
		batchv1.JobCondition{Type: batchv1.JobSuccessCriteriaMet, Status: corev1.ConditionTrue},
		batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
	)
	if err := k8sClient.Status().Update(ctx, &initJob); err != nil {
		t.Fatalf("failed to update init Job status: %v", err)
	}

	testutil.WaitForWorkspacePhase(t, ctx, k8sClient, client.ObjectKeyFromObject(ws), agentv1alpha1.AgentWorkspacePhaseReady, 60*time.Second)

	// Create AgentRun with workspace ref and CC
	run := testutil.NewAgentRun(ns, "tee-ws-run",
		testutil.WithWorkspaceRef("tee-ws"),
		testutil.WithConfidentialComputing(&agentv1alpha1.ConfidentialComputingSpec{
			TEE: agentv1alpha1.TEETypeAMDSEVSNP,
		}),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var r agentv1alpha1.AgentRun
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(run), &r); err != nil {
			return false
		}
		return r.Status.JobName != ""
	})

	var wsJob batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "tee-ws-run", Namespace: ns}, &wsJob); err != nil {
		t.Fatalf("failed to get workspace Job: %v", err)
	}

	podSpec := wsJob.Spec.Template.Spec

	// RuntimeClassName
	if podSpec.RuntimeClassName == nil || *podSpec.RuntimeClassName != "kata-cc" {
		got := "<nil>"
		if podSpec.RuntimeClassName != nil {
			got = *podSpec.RuntimeClassName
		}
		t.Fatalf("workspace Job RuntimeClassName = %s, want kata-cc", got)
	}

	// Node affinity
	if podSpec.Affinity == nil || podSpec.Affinity.NodeAffinity == nil {
		t.Fatal("expected NodeAffinity on workspace Job")
	}

	t.Log("Workspace Job correctly configured with CC runtimeClassName and node affinity")
}
