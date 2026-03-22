//go:build e2e_gvisor

package e2e_gvisor

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestE2E_GVisor_SandboxIsolation verifies that gVisor actually intercepts
// syscalls by checking that dmesg output shows the gVisor banner.
func TestE2E_GVisor_SandboxIsolation(t *testing.T) {
	ns := createNamespace(t, "e2e-gvisor-sandbox")

	runtimeClass := "gvisor"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gvisor-sandbox-check",
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			RuntimeClassName: &runtimeClass,
			RestartPolicy:    corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "check",
					Image:   "busybox:1.37",
					Command: []string{"sleep", "300"},
				},
			},
		},
	}
	if err := k8sClient.Create(ctx, pod); err != nil {
		t.Fatalf("failed to create Pod: %v", err)
	}
	t.Cleanup(func() { _ = k8sClient.Delete(ctx, pod) })

	var lastPod corev1.Pod
	deadline := time.After(60 * time.Second)
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	podRunning := false
	for !podRunning {
		select {
		case <-deadline:
			out, _ := exec.Command("kubectl", "--context", "kind-"+clusterName,
				"-n", ns, "describe", "pod", "gvisor-sandbox-check").CombinedOutput()
			t.Fatalf("Pod never reached Running (phase=%s). Describe:\n%s", lastPod.Status.Phase, string(out))
		case <-tick.C:
			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &lastPod); err == nil {
				podRunning = lastPod.Status.Phase == corev1.PodRunning
			}
		}
	}

	// gVisor intercepts dmesg — output starts with "Starting gVisor..."
	var dmesgBuf bytes.Buffer
	dmesgCmd := exec.Command("kubectl", "--context", "kind-"+clusterName,
		"-n", ns, "exec", "gvisor-sandbox-check", "--", "dmesg")
	dmesgCmd.Stdout = &dmesgBuf
	dmesgCmd.Stderr = &dmesgBuf
	if err := dmesgCmd.Run(); err != nil {
		t.Fatalf("failed to exec dmesg in gvisor pod: %v\noutput: %s", err, dmesgBuf.String())
	}
	dmesgOut := dmesgBuf.String()
	t.Logf("dmesg output (first 200 chars): %.200s", dmesgOut)

	if !strings.Contains(dmesgOut, "gVisor") {
		t.Errorf("dmesg does not contain 'gVisor' — container may not be running in gVisor sandbox")
	} else {
		t.Log("Sandbox isolation confirmed: gVisor intercepting syscalls")
	}
}

func TestE2E_GVisor_AgentRunWithRuntimeClassName(t *testing.T) {
	ns := createNamespace(t, "e2e-gvisor")

	run := testutil.NewAgentRun(ns, "gvisor-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithRuntimeClassName("gvisor"),
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

	// Verify Job PodSpec has runtimeClassName set
	var job batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "gvisor-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "gvisor" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want gvisor", got)
	}

	// Wait for Pod to be scheduled on a node
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "gvisor-run",
		}); err != nil {
			return false
		}
		if len(podList.Items) == 0 {
			return false
		}
		return podList.Items[0].Spec.NodeName != ""
	})

	// Verify Pod is using gvisor runtime
	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
		"app.kubernetes.io/instance": "gvisor-run",
	}); err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}
	pod := podList.Items[0]
	if pod.Spec.RuntimeClassName == nil || *pod.Spec.RuntimeClassName != "gvisor" {
		got := "<nil>"
		if pod.Spec.RuntimeClassName != nil {
			got = *pod.Spec.RuntimeClassName
		}
		t.Errorf("Pod RuntimeClassName = %s, want gvisor", got)
	}
	t.Logf("Pod %s scheduled on node %s with runtimeClassName=gvisor", pod.Name, pod.Spec.NodeName)
}

func TestE2E_GVisor_DefaultRuntimeClassNameFromHarness(t *testing.T) {
	ns := createNamespace(t, "e2e-gvisor-default")

	harness := testutil.NewAgentHarness(ns, "agent-harness",
		testutil.WithDefaultRuntimeClassName("gvisor"),
	)
	if err := k8sClient.Create(ctx, harness); err != nil {
		t.Fatalf("failed to create AgentHarness: %v", err)
	}

	// AgentRun without explicit runtimeClassName — should inherit from harness
	run := testutil.NewAgentRun(ns, "gvisor-default-run",
		testutil.WithHarnessRef("agent-harness"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "gvisor-default-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "gvisor" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Errorf("Job RuntimeClassName = %s, want gvisor (inherited from AgentHarness)", got)
	}

	// Wait for Pod to be scheduled — proves the runtime is actually available
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "gvisor-default-run",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})
	t.Log("Pod scheduled with inherited runtimeClassName=gvisor")
}

