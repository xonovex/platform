package builder

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildWorkspacePVC_Basic(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
			UID:       types.UID("test-uid"),
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			StorageSize: "10Gi",
		},
	}

	pvc := BuildWorkspacePVC("my-workspace-ws", ws)

	if pvc.Name != "my-workspace-ws" {
		t.Errorf("expected name my-workspace-ws, got %s", pvc.Name)
	}
	if pvc.Namespace != "default" {
		t.Errorf("expected namespace default, got %s", pvc.Namespace)
	}
	if pvc.Labels["app.kubernetes.io/component"] != "workspace" {
		t.Errorf("expected component label workspace, got %s", pvc.Labels["app.kubernetes.io/component"])
	}
	if pvc.Labels["app.kubernetes.io/instance"] != "my-workspace" {
		t.Errorf("expected instance label my-workspace, got %s", pvc.Labels["app.kubernetes.io/instance"])
	}

	// Check RWX access mode
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteMany {
		t.Errorf("expected ReadWriteMany access mode, got %v", pvc.Spec.AccessModes)
	}

	// Check owner reference
	if len(pvc.OwnerReferences) != 1 {
		t.Fatalf("expected 1 owner reference, got %d", len(pvc.OwnerReferences))
	}
	if pvc.OwnerReferences[0].Kind != "AgentWorkspace" {
		t.Errorf("expected owner kind AgentWorkspace, got %s", pvc.OwnerReferences[0].Kind)
	}
	if pvc.OwnerReferences[0].Name != "my-workspace" {
		t.Errorf("expected owner name my-workspace, got %s", pvc.OwnerReferences[0].Name)
	}

	// Check storage size
	expectedSize := resource.MustParse("10Gi")
	actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if !actualSize.Equal(expectedSize) {
		t.Errorf("expected storage size 10Gi, got %s", actualSize.String())
	}
}

func TestBuildWorkspacePVC_DefaultStorageSize(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
			UID:       types.UID("test-uid"),
		},
	}

	pvc := BuildWorkspacePVC("my-workspace-ws", ws)

	expectedSize := resource.MustParse("10Gi")
	actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if !actualSize.Equal(expectedSize) {
		t.Errorf("expected default storage size 10Gi, got %s", actualSize.String())
	}
}

func TestBuildWorkspacePVC_WithStorageClass(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
			UID:       types.UID("test-uid"),
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			StorageClass: "nfs-csi",
			StorageSize:  "20Gi",
		},
	}

	pvc := BuildWorkspacePVC("my-workspace-ws", ws)

	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "nfs-csi" {
		t.Errorf("expected storage class nfs-csi")
	}
}

func TestBuildSharedVolumePVC_Basic(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
			UID:       types.UID("test-uid"),
		},
	}

	vol := agentv1alpha1.SharedVolumeSpec{
		Name:        "claude-config",
		MountPath:   "/root/.claude",
		StorageSize: "2Gi",
	}

	pvc := BuildSharedVolumePVC("my-workspace-claude-config", ws, vol)

	if pvc.Name != "my-workspace-claude-config" {
		t.Errorf("expected name my-workspace-claude-config, got %s", pvc.Name)
	}
	if pvc.Labels["app.kubernetes.io/component"] != "shared-volume" {
		t.Errorf("expected component label shared-volume, got %s", pvc.Labels["app.kubernetes.io/component"])
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != corev1.ReadWriteMany {
		t.Errorf("expected ReadWriteMany access mode")
	}
	if pvc.OwnerReferences[0].Kind != "AgentWorkspace" {
		t.Errorf("expected owner kind AgentWorkspace, got %s", pvc.OwnerReferences[0].Kind)
	}

	expectedSize := resource.MustParse("2Gi")
	actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if !actualSize.Equal(expectedSize) {
		t.Errorf("expected storage size 2Gi, got %s", actualSize.String())
	}
}

func TestBuildSharedVolumePVC_DefaultStorageSize(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
			UID:       types.UID("test-uid"),
		},
	}

	vol := agentv1alpha1.SharedVolumeSpec{
		Name:      "claude-config",
		MountPath: "/root/.claude",
	}

	pvc := BuildSharedVolumePVC("my-workspace-claude-config", ws, vol)

	expectedSize := resource.MustParse("1Gi")
	actualSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if !actualSize.Equal(expectedSize) {
		t.Errorf("expected default storage size 1Gi, got %s", actualSize.String())
	}
}

