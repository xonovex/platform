package builder

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildInitContainers(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	containers := BuildInitContainers(run, "node:latest", agentv1alpha1.WorkspaceTypeGit, nil)

	if len(containers) != 1 {
		t.Fatalf("len(containers) = %d, want 1", len(containers))
	}

	c := containers[0]
	if c.Name != "git-clone" {
		t.Errorf("name = %q, want %q", c.Name, "git-clone")
	}
	if c.Image != "node:latest" {
		t.Errorf("image = %q, want %q", c.Image, "node:latest")
	}
	if len(c.Command) != 1 || c.Command[0] != "sh" {
		t.Errorf("command = %v, want [sh]", c.Command)
	}
	if len(c.VolumeMounts) != 1 || c.VolumeMounts[0].MountPath != "/workspace" {
		t.Errorf("volume mount path = %q, want %q", c.VolumeMounts[0].MountPath, "/workspace")
	}
}

func TestBuildCloneScript_Basic(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL: "https://github.com/example/repo.git",
	}

	script := buildCloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

	if !strings.Contains(script, "set -e") {
		t.Error("script missing 'set -e'")
	}
	if !strings.Contains(script, "git clone") {
		t.Error("script missing 'git clone'")
	}
	if !strings.Contains(script, "'https://github.com/example/repo.git'") {
		t.Error("script missing quoted repo URL")
	}
	if !strings.Contains(script, "--single-branch --depth 1") {
		t.Error("script missing shallow clone flags")
	}
	if strings.Contains(script, "--branch") {
		t.Error("script should not have --branch when branch is empty")
	}
}

func TestBuildCloneScript_WithBranch(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:    "https://github.com/example/repo.git",
		Branch: "develop",
	}

	script := buildCloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

	if !strings.Contains(script, "--branch 'develop'") {
		t.Errorf("script missing '--branch 'develop'', got:\n%s", script)
	}
}

func TestBuildCloneScript_InjectionQuoted(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:    "https://github.com/example/repo.git",
		Branch: "main; rm -rf /",
		Commit: "abc1234",
	}

	script := buildCloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

	// The branch should be single-quoted so the semicolon is not interpreted
	if !strings.Contains(script, "'main; rm -rf /'") {
		t.Errorf("branch not properly quoted in script:\n%s", script)
	}
	// URL should be quoted with -- separator
	if !strings.Contains(script, "-- 'https://github.com/example/repo.git'") {
		t.Errorf("URL not properly quoted in script:\n%s", script)
	}
	// Commit should be quoted
	if !strings.Contains(script, "git fetch origin 'abc1234'") {
		t.Errorf("commit not properly quoted in script:\n%s", script)
	}
}

func TestBuildCloneScript_WithCommit(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:    "https://github.com/example/repo.git",
		Commit: "abc123",
	}

	script := buildCloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

	if !strings.Contains(script, "git fetch origin 'abc123'") {
		t.Errorf("script missing quoted commit in fetch, got:\n%s", script)
	}
	if !strings.Contains(script, "git checkout 'abc123'") {
		t.Errorf("script missing quoted commit in checkout, got:\n%s", script)
	}
}

func TestBuildCloneScript_WithJujutsu(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:    "https://github.com/example/repo.git",
		Branch: "main",
	}

	script := buildCloneScript(repo, agentv1alpha1.WorkspaceTypeJujutsu)

	if !strings.Contains(script, "git clone") {
		t.Error("jj script should still use git clone")
	}
	if !strings.Contains(script, "jj git init --colocate") {
		t.Error("jj script missing 'jj git init --colocate'")
	}
}

func TestBuildCloneScript_GitNoJJInit(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL: "https://github.com/example/repo.git",
	}

	script := buildCloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

	if strings.Contains(script, "jj") {
		t.Error("git-only script should not contain 'jj'")
	}
}

func TestResolveToolchain(t *testing.T) {
	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Image: "ghcr.io/xonovex/agent@sha256:abc"},
	}
	tcl := ResolveToolchain(tc)
	if tcl == nil || tcl.Image() != "ghcr.io/xonovex/agent@sha256:abc" {
		t.Fatalf("ResolveToolchain(nix).Image() = %v, want the pre-built image", tcl)
	}
	if !tcl.Pinned() {
		t.Error("nix toolchain must report Pinned()=true")
	}
	if ResolveToolchain(nil) != nil {
		t.Error("ResolveToolchain(nil) must be nil")
	}
	if ResolveToolchain(&agentv1alpha1.ToolchainSpec{Type: "bogus"}) != nil {
		t.Error("ResolveToolchain(unknown type) must be nil")
	}
}

func TestBuildMainContainers_WithoutNix(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil)
	c := containers[0]

	for _, vm := range c.VolumeMounts {
		if vm.Name == "nix-env" {
			t.Error("unexpected nix-env volume mount when Nix is not configured")
		}
	}
	for _, env := range c.Env {
		if env.Name == "PATH" {
			t.Error("unexpected PATH env var when Nix is not configured")
		}
	}
}

