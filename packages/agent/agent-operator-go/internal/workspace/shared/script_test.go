package shared

import (
	"strings"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestCloneScript_Basic(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"}
	script := CloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

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
	script := CloneScript(repo, agentv1alpha1.WorkspaceTypeGit)
	if !strings.Contains(script, "--branch 'develop'") {
		t.Errorf("script missing '--branch 'develop'', got:\n%s", script)
	}
}

func TestCloneScript_InjectionQuoted(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{
		URL:    "https://github.com/example/repo.git",
		Branch: "main; rm -rf /",
		Commit: "abc1234",
	}
	script := CloneScript(repo, agentv1alpha1.WorkspaceTypeGit)

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
	script := CloneScript(repo, agentv1alpha1.WorkspaceTypeGit)
	if !strings.Contains(script, "git fetch origin 'abc123'") {
		t.Errorf("script missing quoted commit in fetch, got:\n%s", script)
	}
	if !strings.Contains(script, "git checkout 'abc123'") {
		t.Errorf("script missing quoted commit in checkout, got:\n%s", script)
	}
}

func TestCloneScript_WithJujutsu(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git", Branch: "main"}
	script := CloneScript(repo, agentv1alpha1.WorkspaceTypeJujutsu)
	if !strings.Contains(script, "git clone") {
		t.Error("jj script should still use git clone")
	}
	if !strings.Contains(script, "jj git init --colocate") {
		t.Error("jj script missing 'jj git init --colocate'")
	}
}

func TestCloneScript_GitNoJJInit(t *testing.T) {
	repo := agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"}
	script := CloneScript(repo, agentv1alpha1.WorkspaceTypeGit)
	if strings.Contains(script, "jj") {
		t.Error("git-only script should not contain 'jj'")
	}
}