// TestE2E_GVisor_FullCycleWithGitClone proves the entire agent pipeline
// (Secret + AgentProvider + AgentHarness + git clone init container + fake claude)
// works end-to-end inside a gVisor sandbox.
func TestE2E_GVisor_FullCycleWithGitClone(t *testing.T) {
	buildAndLoadE2EAgentImage(t)

	ns := createNamespace(t, "e2e-gvisor-fullcycle")

	// Create Secret with a fake auth token
	secret := testutil.NewSecret(ns, "provider-token", map[string][]byte{
		"api-key": []byte("fake-token-for-e2e"),
	})
	if err := k8sClient.Create(ctx, secret); err != nil {
		t.Fatalf("failed to create Secret: %v", err)
	}

	// Create AgentProvider referencing the secret
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

	// Create AgentRun exercising the full pipeline inside gVisor
	run := testutil.NewAgentRun(ns, "gvisor-fullcycle",
		testutil.WithHarness(&agentv1alpha1.AgentSpec{Type: agentv1alpha1.AgentTypeClaude}),
		testutil.WithPrompt("echo test-prompt"),
		testutil.WithImage(e2eAgentImage),
		testutil.WithWorkspace(&agentv1alpha1.WorkspaceSpec{
			Repository:  agentv1alpha1.RepositorySpec{URL: "https://github.com/octocat/Hello-World.git"},
			StorageSize: "1Gi",
		}),
		testutil.WithProviderRef("test-provider"),
		testutil.WithRuntimeClassName("gvisor"),
	)
	if err := k8sClient.Create(ctx, run); err != nil {
		t.Fatalf("failed to create AgentRun: %v", err)
	}

	runKey := client.ObjectKeyFromObject(run)

	// Wait for WorkspacePVC to be populated (proves Initializing step happened)
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

	// Wait for Pod to be scheduled on a node
	testutil.WaitForCondition(t, ctx, 120*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "gvisor-fullcycle",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})

	// Wait for Succeeded — proves init container (git clone) + main container
	// (fake claude) both completed inside gVisor
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
	t.Logf("StartTime: %s, CompletionTime: %s", runStatus.Status.StartTime, runStatus.Status.CompletionTime)

	// Verify Job spec correctness
	jobKey := types.NamespacedName{Name: runStatus.Status.JobName, Namespace: ns}
	var job batchv1.Job
	if err := k8sClient.Get(ctx, jobKey, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}

	// Labels
	if got := job.Labels["agent.xonovex.com/agent-type"]; got != "claude" {
		t.Errorf("job label agent-type = %q, want %q", got, "claude")
	}

	podSpec := job.Spec.Template.Spec

	// RuntimeClassName
	rc := podSpec.RuntimeClassName
	if rc == nil || *rc != "gvisor" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want gvisor", got)
	}

	// Init container
	if len(podSpec.InitContainers) == 0 {
		t.Fatal("expected at least one init container")
	}
	initC := podSpec.InitContainers[0]
	if initC.Image != e2eAgentImage {
		t.Errorf("init container image = %q, want %q", initC.Image, e2eAgentImage)
	}

	// Main container
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

	// Prompt args
	wantArgs := []string{"--permission-mode", "bypassPermissions", "--print", "--prompt", "echo test-prompt"}
	if len(mainC.Args) != len(wantArgs) {
		t.Errorf("main container args = %v, want %v", mainC.Args, wantArgs)
	} else {
		for i, want := range wantArgs {
			if mainC.Args[i] != want {
				t.Errorf("main container args[%d] = %q, want %q", i, mainC.Args[i], want)
			}
		}
	}

	// Provider env vars injected into main container
	envMap := make(map[string]string)
	for _, e := range mainC.Env {
		envMap[e.Name] = e.Value
	}
	if v, ok := envMap["TEST_ENV_VAR"]; !ok || v != "e2e-value" {
		t.Errorf("expected TEST_ENV_VAR=e2e-value in main container env, got %q (present=%v)", v, ok)
	}

	// PVC volume mounted
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

	t.Log("Full workflow completed inside gVisor sandbox")
}

// TestE2E_GVisor_WorkspaceJobWithRuntimeClassName verifies that workspace-based
// AgentRun Jobs get runtimeClassName but workspace init Jobs do not.
func TestE2E_GVisor_WorkspaceJobWithRuntimeClassName(t *testing.T) {
	ns := createNamespace(t, "e2e-gvisor-ws")

	// Create workspace with a repository
	ws := testutil.NewAgentWorkspace(ns, "gvisor-ws",
		testutil.WithWorkspaceStorageSize("1Gi"),
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

	// Verify workspace init Job does NOT have runtimeClassName
	var initJob batchv1.Job
	initJobKey := types.NamespacedName{Name: "gvisor-ws-init", Namespace: ns}
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

	// Create AgentRun with workspace ref and runtimeClassName
	run := testutil.NewAgentRun(ns, "gvisor-ws-run",
		testutil.WithWorkspaceRef("gvisor-ws"),
		testutil.WithRuntimeClassName("gvisor"),
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

	// Verify the workspace-based Job has runtimeClassName=gvisor
	var wsJob batchv1.Job
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "gvisor-ws-run", Namespace: ns}, &wsJob); err != nil {
		t.Fatalf("failed to get workspace Job: %v", err)
	}
	wsRC := wsJob.Spec.Template.Spec.RuntimeClassName
	if wsRC == nil || *wsRC != "gvisor" {
		got := "<nil>"
		if wsRC != nil {
			got = *wsRC
		}
		t.Fatalf("workspace Job RuntimeClassName = %s, want gvisor", got)
	}
	t.Logf("workspace init Job: no runtimeClassName (correct), agent Job: runtimeClassName=gvisor")
}