func TestBuildMainContainers_Claude(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil)

	if len(containers) != 1 {
		t.Fatalf("len(containers) = %d, want 1", len(containers))
	}

	c := containers[0]
	if c.Name != "agent" {
		t.Errorf("name = %q, want %q", c.Name, "agent")
	}
	if c.Command[0] != "claude" {
		t.Errorf("command = %v, want [claude]", c.Command)
	}
	if c.WorkingDir != "/workspace" {
		t.Errorf("workingDir = %q, want %q", c.WorkingDir, "/workspace")
	}

	// Claude should have --permission-mode bypassPermissions
	foundPermFlag := false
	for i, arg := range c.Args {
		if arg == "--permission-mode" && i+1 < len(c.Args) && c.Args[i+1] == "bypassPermissions" {
			foundPermFlag = true
			break
		}
	}
	if !foundPermFlag {
		t.Errorf("args missing '--permission-mode bypassPermissions', got %v", c.Args)
	}
}

func TestBuildMainContainers_ClaudeWithPrompt(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
			Prompt: "Fix the tests",
		},
	}

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil)

	args := containers[0].Args
	foundPrint := false
	foundPrompt := false
	for i, arg := range args {
		if arg == "--print" {
			foundPrint = true
		}
		if arg == "--prompt" && i+1 < len(args) && args[i+1] == "Fix the tests" {
			foundPrompt = true
		}
	}
	if !foundPrint {
		t.Errorf("args missing '--print', got %v", args)
	}
	if !foundPrompt {
		t.Errorf("args missing '--prompt Fix the tests', got %v", args)
	}
}

func TestBuildMainContainers_Opencode(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
			Provider: &agentv1alpha1.ProviderSpec{
				CliArgs: []string{"--model", "google/gemini-2.5-pro"},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeOpencode, nil)

	c := containers[0]
	if c.Command[0] != "opencode" {
		t.Errorf("command = %v, want [opencode]", c.Command)
	}
	if len(c.Args) < 2 || c.Args[0] != "--model" || c.Args[1] != "google/gemini-2.5-pro" {
		t.Errorf("args = %v, want [--model google/gemini-2.5-pro]", c.Args)
	}
}

func TestBuildMainContainers_WithProviderEnv(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	providerEnv := map[string]string{
		"ANTHROPIC_BASE_URL": "http://proxy:8080",
		"API_TIMEOUT_MS":     "60000",
	}

	containers := BuildMainContainers(run, providerEnv, "image", agentv1alpha1.AgentTypeClaude, nil)

	envMap := make(map[string]string)
	for _, env := range containers[0].Env {
		envMap[env.Name] = env.Value
	}

	if envMap["ANTHROPIC_BASE_URL"] != "http://proxy:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", envMap["ANTHROPIC_BASE_URL"], "http://proxy:8080")
	}
	if envMap["API_TIMEOUT_MS"] != "60000" {
		t.Errorf("API_TIMEOUT_MS = %q, want %q", envMap["API_TIMEOUT_MS"], "60000")
	}
}

func TestBuildMainContainers_SecurityContextDefaults(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil)
	sc := containers[0].SecurityContext

	if sc == nil {
		t.Fatal("SecurityContext should not be nil")
	}
	if *sc.AllowPrivilegeEscalation != false {
		t.Error("AllowPrivilegeEscalation should be false")
	}
	if *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot should be true")
	}
	if *sc.ReadOnlyRootFilesystem != true {
		t.Error("ReadOnlyRootFilesystem should be true")
	}
	if sc.Capabilities == nil || len(sc.Capabilities.Drop) != 1 || sc.Capabilities.Drop[0] != "ALL" {
		t.Error("Capabilities.Drop should be [ALL]")
	}
}

func TestBuildMainContainers_SecurityContextOverride(t *testing.T) {
	allowPrivEsc := true
	override := &corev1.SecurityContext{
		AllowPrivilegeEscalation: &allowPrivEsc,
	}

	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, override)
	sc := containers[0].SecurityContext

	if *sc.AllowPrivilegeEscalation != true {
		t.Error("AllowPrivilegeEscalation override should be true")
	}
	if *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot default should be preserved")
	}
}

func TestBuildMainContainers_TmpVolumeMount(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil)
	foundTmp := false
	for _, vm := range containers[0].VolumeMounts {
		if vm.Name == "tmp" && vm.MountPath == "/tmp" {
			foundTmp = true
		}
	}
	if !foundTmp {
		t.Error("expected /tmp volume mount for ReadOnlyRootFilesystem")
	}
}

func TestBuildInitContainers_SecurityContextDefaults(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	containers := BuildInitContainers(run, "image", agentv1alpha1.WorkspaceTypeGit, nil)
	sc := containers[0].SecurityContext

	if sc == nil {
		t.Fatal("init container SecurityContext should not be nil")
	}
	if *sc.AllowPrivilegeEscalation != false {
		t.Error("AllowPrivilegeEscalation should be false")
	}
	if *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot should be true")
	}
}
