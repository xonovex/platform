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

	containers := BuildInitContainers(run, "node:latest", agentv1alpha1.WorkspaceTypeGit, nil, nil)

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

func TestNixToolchain_InitContainer_WithPackages(t *testing.T) {
	nix := &agentv1alpha1.NixSpec{
		Packages: []string{"nodejs_22", "python3", "ripgrep"},
	}

	tc := NewNixToolchain(nix)
	c := tc.InitContainer()
	if c == nil {
		t.Fatal("expected non-nil container")
	}

	if c.Name != "nix-env" {
		t.Errorf("name = %q, want %q", c.Name, "nix-env")
	}
	if c.Image != "nixos/nix:latest" {
		t.Errorf("image = %q, want %q", c.Image, "nixos/nix:latest")
	}

	script := c.Args[1]
	if !strings.Contains(script, "cp -a /nix/. /nix-env/") {
		t.Error("script missing Nix store bootstrap")
	}
	if !strings.Contains(script, "nixpkgs#nodejs_22") {
		t.Error("script missing nodejs_22 package")
	}
	if !strings.Contains(script, "nixpkgs#python3") {
		t.Error("script missing python3 package")
	}
	if !strings.Contains(script, "nixpkgs#ripgrep") {
		t.Error("script missing ripgrep package")
	}
	if !strings.Contains(script, "profile install --profile /nix/var/nix/profiles/agent") {
		t.Error("script missing profile install command")
	}

	if len(c.VolumeMounts) != 1 || c.VolumeMounts[0].MountPath != "/nix-env" {
		t.Errorf("volume mount = %v, want /nix-env", c.VolumeMounts)
	}
}

func TestNixToolchain_InitContainer_CustomImage(t *testing.T) {
	nix := &agentv1alpha1.NixSpec{
		Packages: []string{"nodejs_22"},
		Image:    "nixos/nix:2.28.3",
	}

	tc := NewNixToolchain(nix)
	c := tc.InitContainer()
	if c.Image != "nixos/nix:2.28.3" {
		t.Errorf("image = %q, want %q", c.Image, "nixos/nix:2.28.3")
	}
}

func TestToolchains_Nil(t *testing.T) {
	tcs := Toolchains(nil)
	if len(tcs) != 0 {
		t.Errorf("len(toolchains) = %d, want 0", len(tcs))
	}
}

func TestToolchains_EmptyPackages(t *testing.T) {
	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{},
	}
	tcs := Toolchains(tc)
	if len(tcs) != 0 {
		t.Errorf("len(toolchains) = %d, want 0", len(tcs))
	}
}

func TestBuildInitContainers_WithNix(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix: &agentv1alpha1.NixSpec{
			Packages: []string{"nodejs_22"},
		},
	}

	containers := BuildInitContainers(run, "node:latest", agentv1alpha1.WorkspaceTypeGit, tc, nil)

	if len(containers) != 2 {
		t.Fatalf("len(containers) = %d, want 2", len(containers))
	}
	if containers[0].Name != "git-clone" {
		t.Errorf("containers[0].Name = %q, want git-clone", containers[0].Name)
	}
	if containers[1].Name != "nix-env" {
		t.Errorf("containers[1].Name = %q, want nix-env", containers[1].Name)
	}
}

func TestBuildMainContainers_WithNix(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix: &agentv1alpha1.NixSpec{
			Packages: []string{"nodejs_22"},
		},
	}

	containers := BuildMainContainers(run, nil, "image:latest", agentv1alpha1.AgentTypeClaude, tc, nil)

	c := containers[0]

	// Should have nix-env volume mount
	foundNix := false
	for _, vm := range c.VolumeMounts {
		if vm.Name == "nix-env" && vm.MountPath == "/nix" {
			foundNix = true
		}
	}
	if !foundNix {
		t.Error("expected nix-env volume mount at /nix")
	}

	// Should have PATH with nix profile bin
	foundPath := false
	for _, env := range c.Env {
		if env.Name == "PATH" && strings.Contains(env.Value, "/nix/var/nix/profiles/agent/bin") {
			foundPath = true
		}
	}
	if !foundPath {
		t.Error("expected PATH env var with nix profile bin")
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

	containers := BuildMainContainers(run, nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil, nil)
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

	containers := BuildMainContainers(run, nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil, nil)

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

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil, nil)

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

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeOpencode, nil, nil)

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

	containers := BuildMainContainers(run, providerEnv, "image", agentv1alpha1.AgentTypeClaude, nil, nil)

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

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil, nil)
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

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil, override)
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

	containers := BuildMainContainers(run, nil, "image", agentv1alpha1.AgentTypeClaude, nil, nil)
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

	containers := BuildInitContainers(run, "image", agentv1alpha1.WorkspaceTypeGit, nil, nil)
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

func TestBuildInitContainers_NixSecurityContext(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://example.com/repo.git"},
			},
		},
	}

	tc := &agentv1alpha1.ToolchainSpec{
		Type: agentv1alpha1.ToolchainTypeNix,
		Nix:  &agentv1alpha1.NixSpec{Packages: []string{"nodejs_22"}},
	}

	containers := BuildInitContainers(run, "image", agentv1alpha1.WorkspaceTypeGit, tc, nil)

	if len(containers) != 2 {
		t.Fatalf("len(containers) = %d, want 2", len(containers))
	}

	nixSC := containers[1].SecurityContext
	if nixSC == nil {
		t.Fatal("nix init container SecurityContext should not be nil")
	}
	if *nixSC.AllowPrivilegeEscalation != false {
		t.Error("nix init container AllowPrivilegeEscalation should be false")
	}
}

func TestNixToolchain_Volumes_DefaultSizeLimit(t *testing.T) {
	nix := &agentv1alpha1.NixSpec{
		Packages: []string{"nodejs_22"},
	}

	tc := NewNixToolchain(nix)
	volumes := tc.Volumes()

	if len(volumes) != 1 {
		t.Fatalf("len(volumes) = %d, want 1", len(volumes))
	}
	vol := volumes[0]
	if vol.EmptyDir == nil {
		t.Fatal("expected EmptyDir volume source")
	}
	if vol.EmptyDir.SizeLimit == nil {
		t.Fatal("expected SizeLimit to be set")
	}
	expected := "10Gi"
	if vol.EmptyDir.SizeLimit.String() != expected {
		t.Errorf("SizeLimit = %q, want %q", vol.EmptyDir.SizeLimit.String(), expected)
	}
}

func TestNixToolchain_Volumes_CustomSizeLimit(t *testing.T) {
	nix := &agentv1alpha1.NixSpec{
		Packages:       []string{"nodejs_22"},
		StoreSizeLimit: "20Gi",
	}

	tc := NewNixToolchain(nix)
	volumes := tc.Volumes()

	if len(volumes) != 1 {
		t.Fatalf("len(volumes) = %d, want 1", len(volumes))
	}
	vol := volumes[0]
	if vol.EmptyDir == nil || vol.EmptyDir.SizeLimit == nil {
		t.Fatal("expected EmptyDir with SizeLimit")
	}
	expected := "20Gi"
	if vol.EmptyDir.SizeLimit.String() != expected {
		t.Errorf("SizeLimit = %q, want %q", vol.EmptyDir.SizeLimit.String(), expected)
	}
}
