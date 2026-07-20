package webhook

import (
	"context"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
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

func sandboxedAgentRunWebhook() *AgentRunWebhook {
	policy := &agentv1alpha1.AgentPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "sandbox-runtime-policy"},
		Spec: agentv1alpha1.AgentPolicySpec{Enforced: agentv1alpha1.AgentPolicyEnforced{
			AllowedRuntimeClassNames: []string{"kata"},
		}},
	}
	return &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(policy).Build()}
}

func TestAgentRunWebhook_Validate_ValidStandalone(t *testing.T) {
	w := &AgentRunWebhook{}
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Image:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RuntimeClassName: &runtimeClassName,
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

func TestAgentRunWebhook_Validate_RequiresPinnedImageAndSandboxRuntime(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*agentv1alpha1.AgentRun)
		wantPhrase string
	}{
		{
			name: "mutable image",
			mutate: func(run *agentv1alpha1.AgentRun) {
				run.Spec.Image = "ghcr.io/xonovex/agent:latest"
			},
			wantPhrase: "immutable @sha256",
		},
		{
			name: "missing runtime class",
			mutate: func(run *agentv1alpha1.AgentRun) {
				run.Spec.RuntimeClassName = nil
			},
			wantPhrase: "sandboxed runtimeClassName",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := baseRun()
			test.mutate(run)

			_, err := (&AgentRunWebhook{}).ValidateCreate(context.Background(), run)
			if err == nil || !strings.Contains(err.Error(), test.wantPhrase) {
				t.Fatalf("ValidateCreate() error = %v, want phrase %q", err, test.wantPhrase)
			}
		})
	}
}

func TestAgentRunWebhook_Default_AppliesInlineHarnessExecutionDefaults(t *testing.T) {
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{Spec: agentv1alpha1.AgentRunSpec{Harness: &agentv1alpha1.AgentSpec{
		DefaultImage:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		DefaultRuntimeClassName: &runtimeClassName,
	}}}

	if err := (&AgentRunWebhook{}).Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Image != run.Spec.Harness.DefaultImage {
		t.Errorf("Image = %q, want harness default %q", run.Spec.Image, run.Spec.Harness.DefaultImage)
	}
	if run.Spec.RuntimeClassName == nil || *run.Spec.RuntimeClassName != runtimeClassName {
		t.Errorf("RuntimeClassName = %v, want %q", run.Spec.RuntimeClassName, runtimeClassName)
	}
}

func runWithNix(nix *agentv1alpha1.NixSpec) *agentv1alpha1.AgentRun {
	runtimeClassName := "kata"
	return &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			RuntimeClassName: &runtimeClassName,
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"},
			},
			Toolchain: &agentv1alpha1.ToolchainSpec{Type: agentv1alpha1.ToolchainTypeNix, Nix: nix},
		},
	}
}

func TestAgentRunWebhook_Validate_NixSpec(t *testing.T) {
	w := sandboxedAgentRunWebhook()
	cases := []struct {
		name    string
		nix     *agentv1alpha1.NixSpec
		wantErr bool
	}{
		{"valid packages", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, false},
		{"valid flake", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", FlakeRef: "/repo", Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, false},
		{"missing rev", &agentv1alpha1.NixSpec{Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"packages and flake", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", Packages: []string{"ripgrep"}, FlakeRef: "/repo", Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"no source", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"missing image", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", Packages: []string{"ripgrep"}}, true},
		{"moving image tag", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent:latest"}, true},
		{"malformed digest", &agentv1alpha1.NixSpec{NixpkgsRev: "abc", Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:not-a-digest"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := w.ValidateCreate(context.Background(), runWithNix(tc.nix))
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateCreate(nix=%+v) err = %v, wantErr %v", tc.nix, err, tc.wantErr)
			}
		})
	}
}

func TestAgentRunWebhook_Validate_ValidWorkspaceRef(t *testing.T) {
	w := sandboxedAgentRunWebhook()
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			WorkspaceRef:     "my-workspace",
			Image:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RuntimeClassName: &runtimeClassName,
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentRunWebhook_Validate_MissingWorkspaceAndWorkspaceRef(t *testing.T) {
	w := sandboxedAgentRunWebhook()
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for missing workspace and workspaceRef")
	}
}

func TestAgentRunWebhook_Validate_MissingRepoURL(t *testing.T) {
	w := sandboxedAgentRunWebhook()
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
	w := sandboxedAgentRunWebhook()
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
	w := sandboxedAgentRunWebhook()
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
	w := sandboxedAgentRunWebhook()
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
	w := sandboxedAgentRunWebhook()
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Image:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RuntimeClassName: &runtimeClassName,
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
	w := sandboxedAgentRunWebhook()
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Image:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RuntimeClassName: &runtimeClassName,
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
	w := sandboxedAgentRunWebhook()
	runtimeClassName := "kata"
	image := "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	oldRun := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Image:            image,
			RuntimeClassName: &runtimeClassName,
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}
	newRun := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Image:            image,
			RuntimeClassName: &runtimeClassName,
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

func TestAgentRunWebhook_ValidateUpdateRejectsExecutionMutation(t *testing.T) {
	w := &AgentRunWebhook{}
	oldRun := baseRun()
	newRun := oldRun.DeepCopy()
	newRun.Spec.Prompt = "changed after execution started"

	if _, err := w.ValidateUpdate(context.Background(), oldRun, newRun); err == nil {
		t.Fatal("expected immutable spec error")
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

func TestAgentRunWebhook_Validate_MaliciousURL(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://example.com/repo.git; rm -rf /",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for malicious URL")
	}
}

func TestAgentRunWebhook_Validate_MaliciousBranch(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL:    "https://github.com/example/repo.git",
					Branch: "main$(whoami)",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for malicious branch")
	}
}

func TestAgentRunWebhook_Validate_MaliciousCommit(t *testing.T) {
	w := &AgentRunWebhook{}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL:    "https://github.com/example/repo.git",
					Commit: "abc1234; cat /etc/passwd",
				},
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), run)
	if err == nil {
		t.Error("ValidateCreate() expected error for malicious commit")
	}
}

func TestAgentRunWebhook_Validate_InlineHarnessOnly(t *testing.T) {
	w := sandboxedAgentRunWebhook()
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Harness: &agentv1alpha1.AgentSpec{
				Type:                    agentv1alpha1.AgentTypeOpencode,
				DefaultImage:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				DefaultRuntimeClassName: &runtimeClassName,
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