func TestBuildWorkspaceInitJob_Basic(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL:    "https://github.com/org/repo.git",
				Branch: "main",
			},
		},
	}

	job := BuildWorkspaceInitJob(ws, "my-workspace-ws", "alpine/git:latest")

	if job.Name != "my-workspace-init" {
		t.Errorf("expected job name my-workspace-init, got %s", job.Name)
	}
	if job.Labels["app.kubernetes.io/component"] != "workspace-init" {
		t.Errorf("expected component label workspace-init, got %s", job.Labels["app.kubernetes.io/component"])
	}

	// Check single container (not init container)
	if len(job.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(job.Spec.Template.Spec.Containers))
	}
	container := job.Spec.Template.Spec.Containers[0]
	if container.Name != "git-clone" {
		t.Errorf("expected container name git-clone, got %s", container.Name)
	}
	if container.Image != "alpine/git:latest" {
		t.Errorf("expected image alpine/git:latest, got %s", container.Image)
	}

	// Check volume mount
	if len(container.VolumeMounts) != 1 || container.VolumeMounts[0].MountPath != "/workspace" {
		t.Errorf("expected workspace volume mount at /workspace")
	}

	// Check clone script contains repo URL
	script := container.Args[1]
	if !containsStr(script, "'https://github.com/org/repo.git'") {
		t.Errorf("expected clone script to contain quoted repo URL")
	}
	if !containsStr(script, "--branch 'main'") {
		t.Errorf("expected clone script to contain quoted branch")
	}

	// Check PVC volume
	if len(job.Spec.Template.Spec.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(job.Spec.Template.Spec.Volumes))
	}
	if job.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "my-workspace-ws" {
		t.Errorf("expected PVC claim name my-workspace-ws")
	}
}

func TestBuildWorktreeInitContainers_Basic(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "agent-1",
		},
	}

	containers := BuildWorktreeInitContainers(run, "node:trixie-slim", agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "main")

	if len(containers) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(containers))
	}
	container := containers[0]
	if container.Name != "git-worktree" {
		t.Errorf("expected container name git-worktree, got %s", container.Name)
	}

	script := container.Args[1]
	if !containsStr(script, "git worktree add '/workspace-wt/agent-1' -b 'agent-1-work' 'main'") {
		t.Errorf("expected quoted worktree command in script, got: %s", script)
	}
}

func TestBuildWorktreeInitContainers_DefaultSourceBranch(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "agent-1",
		},
	}

	containers := BuildWorktreeInitContainers(run, "node:trixie-slim", agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "")
	script := containers[0].Args[1]
	if !containsStr(script, "git worktree add '/workspace-wt/agent-1' -b 'agent-1-work' 'HEAD'") {
		t.Errorf("expected quoted HEAD as default source branch, got: %s", script)
	}
}

func TestBuildWorkspaceMainContainers_Basic(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "agent-1",
		},
	}

	containers := BuildWorkspaceMainContainers(run, nil, "node:trixie-slim", agentv1alpha1.AgentTypeClaude, nil, nil, nil)

	if len(containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(containers))
	}
	container := containers[0]
	if container.WorkingDir != "/workspace-wt/agent-1" {
		t.Errorf("expected working dir /workspace-wt/agent-1, got %s", container.WorkingDir)
	}
	// Should have workspace volume mount
	if len(container.VolumeMounts) != 1 || container.VolumeMounts[0].Name != "workspace" {
		t.Errorf("expected workspace volume mount")
	}
}

func TestBuildWorkspaceMainContainers_WithSharedVolumes(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "agent-1",
		},
	}

	sharedVolumes := []agentv1alpha1.SharedVolumeSpec{
		{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "1Gi"},
		{Name: "opencode-config", MountPath: "/root/.opencode", StorageSize: "512Mi"},
	}
	sharedVolumePVCs := map[string]string{
		"claude-config":   "ws-claude-config",
		"opencode-config": "ws-opencode-config",
	}

	containers := BuildWorkspaceMainContainers(run, nil, "node:trixie-slim", agentv1alpha1.AgentTypeClaude, sharedVolumes, sharedVolumePVCs, nil)

	container := containers[0]
	// workspace + 2 shared volumes = 3 volume mounts
	if len(container.VolumeMounts) != 3 {
		t.Fatalf("expected 3 volume mounts, got %d", len(container.VolumeMounts))
	}

	foundClaude := false
	foundOpencode := false
	for _, vm := range container.VolumeMounts {
		if vm.Name == "claude-config" && vm.MountPath == "/root/.claude" {
			foundClaude = true
		}
		if vm.Name == "opencode-config" && vm.MountPath == "/root/.opencode" {
			foundOpencode = true
		}
	}
	if !foundClaude {
		t.Error("expected claude-config volume mount at /root/.claude")
	}
	if !foundOpencode {
		t.Error("expected opencode-config volume mount at /root/.opencode")
	}
}

