package shared

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildWorkspaceInitJob_Basic(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "my-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git", Branch: "main"},
		},
	}

	job := mustBuildWorkspaceInitJob(t, ws, "my-workspace-ws", "alpine/git:latest", nil)

	if job.Name != "my-workspace-init" {
		t.Errorf("job name = %s, want my-workspace-init", job.Name)
	}
	if job.Labels["app.kubernetes.io/component"] != "workspace-init" {
		t.Errorf("component = %s, want workspace-init", job.Labels["app.kubernetes.io/component"])
	}
	if len(job.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("containers = %d, want 1", len(job.Spec.Template.Spec.Containers))
	}
	container := job.Spec.Template.Spec.Containers[0]
	if container.Name != "git-clone" || container.Image != "alpine/git:latest" {
		t.Errorf("container = %s/%s", container.Name, container.Image)
	}
	if !hasMount(container.VolumeMounts, "workspace", "/workspace") {
		t.Error("expected workspace volume mount at /workspace")
	}
	script := container.Args[1]
	if !strings.Contains(script, "'https://github.com/org/repo.git'") {
		t.Error("expected clone script to contain quoted repo URL")
	}
	if !strings.Contains(script, "--branch 'main'") {
		t.Error("expected clone script to contain quoted branch")
	}
	if !hasPVC(job.Spec.Template.Spec.Volumes, "workspace", "my-workspace-ws") {
		t.Error("expected PVC volume my-workspace-ws")
	}
	if job.Spec.Template.Spec.ServiceAccountName != AgentServiceAccountName ||
		job.Spec.Template.Spec.AutomountServiceAccountToken == nil || *job.Spec.Template.Spec.AutomountServiceAccountToken {
		t.Error("workspace init Job must use the zero-RBAC ServiceAccount without a token")
	}
	if len(container.Resources.Requests) == 0 || len(container.Resources.Limits) == 0 {
		t.Error("workspace init container must have resource requests and limits")
	}
	if !hasMount(container.VolumeMounts, homeVolumeName, agentHome) || !hasMount(container.VolumeMounts, "tmp", "/tmp") {
		t.Error("workspace init container must have writable HOME and /tmp mounts")
	}
}

func TestBuildWorkspaceInitJob_WithJujutsu(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "jj-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git", Branch: "main"},
			Type:       agentv1alpha1.WorkspaceTypeJujutsu,
		},
	}

	job := mustBuildWorkspaceInitJob(t, ws, "jj-workspace-ws", "alpine/git:latest", nil)
	script := job.Spec.Template.Spec.Containers[0].Args[1]
	if !strings.Contains(script, "git clone") {
		t.Error("jj workspace init should still use git clone")
	}
	if !strings.Contains(script, "jj git init --colocate") {
		t.Errorf("jj workspace init missing 'jj git init --colocate', got:\n%s", script)
	}
}

func TestBuildWorkspaceInitJob_RejectsMissingRuntimeClassName(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "my-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git", Branch: "main"},
		},
	}
	_, err := BuildWorkspaceInitJob(ws, "my-workspace-ws", "alpine/git:latest", nil)
	if err == nil {
		t.Fatal("BuildWorkspaceInitJob() error = nil, want missing runtimeClassName error")
	}
}

func TestBuildWorkspaceInitJob_WithRuntimeClassName(t *testing.T) {
	runtimeClass := "kata"
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "my-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository:       agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git", Branch: "main"},
			RuntimeClassName: &runtimeClass,
		},
	}
	job := mustBuildWorkspaceInitJob(t, ws, "my-workspace-ws", "alpine/git:latest", ws.Spec.RuntimeClassName)
	if job.Spec.Template.Spec.RuntimeClassName == nil || *job.Spec.Template.Spec.RuntimeClassName != "kata" {
		t.Error("expected RuntimeClassName kata")
	}
}

func worktreeRun() *agentv1alpha1.AgentRun {
	return &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{Name: "agent-1"}}
}

func TestBuildWorktreeInitContainers_Basic(t *testing.T) {
	containers := mustBuildWorktreeInitContainers(t, worktreeRun(), "node:trixie-slim", agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "main", nil)
	if len(containers) != 1 {
		t.Fatalf("init containers = %d, want 1", len(containers))
	}
	if containers[0].Name != "git-worktree" {
		t.Errorf("name = %s, want git-worktree", containers[0].Name)
	}
	if script := containers[0].Args[1]; !strings.Contains(script, "git worktree add '/workspace-wt/agent-1' -b 'agent-1-work' 'main'") {
		t.Errorf("expected quoted worktree command, got: %s", script)
	}
}

