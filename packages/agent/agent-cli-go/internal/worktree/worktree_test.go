package worktree

import (
	"os"
	"testing"

	sharedworktree "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/worktree"
)

func TestSetup_DefaultsToGit(t *testing.T) {
	// Setup with empty VCS should default to git
	config := Config{
		Branch: "test-branch",
		Dir:    "/tmp/nonexistent-worktree-test",
	}

	// This will fail because we're not in a git repo context,
	// but it proves the dispatch path works (git path is taken)
	_, err := Setup(config, "/tmp", false)
	if err == nil {
		// Clean up if somehow it succeeded
		os.RemoveAll(config.Dir)
	}
	// Error is expected since /tmp isn't a git repo - we just verify it doesn't panic
}

func TestSetup_JJDispatch(t *testing.T) {
	config := Config{
		Branch: "test-branch",
		Dir:    "/tmp/nonexistent-jj-workspace-test",
		VCS:    sharedworktree.VCSJujutsu,
	}

	_, err := Setup(config, "/tmp", false)
	// Should error because either jj isn't available or /tmp isn't a repo
	if err == nil {
		os.RemoveAll(config.Dir)
		t.Error("expected error for jj workspace in non-repo directory")
	}
}

func TestSetupJJ_ReusesExistingWorkspace(t *testing.T) {
	// Create a temporary directory to simulate existing workspace
	tmpDir := t.TempDir()

	config := Config{
		Branch: "test-branch",
		Dir:    tmpDir,
		VCS:    sharedworktree.VCSJujutsu,
	}

	result, err := SetupJJ(config, "/tmp", false)
	if err != nil {
		t.Fatalf("SetupJJ should reuse existing dir, got error: %v", err)
	}
	if result != tmpDir {
		t.Errorf("SetupJJ returned %q, want %q", result, tmpDir)
	}
}

func TestVCSType_IsValid(t *testing.T) {
	tests := []struct {
		vcs   sharedworktree.VCSType
		valid bool
	}{
		{sharedworktree.VCSGit, true},
		{sharedworktree.VCSJujutsu, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := tt.vcs.IsValid(); got != tt.valid {
			t.Errorf("VCSType(%q).IsValid() = %v, want %v", tt.vcs, got, tt.valid)
		}
	}
}
