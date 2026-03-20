//go:build e2e_kata

package e2e_kata

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

// TestE2E_Kata_VMIsolation verifies that Kata actually runs containers inside
// a hypervisor VM by comparing the guest kernel to the host kernel.
func TestE2E_Kata_VMIsolation(t *testing.T) {
	ns := createNamespace(t, "e2e-kata-vm")

	// Get the host kernel version from the kind node
	var hostKernelBuf bytes.Buffer
	hostCmd := exec.Command("docker", "exec", clusterName+"-control-plane", "uname", "-r")
	hostCmd.Stdout = &hostKernelBuf
	if err := hostCmd.Run(); err != nil {
		t.Fatalf("failed to get host kernel: %v", err)
	}
	hostKernel := strings.TrimSpace(hostKernelBuf.String())
	t.Logf("Host kernel: %s", hostKernel)

	// Create a pod with kata runtime that sleeps so we can exec into it
	runtimeClass := "kata"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kata-vm-check",
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

	// Wait for pod to be Running
	var lastPod corev1.Pod
	deadline := time.After(180 * time.Second)
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	podRunning := false
	for !podRunning {
		select {
		case <-deadline:
			out, _ := exec.Command("kubectl", "--context", "kind-"+clusterName,
				"-n", ns, "describe", "pod", "kata-vm-check").CombinedOutput()
			describeOut := string(out)
			// Kata VMs need privileged containers for /dev/kvm access inside kind nodes.
			// When running in unprivileged kind, the VM fails to connect via vsock.
			if strings.Contains(describeOut, "vsock") || strings.Contains(describeOut, "QEMU") {
				t.Skipf("Kata VM cannot start inside kind (likely unprivileged container). Use a real cluster or privileged kind.\nDescribe:\n%s", describeOut)
			}
			t.Fatalf("Pod never reached Running (phase=%s). Describe:\n%s", lastPod.Status.Phase, describeOut)
		case <-tick.C:
			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &lastPod); err == nil {
				podRunning = lastPod.Status.Phase == corev1.PodRunning
			}
		}
	}

	// Exec into the pod and get the guest kernel version
	var guestKernelBuf bytes.Buffer
	execCmd := exec.Command("kubectl", "--context", "kind-"+clusterName,
		"-n", ns, "exec", "kata-vm-check", "--", "uname", "-r")
	execCmd.Stdout = &guestKernelBuf
	execCmd.Stderr = &guestKernelBuf
	if err := execCmd.Run(); err != nil {
		t.Fatalf("failed to exec uname in kata pod: %v\noutput: %s", err, guestKernelBuf.String())
	}
	guestKernel := strings.TrimSpace(guestKernelBuf.String())
	t.Logf("Guest kernel: %s", guestKernel)

	// Kata runs a lightweight guest kernel — it must differ from the host kernel
	if guestKernel == hostKernel {
		t.Errorf("guest kernel (%s) matches host kernel — container is NOT running in a Kata VM", guestKernel)
	} else {
		t.Logf("VM isolation confirmed: host=%s, guest=%s", hostKernel, guestKernel)
	}

	// Additional check: /dev/pmem0 exists inside Kata VMs (rootfs is passed via PMEM)
	var pmemBuf bytes.Buffer
	pmemCmd := exec.Command("kubectl", "--context", "kind-"+clusterName,
		"-n", ns, "exec", "kata-vm-check", "--", "ls", "/dev/pmem0")
	pmemCmd.Stdout = &pmemBuf
	pmemCmd.Stderr = &pmemBuf
	if err := pmemCmd.Run(); err != nil {
		t.Logf("/dev/pmem0 not found (may vary by Kata config): %s", pmemBuf.String())
	} else {
		t.Logf("/dev/pmem0 present — confirms Kata VM")
	}
}

func TestE2E_Kata_AgentRunWithRuntimeClassName(t *testing.T) {
	ns := createNamespace(t, "e2e-kata")

	run := testutil.NewAgentRun(ns, "kata-run",
		testutil.WithImage("busybox:1.37"),
		testutil.WithRuntimeClassName("kata"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "kata-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "kata" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata", got)
	}

	// Wait for Pod to be scheduled on a node
	testutil.WaitForCondition(t, ctx, 180*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "kata-run",
		}); err != nil {
			return false
		}
		if len(podList.Items) == 0 {
			return false
		}
		return podList.Items[0].Spec.NodeName != ""
	})

	// Verify Pod is using kata runtime
	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
		"app.kubernetes.io/instance": "kata-run",
	}); err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}
	pod := podList.Items[0]
	if pod.Spec.RuntimeClassName == nil || *pod.Spec.RuntimeClassName != "kata" {
		got := "<nil>"
		if pod.Spec.RuntimeClassName != nil {
			got = *pod.Spec.RuntimeClassName
		}
		t.Errorf("Pod RuntimeClassName = %s, want kata", got)
	}
	t.Logf("Pod %s scheduled on node %s with runtimeClassName=kata", pod.Name, pod.Spec.NodeName)
}

