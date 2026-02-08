package builder

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildInitContainers(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	containers := BuildInitContainers(run, "node:latest")

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
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	script := buildCloneScript(run)

	if !strings.Contains(script, "set -e") {
		t.Error("script missing 'set -e'")
	}
	if !strings.Contains(script, "git clone") {
		t.Error("script missing 'git clone'")
	}
	if !strings.Contains(script, "https://github.com/example/repo.git") {
		t.Error("script missing repo URL")
	}
	if !strings.Contains(script, "--single-branch --depth 1") {
		t.Error("script missing shallow clone flags")
	}
	if strings.Contains(script, "--branch") {
		t.Error("script should not have --branch when branch is empty")
	}
}

func TestBuildCloneScript_WithBranch(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL:    "https://github.com/example/repo.git",
				Branch: "develop",
			},
		},
	}

	script := buildCloneScript(run)

	if !strings.Contains(script, "--branch develop") {
		t.Error("script missing '--branch develop'")
	}
}

func TestBuildCloneScript_WithCommit(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL:    "https://github.com/example/repo.git",
				Commit: "abc123",
			},
		},
	}

	script := buildCloneScript(run)

	if !strings.Contains(script, "git fetch origin abc123") {
		t.Error("script missing 'git fetch origin abc123'")
	}
	if !strings.Contains(script, "git checkout abc123") {
		t.Error("script missing 'git checkout abc123'")
	}
}

func TestBuildCloneScript_WithWorktree(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
			Worktree: &agentv1alpha1.WorktreeSpec{
				Branch:       "feature-branch",
				SourceBranch: "main",
			},
		},
	}

	script := buildCloneScript(run)

	if !strings.Contains(script, "git worktree add /workspace-wt -b feature-branch main") {
		t.Errorf("script missing worktree command, got:\n%s", script)
	}
}

func TestBuildCloneScript_WorktreeDefaultSource(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
			Worktree: &agentv1alpha1.WorktreeSpec{
				Branch: "my-branch",
			},
		},
	}

	script := buildCloneScript(run)

	if !strings.Contains(script, "git worktree add /workspace-wt -b my-branch HEAD") {
		t.Errorf("script should default source to HEAD, got:\n%s", script)
	}
}

func TestBuildMainContainers_Claude(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image:latest")

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
			Agent: agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
			Prompt: "Fix the tests",
		},
	}

	containers := BuildMainContainers(run, nil, "image")

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
			Agent: agentv1alpha1.AgentTypeOpencode,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
			Provider: &agentv1alpha1.ProviderSpec{
				CliArgs: []string{"--model", "google/gemini-2.5-pro"},
			},
		},
	}

	containers := BuildMainContainers(run, nil, "image")

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
			Agent: agentv1alpha1.AgentTypeClaude,
			Repository: agentv1alpha1.RepositorySpec{
				URL: "https://github.com/example/repo.git",
			},
		},
	}

	providerEnv := map[string]string{
		"ANTHROPIC_BASE_URL": "http://proxy:8080",
		"API_TIMEOUT_MS":     "60000",
	}

	containers := BuildMainContainers(run, providerEnv, "image")

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
