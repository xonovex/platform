//go:build e2e

package e2e

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

const e2eAgentImage = "e2e-agent:e2e"

func TestE2E_FullCycleWithPrompt(t *testing.T) {
	if useExistingCluster {
		t.Skip("full-cycle test requires Kind (skipped with USE_EXISTING_CLUSTER=true)")
	}

	// Build the test image with a fake claude binary
	workspaceRoot := gitWorkspaceRoot(t)
	dockerfile := filepath.Join(workspaceRoot, "packages", "agent", "agent-operator-go", "test", "testdata", "Dockerfile.e2e-agent")

	buildCmd := exec.Command("docker", "build", "-f", dockerfile, "-t", e2eAgentImage, ".")
	buildCmd.Dir = workspaceRoot
	out, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker build failed: %v\n%s", err, out)
	}

	// Retry kind load — containerd may not be ready immediately after cluster creation
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

	ns := createNamespace(t, "e2e-fullcycle")

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

	// Create AgentConfig with storage defaults
	agentConfig := testutil.NewAgentConfig(ns, "default",
		testutil.WithStorageSize("1Gi"),
	)
	if err := k8sClient.Create(ctx, agentConfig); err != nil {
		t.Fatalf("failed to create AgentConfig: %v", err)
	}

	// Create AgentRun exercising the full pipeline
	run := testutil.NewAgentRun(ns, "fullcycle-run",
		testutil.WithAgent(agentv1alpha1.AgentTypeClaude),
		testutil.WithPrompt("echo test-prompt"),
		testutil.WithImage(e2eAgentImage),
		testutil.WithRepository("https://github.com/octocat/Hello-World.git"),
		testutil.WithProviderRef("test-provider"),
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

	// Wait for Job creation (JobName populated)
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
			"app.kubernetes.io/instance": "fullcycle-run",
		}); err != nil {
			return false
		}
		return len(podList.Items) > 0 && podList.Items[0].Spec.NodeName != ""
	})

	// Wait for Succeeded terminal phase (the fake claude binary exits 0 quickly,
	// so the reconciler may transition through Running → Succeeded very fast)
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

	// Prompt args: --permission-mode bypassPermissions --print --prompt "echo test-prompt"
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
}
