package webhook

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentWorkspaceWebhook_Default_SetsStorageSize(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	if err := webhook.Default(context.Background(), ws); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ws.Spec.StorageSize != "10Gi" {
		t.Errorf("expected default storage size 10Gi, got %s", ws.Spec.StorageSize)
	}
}

func TestAgentWorkspaceWebhook_Default_SetsSharedVolumeStorageSize(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
			SharedVolumes: []agentv1alpha1.SharedVolumeSpec{
				{Name: "claude-config", MountPath: "/root/.claude"},
				{Name: "opencode-config", MountPath: "/root/.opencode", StorageSize: "2Gi"},
			},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	if err := webhook.Default(context.Background(), ws); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ws.Spec.SharedVolumes[0].StorageSize != "1Gi" {
		t.Errorf("expected default shared volume storage size 1Gi, got %s", ws.Spec.SharedVolumes[0].StorageSize)
	}
	if ws.Spec.SharedVolumes[1].StorageSize != "2Gi" {
		t.Errorf("expected preserved storage size 2Gi, got %s", ws.Spec.SharedVolumes[1].StorageSize)
	}
}

func TestAgentWorkspaceWebhook_Default_PreservesExistingValues(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository:  agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
			StorageSize: "50Gi",
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	if err := webhook.Default(context.Background(), ws); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ws.Spec.StorageSize != "50Gi" {
		t.Errorf("expected preserved storage size 50Gi, got %s", ws.Spec.StorageSize)
	}
}

func TestAgentWorkspaceWebhook_Validate_Valid(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateCreate(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentWorkspaceWebhook_Validate_ValidWithSharedVolumes(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
			SharedVolumes: []agentv1alpha1.SharedVolumeSpec{
				{Name: "claude-config", MountPath: "/root/.claude"},
				{Name: "opencode-config", MountPath: "/root/.opencode"},
			},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateCreate(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentWorkspaceWebhook_Validate_MissingRepoURL(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateCreate(context.Background(), ws)
	if err == nil {
		t.Fatal("expected error for missing repository URL")
	}
}

func TestAgentWorkspaceWebhook_Validate_EmptySharedVolumeName(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
			SharedVolumes: []agentv1alpha1.SharedVolumeSpec{
				{Name: "", MountPath: "/root/.claude"},
			},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateCreate(context.Background(), ws)
	if err == nil {
		t.Fatal("expected error for empty shared volume name")
	}
}

func TestAgentWorkspaceWebhook_Validate_EmptySharedVolumeMountPath(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
			SharedVolumes: []agentv1alpha1.SharedVolumeSpec{
				{Name: "claude-config", MountPath: ""},
			},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateCreate(context.Background(), ws)
	if err == nil {
		t.Fatal("expected error for empty shared volume mount path")
	}
}

func TestAgentWorkspaceWebhook_Validate_DuplicateSharedVolumeName(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
			SharedVolumes: []agentv1alpha1.SharedVolumeSpec{
				{Name: "claude-config", MountPath: "/root/.claude"},
				{Name: "claude-config", MountPath: "/opt/.claude"},
			},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateCreate(context.Background(), ws)
	if err == nil {
		t.Fatal("expected error for duplicate shared volume name")
	}
}

func TestAgentWorkspaceWebhook_ValidateUpdate(t *testing.T) {
	old := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
		},
	}
	new := &agentv1alpha1.AgentWorkspace{
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo2.git"},
		},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateUpdate(context.Background(), old, new)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentWorkspaceWebhook_ValidateDelete(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}

	webhook := &AgentWorkspaceWebhook{}
	_, err := webhook.ValidateDelete(context.Background(), ws)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
