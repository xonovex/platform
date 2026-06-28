package shared

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildJob_Basic(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-run",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	job := BuildJob(run, nil, "test-pvc", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	if job.Name != "test-run" {
		t.Errorf("job name = %q, want %q", job.Name, "test-run")
	}
	if job.Namespace != "default" {
		t.Errorf("job namespace = %q, want %q", job.Namespace, "default")
	}
	if job.Labels["agent.xonovex.com/agent-type"] != "claude" {
		t.Errorf("agent-type label = %q, want %q", job.Labels["agent.xonovex.com/agent-type"], "claude")
	}

	expectedTimeout := int64(3600)
	if *job.Spec.ActiveDeadlineSeconds != expectedTimeout {
		t.Errorf("ActiveDeadlineSeconds = %d, want %d", *job.Spec.ActiveDeadlineSeconds, expectedTimeout)
	}
	if *job.Spec.BackoffLimit != 0 {
		t.Errorf("BackoffLimit = %d, want 0", *job.Spec.BackoffLimit)
	}

	podSpec := job.Spec.Template.Spec
	if podSpec.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("RestartPolicy = %q, want %q", podSpec.RestartPolicy, corev1.RestartPolicyNever)
	}
	if len(podSpec.InitContainers) != 1 {
		t.Fatalf("len(InitContainers) = %d, want 1", len(podSpec.InitContainers))
	}
	if len(podSpec.Containers) != 1 {
		t.Fatalf("len(Containers) = %d, want 1", len(podSpec.Containers))
	}
	if podSpec.Volumes[0].PersistentVolumeClaim.ClaimName != "test-pvc" {
		t.Errorf("PVC claim = %q, want %q", podSpec.Volumes[0].PersistentVolumeClaim.ClaimName, "test-pvc")
	}
}

func TestBuildJob_CustomTimeout(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", 30*time.Minute, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	expected := int64(1800)
	if *job.Spec.ActiveDeadlineSeconds != expected {
		t.Errorf("ActiveDeadlineSeconds = %d, want %d", *job.Spec.ActiveDeadlineSeconds, expected)
	}
}

func TestBuildJob_CustomImage(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "custom-image:latest", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	initImage := job.Spec.Template.Spec.InitContainers[0].Image
	if initImage != "custom-image:latest" {
		t.Errorf("init container image = %q, want %q", initImage, "custom-image:latest")
	}
	mainImage := job.Spec.Template.Spec.Containers[0].Image
	if mainImage != "custom-image:latest" {
		t.Errorf("main container image = %q, want %q", mainImage, "custom-image:latest")
	}
}

func TestBuildJob_WithResources(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	resources := job.Spec.Template.Spec.Containers[0].Resources
	if resources.Requests.Cpu().String() != "500m" {
		t.Errorf("CPU request = %q, want %q", resources.Requests.Cpu().String(), "500m")
	}
	if resources.Limits.Memory().String() != "4Gi" {
		t.Errorf("memory limit = %q, want %q", resources.Limits.Memory().String(), "4Gi")
	}
}

