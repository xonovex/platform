package shared

import (
	"strings"
	"testing"
	"time"

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

	job := BuildWorkspaceInitJob(ws, "my-workspace-ws", "alpine/git:latest", nil)

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
	if len(container.VolumeMounts) != 1 || container.VolumeMounts[0].MountPath != "/workspace" {
		t.Error("expected workspace volume mount at /workspace")
	}
	script := container.Args[1]
	if !strings.Contains(script, "'https://github.com/org/repo.git'") {
		t.Error("expected clone script to contain quoted repo URL")
	}
	if !strings.Contains(script, "--branch 'main'") {
		t.Error("expected clone script to contain quoted branch")
	}
	if len(job.Spec.Template.Spec.Volumes) != 1 || job.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "my-workspace-ws" {
		t.Error("expected PVC volume my-workspace-ws")
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

	job := BuildWorkspaceInitJob(ws, "jj-workspace-ws", "alpine/git:latest", nil)
	script := job.Spec.Template.Spec.Containers[0].Args[1]
	if !strings.Contains(script, "git clone") {
		t.Error("jj workspace init should still use git clone")
	}
	if !strings.Contains(script, "jj git init --colocate") {
		t.Errorf("jj workspace init missing 'jj git init --colocate', got:\n%s", script)
	}
}

func TestBuildWorkspaceInitJob_NoRuntimeClassName(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "my-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git", Branch: "main"},
		},
	}
	job := BuildWorkspaceInitJob(ws, "my-workspace-ws", "alpine/git:latest", nil)
	if job.Spec.Template.Spec.RuntimeClassName != nil {
		t.Errorf("expected nil RuntimeClassName, got %q", *job.Spec.Template.Spec.RuntimeClassName)
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
	job := BuildWorkspaceInitJob(ws, "my-workspace-ws", "alpine/git:latest", ws.Spec.RuntimeClassName)
	if job.Spec.Template.Spec.RuntimeClassName == nil || *job.Spec.Template.Spec.RuntimeClassName != "kata" {
		t.Error("expected RuntimeClassName kata")
	}
}

func worktreeRun() *agentv1alpha1.AgentRun {
	return &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{Name: "agent-1"}}
}

func TestBuildWorktreeInitContainers_Basic(t *testing.T) {
	containers := BuildWorktreeInitContainers(worktreeRun(), "node:trixie-slim", agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "main", nil)
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
	containers := BuildWorktreeInitContainers(worktreeRun(), "node:trixie-slim", agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "", nil)
	if script := containers[0].Args[1]; !strings.Contains(script, "git worktree add '/workspace-wt/agent-1' -b 'agent-1-work' 'HEAD'") {
		t.Errorf("expected HEAD default source branch, got: %s", script)
	}
}

func TestBuildWorktreeInitContainers_WithJujutsu(t *testing.T) {
	containers := BuildWorktreeInitContainers(worktreeRun(), "node:trixie-slim", agentv1alpha1.WorkspaceTypeJujutsu, "agent-1-work", "main", nil)
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
	containers := BuildWorkspaceMainContainers(worktreeRun(), nil, "node:trixie-slim", agentv1alpha1.AgentTypeClaude, nil, nil, nil)
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

	containers := BuildWorkspaceMainContainers(worktreeRun(), nil, "node:trixie-slim", agentv1alpha1.AgentTypeClaude, sharedVolumes, sharedVolumePVCs, nil)
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

	job := BuildJob(run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
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
	if len(job.Spec.Template.Spec.Volumes) != 2 || job.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "my-workspace-ws" {
		t.Error("expected workspace + tmp volumes with my-workspace-ws PVC")
	}
}

func TestBuildWorkspaceJob_WithSharedVolumes(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-1", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace"},
	}
	sharedVolumes := []agentv1alpha1.SharedVolumeSpec{{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "1Gi"}}
	sharedVolumePVCs := map[string]string{"claude-config": "my-workspace-claude-config"}

	job := BuildJob(run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
		SharedVolumes:    sharedVolumes,
		SharedVolumePVCs: sharedVolumePVCs,
		WorktreeBranch:   "agent-1-work",
		WorkspaceRef:     "my-workspace",
	})

	if len(job.Spec.Template.Spec.Volumes) != 3 {
		t.Fatalf("volumes = %d, want 3", len(job.Spec.Template.Spec.Volumes))
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

func TestBuildWorkspaceJob_WithRuntimeClassName(t *testing.T) {
	runtimeClass := "kata"
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-1", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{WorkspaceRef: "my-workspace", RuntimeClassName: &runtimeClass},
	}

	job := BuildJob(run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
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

	job := BuildJob(run, nil, "my-workspace-ws", "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, nil, nil, &WorkspaceBinding{
		WorktreeBranch: "agent-1-work",
		WorkspaceRef:   "my-workspace",
	})

	if job.Spec.TTLSecondsAfterFinished == nil || *job.Spec.TTLSecondsAfterFinished != 3600 {
		t.Error("expected default TTL 3600")
	}
}