func TestBuildWorktreeInitContainers_DefaultSourceBranch(t *testing.T) {
	containers := mustBuildWorktreeInitContainers(t, worktreeRun(), "node:trixie-slim", agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "", nil)
	if script := containers[0].Args[1]; !strings.Contains(script, "git worktree add '/workspace-wt/agent-1' -b 'agent-1-work' 'HEAD'") {
		t.Errorf("expected HEAD default source branch, got: %s", script)
	}
}

func TestBuildWorktreeInitContainers_WithJujutsu(t *testing.T) {
	containers := mustBuildWorktreeInitContainers(t, worktreeRun(), "node:trixie-slim", agentv1alpha1.WorkspaceTypeJujutsu, "agent-1-work", "main", nil)
	if containers[0].Name != "jj-workspace" {
		t.Errorf("name = %s, want jj-workspace", containers[0].Name)
	}
	script := containers[0].Args[1]
	if !strings.Contains(script, "jj workspace add '/workspace-wt/agent-1' --revision 'main'") {
		t.Errorf("expected quoted jj workspace add, got: %s", script)
	}
	if strings.Contains(script, "git worktree") {
		t.Error("jj mode should not use git worktree")
	}
}

func TestBuildWorkspaceMainContainers_Basic(t *testing.T) {
	containers := mustBuildWorkspaceMainContainers(t, worktreeRun(), nil, "node:trixie-slim", agentv1alpha1.AgentTypeClaude, nil, nil, nil)
	if len(containers) != 1 {
		t.Fatalf("containers = %d, want 1", len(containers))
	}
	c := containers[0]
	if c.WorkingDir != "/workspace-wt/agent-1" {
		t.Errorf("working dir = %s, want /workspace-wt/agent-1", c.WorkingDir)
	}
	if len(c.VolumeMounts) != 2 || c.VolumeMounts[0].Name != "workspace" {
		t.Errorf("expected workspace + tmp mounts, got %d", len(c.VolumeMounts))
	}
}

func TestBuildWorkspaceMainContainers_WithSharedVolumes(t *testing.T) {
	sharedVolumes := []agentv1alpha1.SharedVolumeSpec{
		{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "1Gi"},
		{Name: "opencode-config", MountPath: "/root/.opencode", StorageSize: "512Mi"},
	}
	sharedVolumePVCs := map[string]string{"claude-config": "ws-claude-config", "opencode-config": "ws-opencode-config"}

	containers := mustBuildWorkspaceMainContainers(t, worktreeRun(), nil, "node:trixie-slim", agentv1alpha1.AgentTypeClaude, sharedVolumes, sharedVolumePVCs, nil)
	c := containers[0]
	if len(c.VolumeMounts) != 4 {
		t.Fatalf("volume mounts = %d, want 4", len(c.VolumeMounts))
	}
	foundClaude, foundOpencode := false, false
	for _, vm := range c.VolumeMounts {
		if vm.Name == "claude-config" && vm.MountPath == "/root/.claude" {
			foundClaude = true
		}
		if vm.Name == "opencode-config" && vm.MountPath == "/root/.opencode" {
			foundOpencode = true
		}
	}
	if !foundClaude || !foundOpencode {
		t.Error("expected claude-config and opencode-config volume mounts")
	}
}

func TestBuildWorkspaceJob_Basic(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-1", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace"},
	}

	job := mustBuildJob(t, run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
		WorktreeBranch: "agent-1-work",
		WorkspaceRef:   "my-workspace",
	})

	if job.Name != "agent-1" {
		t.Errorf("job name = %s", job.Name)
	}
	if job.Labels["agent.xonovex.com/workspace"] != "my-workspace" {
		t.Errorf("workspace label = %s, want my-workspace", job.Labels["agent.xonovex.com/workspace"])
	}
	if len(job.Spec.Template.Spec.InitContainers) != 1 || job.Spec.Template.Spec.InitContainers[0].Name != "git-worktree" {
		t.Error("expected one git-worktree init container")
	}
	if job.Spec.Template.Spec.Containers[0].WorkingDir != "/workspace-wt/agent-1" {
		t.Errorf("working dir = %s", job.Spec.Template.Spec.Containers[0].WorkingDir)
	}
	if len(job.Spec.Template.Spec.Volumes) != 3 || !hasPVC(job.Spec.Template.Spec.Volumes, "workspace", "my-workspace-ws") {
		t.Error("expected workspace + tmp + home volumes with my-workspace-ws PVC")
	}
}

