package jj

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	wsshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/workspace/shared"
)

func installFakeJJ(t *testing.T, body string) {
	t.Helper()

	binDir := t.TempDir()
	path := filepath.Join(binDir, "jj")
	script := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake jj: %v", err)
	}
	t.Setenv("PATH", binDir)
}

func TestSetup_ReusesExistingWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := New().Setup(wsshared.Config{Branch: "test-branch", Dir: tmpDir}, t.TempDir(), true)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if result != tmpDir {
		t.Errorf("Setup() = %q, want %q", result, tmpDir)
	}
}

func TestSetup_CreatesWorkspace(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "jj.log")
	t.Setenv("JJ_TEST_LOG", logPath)
	installFakeJJ(t, `printf '%s\n' "$*" > "$JJ_TEST_LOG"
/bin/mkdir -p "$3"`)
	repoDir := t.TempDir()
	target := "workspace"

	result, err := New().Setup(wsshared.Config{SourceBranch: "main", Dir: target}, repoDir, true)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	want := filepath.Join(repoDir, target)
	if result != want {
		t.Errorf("Setup() = %q, want %q", result, want)
	}
	log, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read fake jj log: %v", err)
	}
	if got := strings.TrimSpace(string(log)); got != "workspace add "+want+" --revision main" {
		t.Errorf("jj arguments = %q", got)
	}
}

func TestSetup_ReportsWorkspaceCommandFailure(t *testing.T) {
	installFakeJJ(t, "exit 7")

	_, err := New().Setup(
		wsshared.Config{SourceBranch: "main", Dir: "workspace"},
		t.TempDir(),
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "jj workspace add failed") {
		t.Fatalf("Setup() error = %v, want command failure", err)
	}
}

func TestSetup_RequiresJJ(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := New().Setup(
		wsshared.Config{SourceBranch: "main", Dir: "workspace"},
		t.TempDir(),
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "jj is not installed") {
		t.Fatalf("Setup() error = %v, want unavailable-jj error", err)
	}
}

func TestSetup_RequiresSourceRevision(t *testing.T) {
	installFakeJJ(t, "exit 0")

	_, err := New().Setup(wsshared.Config{Dir: "workspace"}, t.TempDir(), false)
	if err == nil || !strings.Contains(err.Error(), "failed to determine source revision") {
		t.Fatalf("Setup() error = %v, want source-revision error", err)
	}
}

func TestAvailable(t *testing.T) {
	installFakeJJ(t, "exit 0")
	if !New().Available() {
		t.Error("Available() = false, want true")
	}

	t.Setenv("PATH", t.TempDir())
	if New().Available() {
		t.Error("Available() = true without jj, want false")
	}
}
