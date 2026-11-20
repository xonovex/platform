//go:build integration

// The cases here drive the real git or jj binary in a temporary repository, so
// they carry the integration build tag and stay out of ci-check. Run them with
// `npx moon run agent-cli-go:go-integration`.
package shared

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost"
)

func runGit(t *testing.T, cwd string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func initializeRepository(t *testing.T) string {
	t.Helper()

	repoDir := t.TempDir()
	runGit(t, repoDir, "init", "-b", "main")
	if err := os.WriteFile(
		filepath.Join(repoDir, "README.md"), []byte("test repository\n"), 0o644,
	); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir,
		"-c", "user.name=Xonovex Tests", "-c", "user.email=tests@xonovex.com",
		"commit", "-m", "test fixture")
	return repoDir
}

func TestExecGit_ReturnsTrimmedStdout(t *testing.T) {
	repoDir := initializeRepository(t)

	output, err := ExecGit(realHost(), []string{"rev-parse", "--abbrev-ref", "HEAD"}, repoDir)

	if err != nil {
		t.Fatalf("ExecGit() error = %v", err)
	}
	if output != "main" {
		t.Errorf("ExecGit() = %q, want %q with no trailing newline", output, "main")
	}
}

// The command runs in the directory it is given, not the test process's own.
func TestExecGit_RunsInTheGivenDirectory(t *testing.T) {
	first := initializeRepository(t)
	second := initializeRepository(t)
	runGit(t, second, "checkout", "-b", "other")

	firstBranch, err := ExecGit(realHost(), []string{"rev-parse", "--abbrev-ref", "HEAD"}, first)
	if err != nil {
		t.Fatalf("ExecGit() error = %v", err)
	}
	secondBranch, err := ExecGit(realHost(), []string{"rev-parse", "--abbrev-ref", "HEAD"}, second)
	if err != nil {
		t.Fatalf("ExecGit() error = %v", err)
	}

	if firstBranch != "main" || secondBranch != "other" {
		t.Errorf("branches = %q and %q, want %q and %q",
			firstBranch, secondBranch, "main", "other")
	}
}

func TestExecGit_ReportsAFailedCommand(t *testing.T) {
	repoDir := initializeRepository(t)

	output, err := ExecGit(realHost(), []string{"rev-parse", "--verify", "absent-branch"}, repoDir)

	if err == nil {
		t.Fatal("ExecGit() must return an error for a failing git command")
	}
	if output != "" {
		t.Errorf("ExecGit() = %q, want empty output alongside the error", output)
	}
}

func TestExecGit_ReportsANonRepositoryDirectory(t *testing.T) {
	if _, err := ExecGit(
		realHost(), []string{"rev-parse", "--abbrev-ref", "HEAD"}, t.TempDir(),
	); err == nil {
		t.Fatal("ExecGit() must return an error outside a git repository")
	}
}

func TestGetCurrentBranchSync_ReturnsTheCheckedOutBranch(t *testing.T) {
	repoDir := initializeRepository(t)
	runGit(t, repoDir, "checkout", "-b", "feature/x")

	if got := GetCurrentBranchSync(realHost(), repoDir); got != "feature/x" {
		t.Errorf("GetCurrentBranchSync() = %q, want %q", got, "feature/x")
	}
}

// A detached HEAD names no branch, so there is no source revision to report.
func TestGetCurrentBranchSync_ReturnsEmptyWhenDetachedInARealRepository(t *testing.T) {
	repoDir := initializeRepository(t)
	head := runGit(t, repoDir, "rev-parse", "HEAD")
	runGit(t, repoDir, "checkout", "--detach", head)

	if got := GetCurrentBranchSync(realHost(), repoDir); got != "" {
		t.Errorf("GetCurrentBranchSync() = %q, want empty for a detached HEAD", got)
	}
}

func TestGetCurrentBranchSync_ReturnsEmptyOutsideARepository(t *testing.T) {
	if got := GetCurrentBranchSync(realHost(), t.TempDir()); got != "" {
		t.Errorf("GetCurrentBranchSync() = %q, want empty outside a repository", got)
	}
}

// realHost is the operating-system adapter, so these cases run the real git and
// jj binaries.
func realHost() oshost.System { return oshost.New(os.Stdin, os.Stdout, os.Stderr) }
