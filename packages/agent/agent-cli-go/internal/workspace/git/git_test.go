package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	wsshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/workspace/shared"
)

// fakeRunner answers from a recording keyed by working directory and command
// line, so these cases exercise Setup without launching git. commonGitDir asks
// the same question of two directories, so the cwd is part of the key.
type fakeRunner struct {
	output   map[string]string
	err      map[string]error
	streamed []string
}

func key(cwd, name string, args []string) string {
	return cwd + "\x00" + name + " " + strings.Join(args, " ")
}

func (f *fakeRunner) Capture(name string, args []string, cwd string) (string, error) {
	k := key(cwd, name, args)
	if err, ok := f.err[k]; ok {
		return "", err
	}
	if out, ok := f.output[k]; ok {
		return out, nil
	}
	return "", errors.New("no recording for " + k)
}

func (f *fakeRunner) Stream(name string, args []string, cwd string) error {
	f.streamed = append(f.streamed, key(cwd, name, args))
	return f.err[key(cwd, name, args)]
}

func (f *fakeRunner) Available(string) bool { return true }

// worktreeAt writes the .git pointer file that marks dir as a git worktree.
func worktreeAt(t *testing.T, dir, gitDir string) {
	t.Helper()
	if err := os.WriteFile(
		filepath.Join(dir, ".git"), []byte("gitdir: "+gitDir+"\n"), 0o600,
	); err != nil {
		t.Fatalf("write .git pointer: %v", err)
	}
}

func commonDir(cwd string) string {
	return key(cwd, "git", []string{"rev-parse", "--git-common-dir"})
}

func TestSetup_ReusesAWorktreeOnTheExpectedBranch(t *testing.T) {
	repoDir := t.TempDir()
	target := t.TempDir()
	// commonGitDir resolves symlinks, so the recorded common dir must exist.
	shared := t.TempDir()
	worktreeAt(t, target, shared)
	runner := &fakeRunner{output: map[string]string{
		commonDir(target):  shared,
		commonDir(repoDir): shared,
		key(target, "git", []string{"rev-parse", "--abbrev-ref", "HEAD"}): "feature/x",
	}}

	dir, err := New(runner).Setup(
		wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, true,
	)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if dir != target {
		t.Errorf("Setup() = %q, want %q", dir, target)
	}
	if len(runner.streamed) != 0 {
		t.Errorf("Setup() ran %v, want nothing for a reusable worktree", runner.streamed)
	}
}

func TestSetup_RejectsADirectoryThatIsNotAWorktree(t *testing.T) {
	target := t.TempDir()

	_, err := New(&fakeRunner{}).Setup(
		wsshared.Config{Branch: "feature/x", Dir: target}, t.TempDir(), false,
	)
	if err == nil || !strings.Contains(err.Error(), "not a worktree") {
		t.Errorf("Setup() error = %v, want a not-a-worktree error", err)
	}
}

func TestSetup_RejectsAWorktreeFromADifferentRepository(t *testing.T) {
	repoDir := t.TempDir()
	target := t.TempDir()
	other, shared := t.TempDir(), t.TempDir()
	worktreeAt(t, target, other)
	runner := &fakeRunner{output: map[string]string{
		commonDir(target):  other,
		commonDir(repoDir): shared,
	}}

	_, err := New(runner).Setup(
		wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, false,
	)
	if err == nil || !strings.Contains(err.Error(), "different repository") {
		t.Errorf("Setup() error = %v, want a different-repository error", err)
	}
}

func TestSetup_RejectsAWorktreeOnTheWrongBranch(t *testing.T) {
	repoDir := t.TempDir()
	target := t.TempDir()
	shared := t.TempDir()
	worktreeAt(t, target, shared)
	runner := &fakeRunner{output: map[string]string{
		commonDir(target):  shared,
		commonDir(repoDir): shared,
		key(target, "git", []string{"rev-parse", "--abbrev-ref", "HEAD"}): "main",
	}}

	_, err := New(runner).Setup(
		wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, false,
	)
	if err == nil || !strings.Contains(err.Error(), "wrong branch") {
		t.Errorf("Setup() error = %v, want a wrong-branch error", err)
	}
}

