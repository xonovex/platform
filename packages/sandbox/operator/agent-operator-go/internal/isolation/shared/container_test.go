package shared

import (
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	sharedlaunch "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
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
	containers := mustBuildInitContainers(t, workspaceRun(), "node:latest", agentv1alpha1.WorkspaceTypeGit, nil)

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
	containers := mustBuildInitContainers(t, workspaceRun(), "image", agentv1alpha1.WorkspaceTypeGit, nil)
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
	containers := mustBuildMainContainers(t, workspaceRun(), nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil)
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

func TestBuildMainContainers_RunsThePlannedProcess(t *testing.T) {
	run := workspaceRun()
	run.ObjectMeta = metav1.ObjectMeta{Name: "test"}
	run.Spec.Prompt = "Fix the tests"
	containers := mustBuildMainContainers(t, run, nil, "image:latest", agentv1alpha1.AgentTypeClaude, nil)

	if len(containers) != 1 {
		t.Fatalf("len(containers) = %d, want 1", len(containers))
	}
	c := containers[0]
	if c.Name != "agent" {
		t.Errorf("name = %q, want %q", c.Name, "agent")
	}
	if !slices.Equal(c.Command, []string{"claude"}) {
		t.Errorf("command = %v, want the harness binary [claude]", c.Command)
	}
	if want := []string{"--permission-mode", "bypassPermissions", "-p", "Fix the tests"}; !slices.Equal(c.Args, want) {
		t.Errorf("args = %v, want %v", c.Args, want)
	}
	if c.Image != "image:latest" {
		t.Errorf("image = %q, want the planned image", c.Image)
	}
	if c.WorkingDir != "/workspace" {
		t.Errorf("workingDir = %q, want %q", c.WorkingDir, "/workspace")
	}
}

func TestBuildMainContainers_SplitsTheArgumentVectorAndKeepsTheProcessEnvironment(t *testing.T) {
	spec := sharedlaunch.Spec{Run: isoshared.RunConfig{Image: "image", Process: isoshared.Process{
		Argv: []string{"opencode", "--model", "google/gemini-2.5-pro"},
		Env:  []string{"PLANNED=1"},
		Dir:  "/workspace",
	}}}
	run := workspaceRun()
	run.Spec.Env = []corev1.EnvVar{{Name: "RUN", Value: "2"}}

	containers, err := buildMainContainers(run, []corev1.EnvVar{{Name: "PROVIDER", Value: "0"}}, spec, nil)

	if err != nil {
		t.Fatalf("buildMainContainers() error = %v", err)
	}
	c := containers[0]
	if !slices.Equal(c.Command, []string{"opencode"}) || !slices.Equal(c.Args, []string{"--model", "google/gemini-2.5-pro"}) {
		t.Errorf("command, args = %v, %v, want [opencode] [--model google/gemini-2.5-pro]", c.Command, c.Args)
	}
	var names []string
	for _, env := range c.Env {
		names = append(names, env.Name+"="+env.Value)
	}
	if want := []string{"PROVIDER=0", "PLANNED=1", "RUN=2"}; !slices.Equal(names, want) {
		t.Errorf("env = %v, want %v: provider, planned process, then the run", names, want)
	}
}

func TestBuildMainContainers_ABareBinaryHasNoArgs(t *testing.T) {
	spec := sharedlaunch.Spec{Run: isoshared.RunConfig{Process: isoshared.Process{Argv: []string{"copilot"}}}}

	containers, err := buildWorkspaceMainContainers(workspaceRun(), nil, spec, nil, nil, nil)

	if err != nil || containers[0].Args != nil {
		t.Fatalf("args = %v, %v, want none", containers[0].Args, err)
	}
}

func TestBuildMainContainers_WithProviderEnv(t *testing.T) {
	providerEnv := []corev1.EnvVar{
		{Name: "ANTHROPIC_BASE_URL", Value: "http://proxy:8080"},
		{Name: "API_TIMEOUT_MS", Value: "60000"},
	}
	containers := mustBuildMainContainers(t, workspaceRun(), providerEnv, "image", agentv1alpha1.AgentTypeClaude, nil)

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
	containers := mustBuildMainContainers(t, workspaceRun(), nil, "image", agentv1alpha1.AgentTypeClaude, nil)
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
	containers := mustBuildMainContainers(t, workspaceRun(), nil, "image", agentv1alpha1.AgentTypeClaude, override)
	sc := containers[0].SecurityContext
	if *sc.AllowPrivilegeEscalation != true {
		t.Error("AllowPrivilegeEscalation override should be true")
	}
	if *sc.RunAsNonRoot != true {
		t.Error("RunAsNonRoot default should be preserved")
	}
}

func TestBuildMainContainers_TmpVolumeMount(t *testing.T) {
	containers := mustBuildMainContainers(t, workspaceRun(), nil, "image", agentv1alpha1.AgentTypeClaude, nil)
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

func TestBuildMainContainers_RefusesAPlanWithNoProcess(t *testing.T) {
	if _, err := buildMainContainers(workspaceRun(), nil, sharedlaunch.Spec{}, nil); err == nil {
		t.Fatal("expected a plan with no process to return an error")
	}
	if _, err := buildWorkspaceMainContainers(workspaceRun(), nil, sharedlaunch.Spec{}, nil, nil, nil); err == nil {
		t.Fatal("expected a plan with no process to return an error")
	}
}
