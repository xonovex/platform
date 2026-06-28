package jj

import (
	"testing"

	wsshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/shared"
)

func TestSetup_ReusesExistingWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := New().Setup(wsshared.Config{Branch: "test-branch", Dir: tmpDir}, "/tmp", false)
	if err != nil {
		t.Fatalf("Setup should reuse existing dir, got error: %v", err)
	}
	if result != tmpDir {
		t.Errorf("Setup returned %q, want %q", result, tmpDir)
	}
}

func TestSetup_NonRepoErrors(t *testing.T) {
	// A fresh (non-existent) dir under a non-repo errors: either jj is absent or
	// /tmp is not a repo.
	_, err := New().Setup(wsshared.Config{Branch: "test-branch", Dir: "/tmp/nonexistent-jj-workspace-test"}, "/tmp", false)
	if err == nil {
		t.Error("expected error for jj workspace in non-repo directory")
	}
}
