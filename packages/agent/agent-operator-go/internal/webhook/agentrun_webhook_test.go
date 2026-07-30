package webhook

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

const testNixRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

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

func sandboxedAgentRunWebhook(namespaces ...string) *AgentRunWebhook {
	namespace := ""
	if len(namespaces) > 0 {
		namespace = namespaces[0]
	}
	policy := &agentv1alpha1.AgentPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "sandbox-runtime-policy", Namespace: namespace},
		Spec: agentv1alpha1.AgentPolicySpec{Enforced: agentv1alpha1.AgentPolicyEnforced{
			AllowedRuntimeClassNames: []string{"kata"},
		}},
	}
	return &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(policy).Build()}
}

func TestAgentRunWebhook_Validate_ValidStandalone(t *testing.T) {
	w := sandboxedAgentRunWebhook()
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

			_, err := sandboxedAgentRunWebhook(run.Namespace).ValidateCreate(context.Background(), run)
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

func TestAgentRunWebhook_Default_SnapshotsReferencedToolchain(t *testing.T) {
	toolchain := &agentv1alpha1.AgentToolchain{
		ObjectMeta: metav1.ObjectMeta{Name: "pinned", Namespace: "test"},
		Spec: agentv1alpha1.ToolchainSpec{
			Type: agentv1alpha1.ToolchainTypeNix,
			Nix: &agentv1alpha1.NixSpec{
				NixpkgsRev: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Packages:   []string{"ripgrep"},
				Image:      "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
	}
	w := &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(toolchain).Build()}
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec:       agentv1alpha1.AgentRunSpec{ToolchainRef: toolchain.Name},
	}

	err := w.Default(context.Background(), run)

	if err != nil {
		t.Fatalf("Default() error = %v", err)
	}
	if run.Spec.ToolchainRef != "" {
		t.Errorf("ToolchainRef = %q, want empty after snapshot", run.Spec.ToolchainRef)
	}
	if run.Spec.Toolchain == nil || run.Spec.Toolchain.Nix == nil {
		t.Fatal("Toolchain snapshot is nil")
	}
	if run.Spec.Toolchain.Nix.Image != toolchain.Spec.Nix.Image {
		t.Errorf("Toolchain image = %q, want %q", run.Spec.Toolchain.Nix.Image, toolchain.Spec.Nix.Image)
	}
}

func TestAgentRunWebhook_Default_SnapshotsReferencedHarnessAndProvider(t *testing.T) {
	runtimeClassName := "kata"
	harness := &agentv1alpha1.AgentHarness{
		ObjectMeta: metav1.ObjectMeta{Name: "harness", Namespace: "test"},
		Spec: agentv1alpha1.AgentSpec{
			Type:                    agentv1alpha1.AgentTypeClaude,
			DefaultProvider:         "provider",
			DefaultImage:            "ghcr.io/xonovex/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			DefaultRuntimeClassName: &runtimeClassName,
		},
	}
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "test"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef:          "glm",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{Name: "provider-auth", Key: "token"},
		},
	}
	w := &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(harness, provider).Build()}
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec:       agentv1alpha1.AgentRunSpec{HarnessRef: harness.Name},
	}

	err := w.Default(context.Background(), run)

	if err != nil {
		t.Fatalf("Default() error = %v", err)
	}
	if run.Spec.HarnessRef != "" || run.Spec.Harness == nil {
		t.Fatalf("HarnessRef = %q, Harness = %#v, want immutable inline snapshot", run.Spec.HarnessRef, run.Spec.Harness)
	}
	if run.Spec.ProviderRef != "" || run.Spec.Provider == nil || run.Spec.Provider.AuthSecretRef == nil {
		t.Fatalf("ProviderRef = %q, Provider = %#v, want immutable inline snapshot", run.Spec.ProviderRef, run.Spec.Provider)
	}
	if run.Spec.Provider.AuthSecretRef.Name != "provider-auth" {
		t.Fatalf("provider Secret = %q, want provider-auth", run.Spec.Provider.AuthSecretRef.Name)
	}
}

func TestAgentRunWebhook_Validate_RejectsUnapprovedSecretReferences(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*agentv1alpha1.AgentRun)
	}{
		{
			name: "environment",
			mutate: func(run *agentv1alpha1.AgentRun) {
				run.Spec.Env = []corev1.EnvVar{{
					Name: "DATABASE_TOKEN",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "database-admin"},
						Key:                  "token",
					}},
				}}
			},
		},
		{
			name: "provider",
			mutate: func(run *agentv1alpha1.AgentRun) {
				run.Spec.Provider = &agentv1alpha1.ProviderSpec{
					AuthSecretRef: &agentv1alpha1.SecretKeyRef{Name: "provider-admin", Key: "token"},
					AuthTokenEnv:  "ANTHROPIC_AUTH_TOKEN",
				}
			},
		},
		{
			name: "repository",
			mutate: func(run *agentv1alpha1.AgentRun) {
				run.Spec.Workspace.Repository.CredentialsSecretRef = &agentv1alpha1.SecretKeyRef{Name: "repository-admin", Key: "credentials"}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := baseRun()
			test.mutate(run)

			_, err := sandboxedAgentRunWebhook(run.Namespace).ValidateCreate(context.Background(), run)

			if err == nil || !strings.Contains(err.Error(), "not allowed") {
				t.Fatalf("ValidateCreate() error = %v, want Secret policy error", err)
			}
		})
	}
}