func TestE2E_Kata_DefaultRuntimeClassNameFromConfig(t *testing.T) {
	ns := createNamespace(t, "e2e-kata-default")

	config := testutil.NewAgentConfig(ns, "agent-config",
		testutil.WithDefaultRuntimeClassName("kata"),
	)
	if err := k8sClient.Create(ctx, config); err != nil {
		t.Fatalf("failed to create AgentConfig: %v", err)
	}

	// AgentRun without explicit runtimeClassName — should inherit from config
	run := testutil.NewAgentRun(ns, "kata-default-run",
		testutil.WithConfigRef("agent-config"),
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
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "kata-default-run", Namespace: ns}, &job); err != nil {
		t.Fatalf("failed to get Job: %v", err)
	}
	rc := job.Spec.Template.Spec.RuntimeClassName
	if rc == nil || *rc != "kata" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Errorf("Job RuntimeClassName = %s, want kata (inherited from AgentConfig)", got)
	}

	// Wait for Pod to be scheduled
	testutil.WaitForCondition(t, ctx, 180*time.Second, func() bool {
		var podList corev1.PodList
		if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
			"app.kubernetes.io/instance": "kata-default-run",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})
	t.Log("Pod scheduled with inherited runtimeClassName=kata")
}

// TestE2E_Kata_FullCycleWithGitClone proves the entire agent pipeline
// (Secret + AgentProvider + AgentConfig + git clone init container + fake claude)
// works end-to-end inside a Kata VM sandbox.
// Skips gracefully if Kata VM cannot start in unprivileged kind.
func TestE2E_Kata_FullCycleWithGitClone(t *testing.T) {
	buildAndLoadE2EAgentImage(t)

	ns := createNamespace(t, "e2e-kata-fullcycle")

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
			"TEST_ENV_VAR":     "e2e-value",
		}),
	)
	if err := k8sClient.Create(ctx, provider); err != nil {
		t.Fatalf("failed to create AgentProvider: %v", err)
	}

	// Create AgentConfig with storage defaults
	agentConfig := testutil.NewAgentConfig(ns, "default",
		testutil.WithStorageSize("1Gi"),
	)
	if err := k8sClient.Create(ctx, agentConfig); err != nil {
		t.Fatalf("failed to create AgentConfig: %v", err)
	}

	// Create AgentRun exercising the full pipeline inside Kata
	run := testutil.NewAgentRun(ns, "kata-fullcycle",
		testutil.WithAgent(agentv1alpha1.AgentTypeClaude),
		testutil.WithConfigRef("default"),
		testutil.WithPrompt("echo test-prompt"),
		testutil.WithImage(e2eAgentImage),
		testutil.WithRepository("https://github.com/octocat/Hello-World.git"),
		testutil.WithProviderRef("test-provider"),
		testutil.WithRuntimeClassName("kata"),
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

	// Wait for Pod to be scheduled — check if Kata VM can start
	podScheduled := false
	deadline := time.After(180 * time.Second)
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for !podScheduled {
		select {
		case <-deadline:
			var podList corev1.PodList
			if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
				"app.kubernetes.io/instance": "kata-fullcycle",
			}); err == nil && len(podList.Items) > 0 {
				pod := podList.Items[0]
				out, _ := exec.Command("kubectl", "--context", "kind-"+clusterName,
					"-n", ns, "describe", "pod", pod.Name).CombinedOutput()
				describeOut := string(out)
				if strings.Contains(describeOut, "vsock") || strings.Contains(describeOut, "QEMU") {
					t.Skipf("Kata VM cannot start inside kind (likely unprivileged container). Use a real cluster or privileged kind.\nDescribe:\n%s", describeOut)
				}
			}
			t.Fatalf("Pod never scheduled within timeout")
		case <-tick.C:
			var podList corev1.PodList
			if err := k8sClient.List(ctx, &podList, client.InNamespace(ns), client.MatchingLabels{
				"app.kubernetes.io/instance": "kata-fullcycle",
			}); err == nil && len(podList.Items) > 0 {
				pod := podList.Items[0]
				if pod.Spec.NodeName != "" {
					podScheduled = true
				}
				for _, cs := range pod.Status.ContainerStatuses {
					if cs.State.Waiting != nil && cs.State.Waiting.Message != "" {
						msg := cs.State.Waiting.Message
						if strings.Contains(msg, "vsock") || strings.Contains(msg, "QEMU") {
							t.Skipf("Kata VM cannot start: %s", msg)
						}
					}
				}
			}
		}
	}

	// Wait for Succeeded — proves init container (git clone) + main container
	// (fake claude) both completed inside Kata
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
	if rc == nil || *rc != "kata" {
		got := "<nil>"
		if rc != nil {
			got = *rc
		}
		t.Fatalf("Job RuntimeClassName = %s, want kata", got)
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

	t.Log("Full workflow completed inside Kata sandbox")
}
