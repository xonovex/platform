//go:build e2e_coco

package e2e_coco

import (
	"os/exec"
	"path/filepath"
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

const e2eAgentImage = "e2e-agent:e2e"

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

func gitWorkspaceRoot(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to find workspace root: %v", err)
	}
	return filepath.Clean(string(out[:len(out)-1]))
}

func buildAndLoadE2EAgentImage(t *testing.T) {
	t.Helper()
	workspaceRoot := gitWorkspaceRoot(t)
	dockerfile := filepath.Join(workspaceRoot, "packages", "agent", "agent-operator-go", "test", "testdata", "Dockerfile.e2e-agent")

	buildCmd := exec.Command("docker", "build", "-f", dockerfile, "-t", e2eAgentImage, ".")
	buildCmd.Dir = workspaceRoot
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker build failed: %v\n%s", err, out)
	}

	var loadErr error
	for range 5 {
		loadErr = runCmd("kind", "load", "docker-image", e2eAgentImage, "--name", clusterName)
		if loadErr == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	if loadErr != nil {
		t.Fatalf("kind load failed after retries: %v", loadErr)
	}
}

// TestE2E_CoCo_AgentRunWithKataCC verifies that an AgentRun with
// runtimeClassName=kata-cc gets the correct runtimeClassName on the Job.
func TestE2E_CoCo_AgentRunWithKataCC(t *testing.T) {
	ns := createNamespace(t, "e2e-coco-cc")

	run := testutil.NewAgentRun(ns, "coco-cc-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithRuntimeClassName("kata-cc"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "coco-cc-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "kata-cc" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-cc", got)
	}

	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "coco-cc-run",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})

	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
		"app.kubernetes.io/instance": "coco-cc-run",
	}); err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}
	pod := podList.Items[0]
	if pod.Spec.RuntimeClassName == nil || *pod.Spec.RuntimeClassName != "kata-cc" {
		got := "<nil>"
		if pod.Spec.RuntimeClassName != nil {
			got = *pod.Spec.RuntimeClassName
		}
		t.Errorf("Pod RuntimeClassName = %s, want kata-cc", got)
	}
	t.Logf("Pod %s scheduled on node %s with runtimeClassName=kata-cc", pod.Name, pod.Spec.NodeName)
}

// TestE2E_CoCo_AgentRunWithKataTDX verifies that runtimeClassName=kata-tdx
// is correctly propagated to the Job.
func TestE2E_CoCo_AgentRunWithKataTDX(t *testing.T) {
	ns := createNamespace(t, "e2e-coco-tdx")

	run := testutil.NewAgentRun(ns, "coco-tdx-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithRuntimeClassName("kata-tdx"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "coco-tdx-run", Namespace: ns}, &job); err != nil {
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
	t.Log("Job correctly configured with runtimeClassName=kata-tdx")
}

// TestE2E_CoCo_DefaultRuntimeClassNameFromHarness verifies that kata-cc
// can be inherited from an AgentHarness default.
func TestE2E_CoCo_DefaultRuntimeClassNameFromHarness(t *testing.T) {
	ns := createNamespace(t, "e2e-coco-harness")

	harness := testutil.NewAgentHarness(ns, "coco-harness",
		testutil.WithDefaultRuntimeClassName("kata-cc"),
	)
	if err := k8sClient.Create(ctx, harness); err != nil {
		t.Fatalf("failed to create AgentHarness: %v", err)
	}

	run := testutil.NewAgentRun(ns, "coco-harness-run",
		testutil.WithHarnessRef("coco-harness"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "coco-harness-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "kata-cc" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Errorf("Job RuntimeClassName = %s, want kata-cc (inherited from AgentHarness)", got)
	}

	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "coco-harness-run",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})
	t.Log("Pod scheduled with inherited runtimeClassName=kata-cc")
}