func TestSetup_CreatesAWorktreeForAnExistingBranch(t *testing.T) {
	repoDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "worktree")
	runner := &fakeRunner{output: map[string]string{
		// An existing branch resolves, so no new branch is created.
		key(repoDir, "git", []string{"rev-parse", "--verify", "refs/heads/existing"}): "abc123",
	}}

	dir, err := New(runner).Setup(
		wsshared.Config{Branch: "existing", Dir: target}, repoDir, true,
	)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if dir != target {
		t.Errorf("Setup() = %q, want %q", dir, target)
	}
	if len(runner.streamed) != 1 ||
		!strings.Contains(runner.streamed[0], "worktree add "+target+" existing") {
		t.Errorf("Setup() ran %v, want one worktree add for the existing branch", runner.streamed)
	}
}

func TestSetup_CreatesANewBranchWorktreeAndRecordsMergeBackTo(t *testing.T) {
	repoDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "worktree")
	runner := &fakeRunner{
		output: map[string]string{
			key(repoDir, "git", []string{"rev-parse", "--abbrev-ref", "HEAD"}): "main",
			key(repoDir, "git", []string{
				"config", "branch.feature/new.mergeBackTo", "main",
			}): "",
		},
		err: map[string]error{
			key(repoDir, "git", []string{"rev-parse", "--verify", "refs/heads/feature/new"}): errors.New("unknown revision"),
		},
	}

	dir, err := New(runner).Setup(
		wsshared.Config{Branch: "feature/new", Dir: target}, repoDir, true,
	)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if dir != target {
		t.Errorf("Setup() = %q, want %q", dir, target)
	}
	if len(runner.streamed) != 1 ||
		!strings.Contains(runner.streamed[0], "-b feature/new main") {
		t.Errorf("Setup() ran %v, want one worktree add branching from main", runner.streamed)
	}
}

func TestSetup_PrefersAnExplicitSourceBranch(t *testing.T) {
	repoDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "worktree")
	runner := &fakeRunner{
		output: map[string]string{
			key(repoDir, "git", []string{
				"config", "branch.feature/new.mergeBackTo", "release",
			}): "",
		},
		err: map[string]error{
			key(repoDir, "git", []string{"rev-parse", "--verify", "refs/heads/feature/new"}): errors.New("unknown revision"),
		},
	}

	if _, err := New(runner).Setup(
		wsshared.Config{Branch: "feature/new", Dir: target, SourceBranch: "release"},
		repoDir, false,
	); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if !strings.Contains(runner.streamed[0], "-b feature/new release") {
		t.Errorf("Setup() ran %v, want the explicit source branch", runner.streamed)
	}
}

func TestSetup_RequiresASourceBranch(t *testing.T) {
	repoDir := t.TempDir()
	runner := &fakeRunner{
		output: map[string]string{
			// A detached HEAD names no branch to start the new one from.
			key(repoDir, "git", []string{"rev-parse", "--abbrev-ref", "HEAD"}): "HEAD",
		},
		err: map[string]error{
			key(repoDir, "git", []string{"rev-parse", "--verify", "refs/heads/feature/new"}): errors.New("unknown revision"),
		},
	}

	_, err := New(runner).Setup(
		wsshared.Config{Branch: "feature/new", Dir: filepath.Join(t.TempDir(), "worktree")},
		repoDir, false,
	)
	if err == nil || !strings.Contains(err.Error(), "source branch") {
		t.Errorf("Setup() error = %v, want a missing-source-branch error", err)
	}
}

func TestSetup_ResolvesARelativeDirectoryAgainstTheRepository(t *testing.T) {
	repoDir := t.TempDir()
	runner := &fakeRunner{output: map[string]string{
		key(repoDir, "git", []string{"rev-parse", "--verify", "refs/heads/existing"}): "abc123",
	}}

	dir, err := New(runner).Setup(
		wsshared.Config{Branch: "existing", Dir: "checkout"}, repoDir, false,
	)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if want := filepath.Join(repoDir, "checkout"); dir != want {
		t.Errorf("Setup() = %q, want %q", dir, want)
	}
}

func TestSetup_ReportsAFailedWorktreeAdd(t *testing.T) {
	repoDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "worktree")
	runner := &fakeRunner{
		output: map[string]string{
			key(repoDir, "git", []string{"rev-parse", "--verify", "refs/heads/existing"}): "abc123",
		},
		err: map[string]error{
			key(repoDir, "git", []string{"worktree", "add", target, "existing"}): errors.New("exit status 128"),
		},
	}

	if _, err := New(runner).Setup(
		wsshared.Config{Branch: "existing", Dir: target}, repoDir, false,
	); err == nil {
		t.Fatal("Setup() must surface a failed worktree add")
	}
}

func TestAvailable(t *testing.T) {
	if !New(&fakeRunner{}).Available() {
		t.Error("git worktree variant must always report available")
	}
}
