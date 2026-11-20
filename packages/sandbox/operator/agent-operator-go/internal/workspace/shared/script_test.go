package shared

import (
	"strings"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// stubStrategy is an in-memory VCSStrategy whose post-clone step is fixed.
type stubStrategy struct{ postClone string }

func (s stubStrategy) PostCloneScript() string { return s.postClone }

func (stubStrategy) WorktreeScript(path, branch, sourceBranch string) string {
	return "worktree " + path + " " + branch + " " + sourceBranch + "\n"
}

func (stubStrategy) InitContainerName() string { return "stub-worktree" }

func TestCloneScript_Basic(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"}
	script := CloneScript(repo, stubStrategy{})

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

func TestCloneScript_WithBranch(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git", Branch: "develop"}
	script := CloneScript(repo, stubStrategy{})
	if !strings.Contains(script, "--branch 'develop'") {
		t.Errorf("script missing '--branch 'develop'', got:\n%s", script)
	}
}

func TestCloneScript_UsesMountedCredentialStore(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:                  "https://github.com/example/private.git",
		CredentialsSecretRef: &agentv1alpha1.SecretKeyRef{Name: "repo-auth", Key: "credentials"},
	}

	script := CloneScript(repo, stubStrategy{})

	if !strings.Contains(script, "credential.helper=store --file="+RepositoryCredentialsFile) {
		t.Fatalf("clone script does not use the mounted credential store:\n%s", script)
	}
	if strings.Contains(script, "repo-auth") {
		t.Fatal("clone script must not expose the Kubernetes Secret name")
	}
}

func TestCloneScript_InjectionQuoted(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:    "https://github.com/example/repo.git",
		Branch: "main; rm -rf /",
		Commit: "abc1234",
	}
	script := CloneScript(repo, stubStrategy{})

	if !strings.Contains(script, "'main; rm -rf /'") {
		t.Errorf("branch not properly quoted in script:\n%s", script)
	}
	if !strings.Contains(script, "-- 'https://github.com/example/repo.git'") {
		t.Errorf("URL not properly quoted in script:\n%s", script)
	}
	if !strings.Contains(script, "git fetch origin 'abc1234'") {
		t.Errorf("commit not properly quoted in script:\n%s", script)
	}
}

func TestCloneScript_WithCommit(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git", Commit: "abc123"}
	script := CloneScript(repo, stubStrategy{})
	if !strings.Contains(script, "git fetch origin 'abc123'") {
		t.Errorf("script missing quoted commit in fetch, got:\n%s", script)
	}
	if !strings.Contains(script, "git checkout 'abc123'") {
		t.Errorf("script missing quoted commit in checkout, got:\n%s", script)
	}
}

func TestCloneScript_AppendsThePostCloneStep(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git", Branch: "main"}
	script := CloneScript(repo, stubStrategy{postClone: "post-clone-step\n"})
	if !strings.Contains(script, "git clone") {
		t.Error("script should clone with git before the post-clone step")
	}
	if !strings.HasSuffix(script, "post-clone-step\n") {
		t.Errorf("script does not end with the post-clone step:\n%s", script)
	}
}

func TestCloneScript_WithoutAStrategyAddsNoPostCloneStep(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"}
	if got, want := CloneScript(repo, nil), CloneScript(repo, stubStrategy{}); got != want {
		t.Errorf("CloneScript(nil) = %q, want %q", got, want)
	}
}

func TestWorktreeScriptAndName_UsesTheStrategy(t *testing.T) {
	script, name := WorktreeScriptAndName(stubStrategy{}, WorktreePath("run"), "feature", "")

	if name != "stub-worktree" {
		t.Errorf("name = %q, want stub-worktree", name)
	}
	if !strings.HasSuffix(script, "worktree /workspace-wt/run feature HEAD\n") {
		t.Errorf("script does not default the source branch to HEAD:\n%s", script)
	}
}

func TestWorktreeScriptAndName_DefaultsWithoutAStrategy(t *testing.T) {
	script, name := WorktreeScriptAndName(nil, WorktreePath("run"), "feature", "main")

	if name != "git-worktree" {
		t.Errorf("name = %q, want git-worktree", name)
	}
	if script != "set -e\ncd /workspace\n" {
		t.Errorf("script = %q, want only the preamble", script)
	}
}

func TestWorkDirIsTheWorktreeOfASharedWorkspace(t *testing.T) {
	standalone := &agentv1alpha1.AgentRun{}
	standalone.Name = "run"
	if got := WorkDir(standalone); got != WorkspaceMountPath {
		t.Errorf("WorkDir(standalone) = %q, want %q", got, WorkspaceMountPath)
	}
	shared := standalone.DeepCopy()
	shared.Spec.WorkspaceRef = "workspace"
	if got := WorkDir(shared); got != WorktreeBasePath+"/run" {
		t.Errorf("WorkDir(shared) = %q, want the worktree of the run", got)
	}
}