// TestE2E_CoCo_FullCycleWithGitClone proves the entire agent pipeline
// (Secret + AgentProvider + git clone + fake claude) works end-to-end
// with the kata-cc runtimeClassName.
func TestE2E_CoCo_FullCycleWithGitClone(t *testing.T) {
	buildAndLoadE2EAgentImage(t)

	ns := createNamespace(t, "e2e-coco-fullcycle")

	secret := testutil.NewSecret(ns, "provider-token", map[string][]byte{
		"api-key": []byte("fake-token-for-e2e"),
	})
	if err := k8sClient.Create(ctx, secret); err != nil {
		t.Fatalf("failed to create Secret: %v", err)
	}

	provider := testutil.NewAgentProvider(ns, "test-provider",
		testutil.WithAuthTokenSecretRef("provider-token", "api-key"),
		testutil.WithEnvironment(map[string]string{
			"ANTHROPIC_API_KEY": "fake-key",
			"TEST_ENV_VAR":      "e2e-value",
		}),
	)
	if err := k8sClient.Create(ctx, provider); err != nil {
		t.Fatalf("failed to create AgentProvider: %v", err)
	}

	run := testutil.NewAgentRun(ns, "coco-fullcycle",
		testutil.WithHarness(&agentv1alpha1.AgentSpec{Type: agentv1alpha1.AgentTypeClaude}),
		testutil.WithPrompt("echo test-prompt"),
		testutil.WithImage(e2eAgentImage),
		testutil.WithWorkspace(&agentv1alpha1.WorkspaceSpec{
			Repository:  agentv1alpha1.RepositorySpec{URL: "https://github.com/octocat/Hello-World.git"},
			StorageSize: "1Gi",
		}),
		testutil.WithProviderRef("test-provider"),
		testutil.WithRuntimeClassName("kata-cc"),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	runKey := client.ObjectKeyFromObject(run)

	// Wait for WorkspacePVC
	var runStatus agentv1alpha1.AgentRun
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		if err := k8sClient.Get(ctx, runKey, &runStatus); err != nil {
			return false
		}
		return runStatus.Status.WorkspacePVC != ""
	})
	t.Logf("WorkspacePVC: %s (phase: %s)", runStatus.Status.WorkspacePVC, runStatus.Status.Phase)

	// Verify PVC exists
	pvcKey := types.NamespacedName{Name: runStatus.Status.WorkspacePVC, Namespace: ns}
	var pvc corev1.PersistentVolumeClaim
	if err := k8sClient.Get(ctx, pvcKey, &pvc); err != nil {
		t.Fatalf("PVC %q not found: %v", pvcKey.Name, err)
	}

	// Wait for Job creation
	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		if err := k8sClient.Get(ctx, runKey, &runStatus); err != nil {
			return false
		}
		return runStatus.Status.JobName != ""
	})
	t.Logf("JobName: %s (phase: %s)", runStatus.Status.JobName, runStatus.Status.Phase)

	// Wait for Pod scheduling
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "coco-fullcycle",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})

	// Wait for Succeeded
	testutil.WaitForAgentRunPhase(t, ctx, k8sClient, runKey, agentv1alpha1.AgentRunPhaseSucceeded, 180*time.Second)

	if err := k8sClient.Get(ctx, runKey, &runStatus); err != nil {
		t.Fatalf("failed to get AgentRun: %v", err)
	}
	if runStatus.Status.StartTime == nil {
		t.Fatal("expected StartTime to be set")
	}
	if runStatus.Status.CompletionTime == nil {
		t.Fatal("expected CompletionTime to be set")
	}

	// Verify Job spec
	jobKey := types.NamespacedName{Name: runStatus.Status.JobName, Namespace: ns}
	var job batchv1.Job
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	if got := job.Labels["agent.xonovex.com/agent-type"]; got != "claude" {
		t.Errorf("job label agent-type = %q, want %q", got, "claude")
	}

	podSpec := job.Spec.Template.Spec

	rc := podSpec.RuntimeClassName
	if rc == nil || *rc != "kata-cc" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata-cc", got)
	}

	if len(podSpec.InitContainers) == 0 {
		t.Fatal("expected at least one init container")
	}
	if podSpec.InitContainers[0].Image != e2eAgentImage {
		t.Errorf("init container image = %q, want %q", podSpec.InitContainers[0].Image, e2eAgentImage)
	}

	if len(podSpec.Containers) == 0 {
		t.Fatal("expected at least one main container")
	}
	mainC := podSpec.Containers[0]
	if mainC.Image != e2eAgentImage {
		t.Errorf("main container image = %q, want %q", mainC.Image, e2eAgentImage)
	}
	if len(mainC.Command) == 0 || mainC.Command[0] != "claude" {
		t.Errorf("main container command = %v, want [claude]", mainC.Command)
	}

	envMap := make(map[string]string)
	for _, e := range mainC.Env {
		envMap[e.Name] = e.Value
	}
	if v, ok := envMap["TEST_ENV_VAR"]; !ok || v != "e2e-value" {
		t.Errorf("expected TEST_ENV_VAR=e2e-value in main container env, got %q (present=%v)", v, ok)
	}

	foundVolume := false
	for _, v := range podSpec.Volumes {
		if v.PersistentVolumeClaim != nil && v.PersistentVolumeClaim.ClaimName == runStatus.Status.WorkspacePVC {
			foundVolume = true
			break
		}
	}
	if !foundVolume {
		t.Errorf("expected PVC %q to be mounted as a volume", runStatus.Status.WorkspacePVC)
	}

	t.Log("Full workflow completed with kata-cc runtimeClassName")
}