func TestAgentRunWebhook_Validate_AllowsPolicyApprovedSecret(t *testing.T) {
	policy := basePolicy()
	policy.ObjectMeta = metav1.ObjectMeta{Name: "sandbox-policy", Namespace: "test-ns"}
	policy.Spec.Enforced.AllowedSecretNames = []string{"provider-auth"}
	w := &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(policy).Build()}
	run := baseRun()
	run.Spec.Provider = &agentv1alpha1.ProviderSpec{
		AuthSecretRef: &agentv1alpha1.SecretKeyRef{Name: "provider-auth", Key: "token"},
		AuthTokenEnv:  "ANTHROPIC_AUTH_TOKEN",
	}

	_, err := w.ValidateCreate(context.Background(), run)

	if err != nil {
		t.Fatalf("ValidateCreate() error = %v, want approved Secret", err)
	}
}

func TestAgentRunWebhook_Validate_RejectsNonPositiveTimeout(t *testing.T) {
	run := baseRun()
	run.Spec.Timeout = &metav1.Duration{}

	_, err := sandboxedAgentRunWebhook(run.Namespace).ValidateCreate(context.Background(), run)

	if err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("ValidateCreate() error = %v, want positive-timeout error", err)
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
		{"valid packages", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, false},
		{"valid flake", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, FlakeRef: "/repo", Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, false},
		{"missing rev", &agentv1alpha1.NixSpec{Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"mutable rev", &agentv1alpha1.NixSpec{NixpkgsRev: "nixos-unstable", Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"packages and flake", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, Packages: []string{"ripgrep"}, FlakeRef: "/repo", Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"no source", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, Image: "ghcr.io/x/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, true},
		{"missing image", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, Packages: []string{"ripgrep"}}, true},
		{"moving image tag", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent:latest"}, true},
		{"malformed digest", &agentv1alpha1.NixSpec{NixpkgsRev: testNixRevision, Packages: []string{"ripgrep"}, Image: "ghcr.io/x/agent@sha256:not-a-digest"}, true},
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

func TestAgentRunWebhook_Validate_InvalidInlineWorkspaceStorageSize(t *testing.T) {
	run := baseRun()
	run.Spec.Workspace.StorageSize = "not-a-quantity"

	_, err := (&AgentRunWebhook{}).ValidateCreate(context.Background(), run)
	if err == nil || !strings.Contains(err.Error(), "storageSize") {
		t.Fatalf("ValidateCreate() error = %v, want storageSize validation error", err)
	}
}

func TestAgentRunWebhook_Validate_ProxyFailsClosed(t *testing.T) {
	run := baseRun()
	run.Spec.Network = agentv1alpha1.NetworkModeProxy

	_, err := (&AgentRunWebhook{}).ValidateCreate(context.Background(), run)
	if err == nil || !strings.Contains(err.Error(), "network=proxy") {
		t.Fatalf("ValidateCreate() error = %v, want unavailable proxy error", err)
	}
}

func TestAgentRunWebhook_Validate_ReferencedHarnessDefaultsCannotBypassPolicy(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*agentv1alpha1.AgentHarness)
		phrase string
	}{
		{
			name: "weak security context",
			mutate: func(harness *agentv1alpha1.AgentHarness) {
				harness.Spec.DefaultSecurityContext = &corev1.SecurityContext{
					AllowPrivilegeEscalation: boolPtr(true),
				}
			},
			phrase: "AllowPrivilegeEscalation",
		},
		{
			name: "custom open egress",
			mutate: func(harness *agentv1alpha1.AgentHarness) {
				harness.Spec.DefaultNetworkPolicy = &agentv1alpha1.AgentNetworkPolicy{
					Egress: []networkingv1.NetworkPolicyEgressRule{{}},
				}
			},
			phrase: "custom egress",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtimeClassName := "kata"
			harness := &agentv1alpha1.AgentHarness{
				ObjectMeta: metav1.ObjectMeta{Name: "secure-harness", Namespace: "test-ns"},
				Spec: agentv1alpha1.AgentSpec{
					Type:                    agentv1alpha1.AgentTypeClaude,
					DefaultImage:            baseRun().Spec.Image,
					DefaultRuntimeClassName: &runtimeClassName,
				},
			}
			test.mutate(harness)
			policy := basePolicy()
			policy.ObjectMeta = metav1.ObjectMeta{Name: "sandbox-policy", Namespace: "test-ns"}
			webhook := &AgentRunWebhook{Client: fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithObjects(harness, policy).
				Build()}
			run := baseRun()
			run.Spec.HarnessRef = harness.Name
			run.Spec.Image = ""
			run.Spec.RuntimeClassName = nil

			_, err := webhook.ValidateCreate(context.Background(), run)
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("ValidateCreate() error = %v, want phrase %q", err, test.phrase)
			}
		})
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
			Provider:    &agentv1alpha1.ProviderSpec{},
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

func TestAgentRunWebhook_RejectsUnknownInlineProviderPreset(t *testing.T) {
	w := sandboxedAgentRunWebhook()
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{PresetRef: "not-a-provider"},
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"},
			},
		},
	}

	if _, err := w.ValidateCreate(context.Background(), run); err == nil {
		t.Fatal("ValidateCreate() error = nil, want unknown-preset error")
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
