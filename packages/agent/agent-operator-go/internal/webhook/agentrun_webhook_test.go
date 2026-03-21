package webhook

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentRunWebhook_Default_SetsTimeout(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}

	if err := w.Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Timeout == nil {
		t.Fatal("Timeout is nil, want non-nil")
	}
	if run.Spec.Timeout.Duration != time.Hour {
		t.Errorf("Timeout = %v, want %v", run.Spec.Timeout.Duration, time.Hour)
	}
}

func TestAgentRunWebhook_Default_PreservesExistingValues(t *testing.T) {
	w := &AgentRunWebhook{}
	customTimeout := metav1.Duration{Duration: 30 * time.Minute}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Timeout: &customTimeout,
		},
	}

	if err := w.Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Timeout.Duration != 30*time.Minute {
		t.Errorf("Timeout = %v, want %v (should not override)", run.Spec.Timeout.Duration, 30*time.Minute)
	}
}

func TestAgentRunWebhook_Validate_ValidStandalone(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	warnings, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("ValidateCreate() warnings = %v, want none", warnings)
	}
}

func TestAgentRunWebhook_Validate_ValidWorkspaceRef(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			WorkspaceRef: "my-workspace",
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentRunWebhook_Validate_MissingWorkspaceAndWorkspaceRef(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for missing workspace and workspaceRef")
	}
}

func TestAgentRunWebhook_Validate_MissingRepoURL(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for missing repo URL in inline workspace")
	}
}

func TestAgentRunWebhook_Validate_BothWorkspaceRefAndInline(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			WorkspaceRef: "my-workspace",
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for both workspaceRef and inline workspace")
	}
}

func TestAgentRunWebhook_Validate_BothHarnessRefAndInline(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			HarnessRef: "my-harness",
			Harness: &agentv1alpha1.AgentSpec{
				Type: agentv1alpha1.AgentTypeClaude,
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for both harnessRef and inline harness")
	}
}

func TestAgentRunWebhook_Validate_BothProviderRefAndInline(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "my-provider",
			Provider: &agentv1alpha1.ProviderSpec{
				Type: "gemini",
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for both providerRef and inline provider")
	}
}

func TestAgentRunWebhook_Validate_BothToolchainRefAndInline(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			ToolchainRef: "my-toolchain",
			Toolchain: &agentv1alpha1.ToolchainSpec{
				Type: agentv1alpha1.ToolchainTypeNix,
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for both toolchainRef and inline toolchain")
	}
}

func TestAgentRunWebhook_Validate_InvalidAgentType(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Harness: &agentv1alpha1.AgentSpec{
				Type: "invalid",
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for invalid agent type")
	}
}

func TestAgentRunWebhook_Validate_ValidAgentType(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Harness: &agentv1alpha1.AgentSpec{
				Type: agentv1alpha1.AgentTypeClaude,
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentRunWebhook_Validate_InvalidWorkspaceType(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Type: "svn",
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for invalid workspace type")
	}
}

func TestAgentRunWebhook_Validate_ValidWorkspaceTypeJujutsu(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Type: agentv1alpha1.WorkspaceTypeJujutsu,
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentRunWebhook_Validate_InvalidToolchainType(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Toolchain: &agentv1alpha1.ToolchainSpec{
				Type: "invalid",
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for invalid toolchain type")
	}
}

func TestAgentRunWebhook_ValidateUpdate(t *testing.T) {
	w := &AgentRunWebhook{}
	oldRun := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}
	newRun := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	_, err := w.ValidateUpdate(context.Background(), oldRun, newRun)
	if err != nil {
		t.Errorf("ValidateUpdate() error = %v", err)
	}
}

func TestAgentRunWebhook_ValidateDelete(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{}

	_, err := w.ValidateDelete(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateDelete() error = %v", err)
	}
}

func TestAgentRunWebhook_Validate_InlineHarnessOnly(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Harness: &agentv1alpha1.AgentSpec{
				Type:         agentv1alpha1.AgentTypeOpencode,
				DefaultImage: "custom:latest",
			},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}