// TestE2E_CoCo_WorkspaceJobWithRuntimeClassName verifies that workspace-based
// AgentRun Jobs get runtimeClassName but workspace init Jobs do not.
func TestE2E_CoCo_WorkspaceJobWithRuntimeClassName(t *testing.T) {
	ns := createNamespace(t, "e2e-coco-ws")

	ws := testutil.NewAgentWorkspace(ns, "coco-ws",
		testutil.WithWorkspaceStorageSize("1Gi"),
	)
	if err := k8sClient.Create(ctx, ws); err != nil {
		t.Fatalf("failed to create AgentWorkspace: %v", err)
	}

	testutil.WaitForCondition(t, ctx, 60*time.Second, func() bool {
		var w agentv1alpha1.AgentWorkspace
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), &w); err != nil {
			return false
		}
		return w.Status.InitJobName != ""
	})

	// Verify workspace init Job does NOT have runtimeClassName
	var initJob batchv1.Job
	initJobKey := types.NamespacedName{Name: "coco-ws-init", Namespace: ns}
	if err := k8sClient.Get(ctx, initJobKey, &initJob); err != nil {
		t.Fatalf("init Job not created: %v", err)
	}
	if initJob.Spec.Template.Spec.RuntimeClassName != nil {
		t.Errorf("workspace init Job RuntimeClassName = %s, want <nil>", *initJob.Spec.Template.Spec.RuntimeClassName)
	}

	// Manually complete the workspace init Job
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

	// Create AgentRun with workspace ref and kata-cc
	run := testutil.NewAgentRun(ns, "coco-ws-run",
		testutil.WithWorkspaceRef("coco-ws"),
		testutil.WithRuntimeClassName("kata-cc"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "coco-ws-run", Namespace: ns}, &wsJob); err != nil {
		t.Fatalf("failed to get workspace Job: %v", err)
	}
	wsRC := wsJob.Spec.Template.Spec.RuntimeClassName
	if wsRC == nil || *wsRC != "kata-cc" {
		got := "<nil>"
		if wsRC != nil {
			got = *wsRC
		}
		t.Fatalf("workspace Job RuntimeClassName = %s, want kata-cc", got)
	}
	t.Log("workspace init Job: no runtimeClassName (correct), agent Job: runtimeClassName=kata-cc")
}