func TestBuildWorkspaceJob_WithSharedVolumes(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-1", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace"},
	}
	sharedVolumes := []agentv1alpha1.SharedVolumeSpec{{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "1Gi"}}
	sharedVolumePVCs := map[string]string{"claude-config": "my-workspace-claude-config"}

	job := mustBuildJob(t, run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
		SharedVolumes:    sharedVolumes,
		SharedVolumePVCs: sharedVolumePVCs,
		WorktreeBranch:   "agent-1-work",
		WorkspaceRef:     "my-workspace",
	})

	if len(job.Spec.Template.Spec.Volumes) != 4 {
		t.Fatalf("volumes = %d, want 4", len(job.Spec.Template.Spec.Volumes))
	}
	foundSharedVol := false
	for _, vol := range job.Spec.Template.Spec.Volumes {
		if vol.Name == "claude-config" && vol.PersistentVolumeClaim.ClaimName == "my-workspace-claude-config" {
			foundSharedVol = true
		}
	}
	if !foundSharedVol {
		t.Error("expected claude-config volume with PVC my-workspace-claude-config")
	}
}

func TestBuildWorkspaceInitJob_MountsRepositoryCredentials(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "private-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{Repository: agentv1alpha1.RepositorySpec{
			URL:                  "https://github.com/org/private.git",
			CredentialsSecretRef: &agentv1alpha1.SecretKeyRef{Name: "repo-auth", Key: "credentials"},
		}},
	}

	job := mustBuildWorkspaceInitJob(t, ws, "private-workspace-ws", "alpine/git:latest", nil)

	if !hasMount(job.Spec.Template.Spec.Containers[0].VolumeMounts, repositoryCredentialsVolumeName, "/var/run/agent-repository-credentials") {
		t.Fatal("clone container is missing repository credential mount")
	}
	foundSecret := false
	for _, volume := range job.Spec.Template.Spec.Volumes {
		if volume.Name == repositoryCredentialsVolumeName && volume.Secret != nil && volume.Secret.SecretName == "repo-auth" {
			foundSecret = true
		}
	}
	if !foundSecret {
		t.Fatal("repository credential Secret volume was not created")
	}
}

func hasMount(mounts []corev1.VolumeMount, name, path string) bool {
	for _, mount := range mounts {
		if mount.Name == name && mount.MountPath == path {
			return true
		}
	}
	return false
}

func hasPVC(volumes []corev1.Volume, name, claimName string) bool {
	for _, volume := range volumes {
		if volume.Name == name && volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == claimName {
			return true
		}
	}
	return false
}

func TestBuildWorkspaceJob_WithRuntimeClassName(t *testing.T) {
	runtimeClass := "kata"
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-1", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace", RuntimeClassName: &runtimeClass},
	}

	job := mustBuildJob(t, run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
		WorktreeBranch: "agent-1-work",
		WorkspaceRef:   "my-workspace",
	})

	if job.Spec.Template.Spec.RuntimeClassName == nil || *job.Spec.Template.Spec.RuntimeClassName != "kata" {
		t.Error("expected RuntimeClassName kata")
	}
}

func TestBuildWorkspaceJob_DefaultTTL(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-1", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace"},
	}

	job := mustBuildJob(t, run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
		WorktreeBranch: "agent-1-work",
		WorkspaceRef:   "my-workspace",
	})

	if job.Spec.TTLSecondsAfterFinished == nil || *job.Spec.TTLSecondsAfterFinished != 3600 {
		t.Error("expected default TTL 3600")
	}
}

func TestBuildWorkspaceInitJob_UnknownWorkspaceTypeReturnsError(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "invalid", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Type:       "unknown",
			Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
		},
	}

	_, err := BuildWorkspaceInitJob(ws, "invalid-ws", "image", nil)

	if err == nil {
		t.Fatal("expected unknown workspace type to return an error")
	}
}
