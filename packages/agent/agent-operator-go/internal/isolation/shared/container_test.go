package shared

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func workspaceRun() *agentv1alpha1.AgentRun {
	return &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"},
			},
		},
	}
}

func TestBuildInitContainers(t *testing.T) {
	containers := BuildInitContainers(workspaceRun(), "node:latest", agentv1alpha1.WorkspaceTypeGit, nil)

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

func TestBuildInitContainers_SecurityContextDefaults(t *testing.T) {
	containers := BuildInitContainers(workspaceRun(), "image", agentv1alpha1.WorkspaceTypeGit, nil)
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

func TestBuildMainContainers_WithoutNix(t *testing.T) {
	containers := BuildMainContainers(workspaceRun(), nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil)
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
	run := workspaceRun()
	run.ObjectMeta = metav1.ObjectMeta{Name: "test"}
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
	run := workspaceRun()
	run.Spec.Prompt = "Fix the tests"
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
	run := workspaceRun()
	run.Spec.Provider = &agentv1alpha1.ProviderSpec{CliArgs: []string{"--model", "google/gemini-2.5-pro"}}
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
	providerEnv := map[string]string{
		"ANTHROPIC_BASE_URL": "http://proxy:8080",
		"API_TIMEOUT_MS":     "60000",
	}
	containers := BuildMainContainers(workspaceRun(), providerEnv, "image", agentv1alpha1.AgentTypeClaude, nil)

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
	containers := BuildMainContainers(workspaceRun(), nil, "image", agentv1alpha1.AgentTypeClaude, nil)
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
	override := &corev1.SecurityContext{AllowPrivilegeEscalation: &allowPrivEsc}
	containers := BuildMainContainers(workspaceRun(), nil, "image", agentv1alpha1.AgentTypeClaude, override)
	sc := containers[0].SecurityContext
	if *sc.AllowPrivilegeEscalation != true {
		t.Error("AllowPrivilegeEscalation override should be true")
	}
	if *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot default should be preserved")
	}
}

func TestBuildMainContainers_TmpVolumeMount(t *testing.T) {
	containers := BuildMainContainers(workspaceRun(), nil, "image", agentv1alpha1.AgentTypeClaude, nil)
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
