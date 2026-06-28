package git

import (
	"os"
	"testing"

	wsshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/workspace/shared"
)

func TestSetup_NonRepoDoesNotPanic(t *testing.T) {
	config := wsshared.Config{Branch: "test-branch", Dir: "/tmp/nonexistent-worktree-test"}
	// /tmp is not a git repo, so this errors; we just verify the git path runs
	// without panicking.
	if _, err := New().Setup(config, "/tmp", false); err == nil {
		os.RemoveAll(config.Dir)
	}
}

func TestAvailable(t *testing.T) {
	if !New().Available() {
		t.Error("git worktree variant must always report available")
	}
}