func TestBuildJob_NodeSelectorAndTolerations(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
			NodeSelector: map[string]string{"gpu": "true"},
			Tolerations: []corev1.Toleration{
				{Key: "gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	podSpec := job.Spec.Template.Spec
	if podSpec.NodeSelector["gpu"] != "true" {
		t.Errorf("NodeSelector[gpu] = %q, want %q", podSpec.NodeSelector["gpu"], "true")
	}
	if len(podSpec.Tolerations) != 1 {
		t.Fatalf("len(Tolerations) = %d, want 1", len(podSpec.Tolerations))
	}
	if podSpec.Tolerations[0].Key != "gpu" {
		t.Errorf("Toleration key = %q, want %q", podSpec.Tolerations[0].Key, "gpu")
	}
}

func TestBuildJob_WithRuntimeClassName(t *testing.T) {
	runtimeClass := "kata"
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
			RuntimeClassName: &runtimeClass,
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	if job.Spec.Template.Spec.RuntimeClassName == nil {
		t.Fatal("expected RuntimeClassName to be set")
	}
	if *job.Spec.Template.Spec.RuntimeClassName != "kata" {
		t.Errorf("RuntimeClassName = %q, want %q", *job.Spec.Template.Spec.RuntimeClassName, "kata")
	}
}

func TestBuildJob_WithoutRuntimeClassName(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	if job.Spec.Template.Spec.RuntimeClassName != nil {
		t.Errorf("expected RuntimeClassName to be nil, got %q", *job.Spec.Template.Spec.RuntimeClassName)
	}
}

func TestBuildJob_WithNixImage(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix: &agentv1alpha1.NixSpec{
			NixpkgsRev: "abc123",
			Packages:   []string{"nodejs_22", "python3"},
			Image:      "ghcr.io/xonovex/agent@sha256:abc",
		},
	}

	job := BuildJob(run, nil, "pvc", "ghcr.io/xonovex/agent@sha256:abc", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, tc, nil, nil)
	podSpec := job.Spec.Template.Spec

	// Image-based nix: NO nix-env init container or volume (per-pod install gone).
	for _, c := range podSpec.InitContainers {
		if c.Name == "nix-env" {
			t.Error("image-based nix must not add a nix-env init container")
		}
	}
	foundHome := false
	for _, v := range podSpec.Volumes {
		if v.Name == "nix-env" {
			t.Error("image-based nix must not add a nix-env volume")
		}
		if v.Name == homeVolumeName && v.EmptyDir != nil {
			foundHome = true
		}
	}
	// Writable HOME emptyDir reconciles readOnlyRootFilesystem with the uid-1000 image.
	if !foundHome {
		t.Error("expected a writable home emptyDir")
	}
	if podSpec.SecurityContext == nil || podSpec.SecurityContext.FSGroup == nil || *podSpec.SecurityContext.FSGroup != 1000 {
		t.Error("expected fsGroup=1000 so uid 1000 owns the HOME emptyDir")
	}
	// Zero-RBAC ServiceAccount, no mounted token.
	if podSpec.ServiceAccountName != AgentServiceAccountName {
		t.Errorf("ServiceAccountName = %q, want %q", podSpec.ServiceAccountName, AgentServiceAccountName)
	}
	if podSpec.AutomountServiceAccountToken == nil || *podSpec.AutomountServiceAccountToken {
		t.Error("AutomountServiceAccountToken must be false (agent never calls the K8s API)")
	}
	// Default resource bounds applied when the run requested none.
	if len(podSpec.Containers[0].Resources.Limits) == 0 {
		t.Error("expected default resource limits (node-DoS bound)")
	}
}

func TestBuildJob_WithoutNix(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	if len(job.Spec.Template.Spec.Volumes) != 2 {
		t.Errorf("len(Volumes) = %d, want 2 (workspace + tmp)", len(job.Spec.Template.Spec.Volumes))
	}
}

func TestBuildJob_Labels(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "my-run", Namespace: "ns"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeOpencode, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	expectedLabels := map[string]string{
		"app.kubernetes.io/name":       "agent-operator",
		"app.kubernetes.io/instance":   "my-run",
		"app.kubernetes.io/component":  "agent-run",
		"agent.xonovex.com/agent-type": "opencode",
	}
	for k, want := range expectedLabels {
		if got := job.Labels[k]; got != want {
			t.Errorf("job label %q = %q, want %q", k, got, want)
		}
		if got := job.Spec.Template.Labels[k]; got != want {
			t.Errorf("pod template label %q = %q, want %q", k, got, want)
		}
	}
}

func TestBuildJob_PodSecurityContext(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	psc := job.Spec.Template.Spec.SecurityContext
	if psc == nil {
		t.Fatal("PodSecurityContext should not be nil")
	}
	if psc.RunAsNonRoot == nil || *psc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot should be true")
	}
	if psc.SeccompProfile == nil || psc.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Error("SeccompProfile should be RuntimeDefault")
	}
}

func TestBuildJob_TmpVolume(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	foundTmp := false
	for _, v := range job.Spec.Template.Spec.Volumes {
		if v.Name == "tmp" && v.EmptyDir != nil {
			foundTmp = true
		}
	}
	if !foundTmp {
		t.Error("expected tmp EmptyDir volume")
	}
}

func TestBuildJob_ContainerSecurityContext(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	// Check main container
	mainSC := job.Spec.Template.Spec.Containers[0].SecurityContext
	if mainSC == nil {
		t.Fatal("main container SecurityContext should not be nil")
	}
	if *mainSC.AllowPrivilegeEscalation != false {
		t.Error("main container AllowPrivilegeEscalation should be false")
	}

	// Check init container
	initSC := job.Spec.Template.Spec.InitContainers[0].SecurityContext
	if initSC == nil {
		t.Fatal("init container SecurityContext should not be nil")
	}
	if *initSC.AllowPrivilegeEscalation != false {
		t.Error("init container AllowPrivilegeEscalation should be false")
	}
}

func TestBuildJob_DefaultTTL(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, nil)

	if job.Spec.TTLSecondsAfterFinished == nil {
		t.Fatal("TTLSecondsAfterFinished should not be nil")
	}
	if *job.Spec.TTLSecondsAfterFinished != 3600 {
		t.Errorf("TTLSecondsAfterFinished = %d, want 3600", *job.Spec.TTLSecondsAfterFinished)
	}
}

func TestBuildJob_ExplicitTTL(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test-run", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	ttl := int32(0)
	job := BuildJob(run, nil, "pvc", "image", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, &ttl, nil)

	if job.Spec.TTLSecondsAfterFinished == nil {
		t.Fatal("TTLSecondsAfterFinished should not be nil")
	}
	if *job.Spec.TTLSecondsAfterFinished != 0 {
		t.Errorf("TTLSecondsAfterFinished = %d, want 0", *job.Spec.TTLSecondsAfterFinished)
	}
}
