package jj

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	wsshared "github.com/xonovex/platform/packages/agent/agent-cli-go/internal/workspace/shared"
)

// fakeRunner answers from a recording keyed by the joined command line, so these
// cases exercise Setup without launching jj or git.
type fakeRunner struct {
	output    map[string]string
	err       map[string]error
	absent    map[string]bool
	streamed  []string
	streamErr error
}

func key(name string, args []string) string {
	return name + " " + strings.Join(args, " ")
}

func (f *fakeRunner) Capture(name string, args []string, _ string) (string, error) {
	k := key(name, args)
	if err, ok := f.err[k]; ok {
		return "", err
	}
	return f.output[k], nil
}

func (f *fakeRunner) Stream(name string, args []string, _ string) error {
	f.streamed = append(f.streamed, key(name, args))
	return f.streamErr
}

func (f *fakeRunner) Available(name string) bool { return !f.absent[name] }

func onBranch(branch string) *fakeRunner {
	return &fakeRunner{output: map[string]string{
		"git rev-parse --abbrev-ref HEAD": branch,
	}}
}

func TestSetup_CreatesAWorkspaceFromTheCurrentBranch(t *testing.T) {
	runner := onBranch("main")
	target := filepath.Join(t.TempDir(), "workspace")

	dir, err := New(runner).Setup(wsshared.Config{Dir: target}, t.TempDir(), false)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if dir != target {
		t.Errorf("Setup() = %q, want %q", dir, target)
	}
	if len(runner.streamed) != 1 ||
		!strings.Contains(runner.streamed[0], "jj workspace add") ||
		!strings.HasSuffix(runner.streamed[0], "--revision main") {
		t.Errorf("Setup() ran %v, want a jj workspace add at revision main", runner.streamed)
	}
}

func TestSetup_ResolvesARelativeDirectoryAgainstTheRepository(t *testing.T) {
	runner := onBranch("main")
	repoDir := t.TempDir()

	dir, err := New(runner).Setup(wsshared.Config{Dir: "checkout"}, repoDir, false)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if want := filepath.Join(repoDir, "checkout"); dir != want {
		t.Errorf("Setup() = %q, want %q", dir, want)
	}
}

func TestSetup_ReusesAnExistingWorkspaceWithoutRunningJJ(t *testing.T) {
	runner := onBranch("main")
	existing := t.TempDir()

	dir, err := New(runner).Setup(wsshared.Config{Dir: existing}, t.TempDir(), true)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if dir != existing {
		t.Errorf("Setup() = %q, want the existing directory %q", dir, existing)
	}
	if len(runner.streamed) != 0 {
		t.Errorf("Setup() ran %v, want nothing for an existing workspace", runner.streamed)
	}
}

func TestSetup_PrefersAnExplicitSourceBranch(t *testing.T) {
	runner := onBranch("main")
	target := filepath.Join(t.TempDir(), "workspace")

	if _, err := New(runner).Setup(
		wsshared.Config{Dir: target, SourceBranch: "release"}, t.TempDir(), false,
	); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if !strings.HasSuffix(runner.streamed[0], "--revision release") {
		t.Errorf("Setup() ran %v, want the explicit source branch", runner.streamed)
	}
}

func TestSetup_RequiresJJOnPath(t *testing.T) {
	runner := onBranch("main")
	runner.absent = map[string]bool{"jj": true}

	_, err := New(runner).Setup(
		wsshared.Config{Dir: filepath.Join(t.TempDir(), "workspace")}, t.TempDir(), false,
	)
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("Setup() error = %v, want an unavailable-jj error", err)
	}
}

func TestSetup_RequiresASourceRevision(t *testing.T) {
	// A detached HEAD names no branch, so there is nothing to branch the
	// workspace from.
	runner := onBranch("HEAD")

	_, err := New(runner).Setup(
		wsshared.Config{Dir: filepath.Join(t.TempDir(), "workspace")}, t.TempDir(), false,
	)
	if err == nil || !strings.Contains(err.Error(), "source revision") {
		t.Errorf("Setup() error = %v, want a missing-source-revision error", err)
	}
}

func TestSetup_ReportsAFailedWorkspaceAdd(t *testing.T) {
	runner := onBranch("main")
	runner.streamErr = errors.New("exit status 1")

	_, err := New(runner).Setup(
		wsshared.Config{Dir: filepath.Join(t.TempDir(), "workspace")}, t.TempDir(), false,
	)
	if err == nil || !strings.Contains(err.Error(), "jj workspace add failed") {
		t.Errorf("Setup() error = %v, want a wrapped workspace-add failure", err)
	}
}

func TestAvailable(t *testing.T) {
	if !New(&fakeRunner{}).Available() {
		t.Error("Available() must be true when jj resolves on PATH")
	}
	if New(&fakeRunner{absent: map[string]bool{"jj": true}}).Available() {
		t.Error("Available() must be false when jj is absent")
	}
}