func TestBuildWorkspaceJob_Basic(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-1",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentRunSpec{
			WorkspaceRef: "my-workspace",
		},
	}

	job := BuildWorkspaceJob(run, nil, "my-workspace-ws", nil, nil, "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "", nil)

	if job.Name != "agent-1" {
		t.Errorf("expected job name agent-1, got %s", job.Name)
	}
	if job.Labels["agent.xonovex.com/workspace"] != "my-workspace" {
		t.Errorf("expected workspace label my-workspace, got %s", job.Labels["agent.xonovex.com/workspace"])
	}

	// Check init containers (worktree setup)
	if len(job.Spec.Template.Spec.InitContainers) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(job.Spec.Template.Spec.InitContainers))
	}
	if job.Spec.Template.Spec.InitContainers[0].Name != "git-worktree" {
		t.Errorf("expected init container git-worktree")
	}

	// Check main container working dir
	if job.Spec.Template.Spec.Containers[0].WorkingDir != "/workspace-wt/agent-1" {
		t.Errorf("expected working dir /workspace-wt/agent-1, got %s", job.Spec.Template.Spec.Containers[0].WorkingDir)
	}

	// Check workspace volume
	if len(job.Spec.Template.Spec.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(job.Spec.Template.Spec.Volumes))
	}
	if job.Spec.Template.Spec.Volumes[0].PersistentVolumeClaim.ClaimName != "my-workspace-ws" {
		t.Errorf("expected workspace PVC claim my-workspace-ws")
	}
}

func TestBuildWorkspaceJob_WithSharedVolumes(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-1",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentRunSpec{
			WorkspaceRef: "my-workspace",
		},
	}

	sharedVolumes := []agentv1alpha1.SharedVolumeSpec{
		{Name: "claude-config", MountPath: "/root/.claude", StorageSize: "1Gi"},
	}
	sharedVolumePVCs := map[string]string{
		"claude-config": "my-workspace-claude-config",
	}

	job := BuildWorkspaceJob(run, nil, "my-workspace-ws", sharedVolumes, sharedVolumePVCs, "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "", nil)

	// workspace + shared volume = 2 volumes
	if len(job.Spec.Template.Spec.Volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(job.Spec.Template.Spec.Volumes))
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
		Spec: agentv1alpha1.AgentRunSpec{
			WorkspaceRef:     "my-workspace",
			RuntimeClassName: &runtimeClass,
		},
	}

	job := BuildWorkspaceJob(run, nil, "my-workspace-ws", nil, nil, "node:trixie-slim", time.Hour, agentv1alpha1.AgentTypeClaude, agentv1alpha1.WorkspaceTypeGit, "agent-1-work", "", nil)

	if job.Spec.Template.Spec.RuntimeClassName == nil {
		t.Fatal("expected RuntimeClassName to be set")
	}
	if *job.Spec.Template.Spec.RuntimeClassName != "kata" {
		t.Errorf("RuntimeClassName = %q, want %q", *job.Spec.Template.Spec.RuntimeClassName, "kata")
	}
}

func TestBuildWorkspaceInitJob_WithJujutsu(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jj-workspace",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL:    "https://github.com/org/repo.git",
				Branch: "main",
			},
			Type: agentv1alpha1.WorkspaceTypeJujutsu,
		},
	}

	job := BuildWorkspaceInitJob(ws, "jj-workspace-ws", "alpine/git:latest")

	script := job.Spec.Template.Spec.Containers[0].Args[1]
	if !containsStr(script, "git clone") {
		t.Error("jj workspace init should still use git clone")
	}
	if !containsStr(script, "jj git init --colocate") {
		t.Errorf("jj workspace init missing 'jj git init --colocate', got:\n%s", script)
	}
}

func TestBuildWorktreeInitContainers_WithJujutsu(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "agent-1",
		},
	}

	containers := BuildWorktreeInitContainers(run, "node:trixie-slim", agentv1alpha1.WorkspaceTypeJujutsu, "agent-1-work", "main")

	if len(containers) != 1 {
		t.Fatalf("expected 1 init container, got %d", len(containers))
	}
	container := containers[0]
	if container.Name != "jj-workspace" {
		t.Errorf("expected container name jj-workspace, got %s", container.Name)
	}

	script := container.Args[1]
	if !containsStr(script, "jj workspace add '/workspace-wt/agent-1' --revision 'main'") {
		t.Errorf("expected quoted jj workspace add command, got: %s", script)
	}
	if containsStr(script, "git worktree") {
		t.Error("jj mode should not use git worktree")
	}
}

func TestBuildWorkspaceInitJob_NoRuntimeClassName(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "my-workspace", Namespace: "default"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL:    "https://github.com/org/repo.git",
				Branch: "main",
			},
		},
	}

	job := BuildWorkspaceInitJob(ws, "my-workspace-ws", "alpine/git:latest")

	if job.Spec.Template.Spec.RuntimeClassName != nil {
		t.Errorf("expected init job RuntimeClassName to be nil, got %q", *job.Spec.Template.Spec.RuntimeClassName)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
