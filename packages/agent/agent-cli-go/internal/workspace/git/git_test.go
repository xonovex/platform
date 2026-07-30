package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	wsshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/workspace/shared"
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
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("test repository\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "-c", "user.name=Xonovex Tests", "-c", "user.email=tests@xonovex.com", "commit", "-m", "test fixture")
	return repoDir
}

func TestSetup_CreatesAndReusesNewBranchWorktree(t *testing.T) {
	repoDir := initializeRepository(t)
	target := filepath.Join(t.TempDir(), "worktree")
	config := wsshared.Config{Branch: "feature/new", SourceBranch: "main", Dir: target}

	result, err := New().Setup(config, repoDir, true)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if result != target {
		t.Errorf("Setup() = %q, want %q", result, target)
	}
	if branch := runGit(t, target, "branch", "--show-current"); branch != config.Branch {
		t.Errorf("worktree branch = %q, want %q", branch, config.Branch)
	}
	if mergeBackTo := runGit(t, repoDir, "config", "--get", "branch.feature/new.mergeBackTo"); mergeBackTo != "main" {
		t.Errorf("mergeBackTo = %q, want main", mergeBackTo)
	}

	reused, err := New().Setup(config, repoDir, true)
	if err != nil {
		t.Fatalf("Setup(reuse) error = %v", err)
	}
	if reused != target {
		t.Errorf("Setup(reuse) = %q, want %q", reused, target)
	}
}

func TestSetup_ReusesSiblingWorktreeFromWorktreeRepositoryDirectory(t *testing.T) {
	repoDir := initializeRepository(t)
	firstWorktree := filepath.Join(t.TempDir(), "first")
	secondWorktree := filepath.Join(t.TempDir(), "second")
	if _, err := New().Setup(wsshared.Config{Branch: "feature/first", SourceBranch: "main", Dir: firstWorktree}, repoDir, false); err != nil {
		t.Fatalf("create first worktree: %v", err)
	}
	if _, err := New().Setup(wsshared.Config{Branch: "feature/second", SourceBranch: "main", Dir: secondWorktree}, repoDir, false); err != nil {
		t.Fatalf("create second worktree: %v", err)
	}

	result, err := New().Setup(wsshared.Config{Branch: "feature/second", Dir: secondWorktree}, firstWorktree, false)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if result != secondWorktree {
		t.Errorf("Setup() = %q, want %q", result, secondWorktree)
	}
}

func TestSetup_CreatesWorktreeForExistingBranch(t *testing.T) {
	repoDir := initializeRepository(t)
	runGit(t, repoDir, "branch", "existing", "main")
	target := filepath.Join(t.TempDir(), "worktree")

	result, err := New().Setup(wsshared.Config{Branch: "existing", Dir: target}, repoDir, true)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if result != target {
		t.Errorf("Setup() = %q, want %q", result, target)
	}
	if branch := runGit(t, target, "branch", "--show-current"); branch != "existing" {
		t.Errorf("worktree branch = %q, want existing", branch)
	}
}

func TestSetup_RejectsExistingNonWorktree(t *testing.T) {
	repoDir := initializeRepository(t)
	target := t.TempDir()

	_, err := New().Setup(wsshared.Config{Branch: "feature/new", Dir: target}, repoDir, false)
	if err == nil || !strings.Contains(err.Error(), "not a worktree") {
		t.Fatalf("Setup() error = %v, want non-worktree error", err)
	}
}

func TestSetup_RejectsWorktreeOnWrongBranch(t *testing.T) {
	repoDir := initializeRepository(t)
	target := filepath.Join(t.TempDir(), "worktree")
	if _, err := New().Setup(wsshared.Config{Branch: "actual", SourceBranch: "main", Dir: target}, repoDir, false); err != nil {
		t.Fatalf("create fixture worktree: %v", err)
	}

	_, err := New().Setup(wsshared.Config{Branch: "expected", Dir: target}, repoDir, false)
	if err == nil || !strings.Contains(err.Error(), "wrong branch") {
		t.Fatalf("Setup() error = %v, want wrong-branch error", err)
	}
}

func TestSetup_RejectsWorktreeFromDifferentRepository(t *testing.T) {
	firstRepo := initializeRepository(t)
	secondRepo := initializeRepository(t)
	target := filepath.Join(t.TempDir(), "worktree")
	if _, err := New().Setup(wsshared.Config{Branch: "feature/first", SourceBranch: "main", Dir: target}, firstRepo, false); err != nil {
		t.Fatalf("create fixture worktree: %v", err)
	}

	_, err := New().Setup(wsshared.Config{Branch: "feature/first", Dir: target}, secondRepo, false)
	if err == nil || !strings.Contains(err.Error(), "different repository") {
		t.Fatalf("Setup() error = %v, want different-repository error", err)
	}
}

func TestSetup_NonRepoReturnsError(t *testing.T) {
	repoDir := t.TempDir()
	config := wsshared.Config{Branch: "test-branch", Dir: filepath.Join(t.TempDir(), "worktree")}

	if _, err := New().Setup(config, repoDir, false); err == nil {
		t.Fatal("Setup() error = nil, want non-repository error")
	}
}

func TestAvailable(t *testing.T) {
	if !New().Available() {
		t.Error("git worktree variant must always report available")
	}
}
