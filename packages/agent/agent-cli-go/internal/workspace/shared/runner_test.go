package shared

import (
	"errors"
	"testing"
)

// fakeRunner answers from a recording keyed by the joined command line, so these
// cases exercise the git plumbing without launching git.
type fakeRunner struct {
	output map[string]string
	err    map[string]error
	calls  []string
}

func key(name string, args []string) string {
	out := name
	for _, arg := range args {
		out += " " + arg
	}
	return out
}

func (f *fakeRunner) Capture(name string, args []string, _ string) (string, error) {
	k := key(name, args)
	f.calls = append(f.calls, k)
	if err, ok := f.err[k]; ok {
		return "", err
	}
	return f.output[k], nil
}

func (f *fakeRunner) Stream(name string, args []string, _ string) error {
	k := key(name, args)
	f.calls = append(f.calls, k)
	return f.err[k]
}

func (f *fakeRunner) Available(name string) bool {
	_, missing := f.err["lookpath "+name]
	return !missing
}

func TestExecGit_ReturnsCapturedStdout(t *testing.T) {
	runner := &fakeRunner{output: map[string]string{"git rev-parse HEAD": "abc123"}}

	got, err := ExecGit(runner, []string{"rev-parse", "HEAD"}, "/repo")
	if err != nil {
		t.Fatalf("ExecGit() error = %v", err)
	}
	if got != "abc123" {
		t.Errorf("ExecGit() = %q, want %q", got, "abc123")
	}
}

func TestExecGit_PropagatesTheCommandError(t *testing.T) {
	runner := &fakeRunner{err: map[string]error{"git rev-parse HEAD": errors.New("boom")}}

	if _, err := ExecGit(runner, []string{"rev-parse", "HEAD"}, "/repo"); err == nil {
		t.Fatal("ExecGit() must surface the runner error")
	}
}

func TestGetCurrentBranchSync_ReturnsTheBranch(t *testing.T) {
	runner := &fakeRunner{output: map[string]string{
		"git rev-parse --abbrev-ref HEAD": "feature/x",
	}}

	if got := GetCurrentBranchSync(runner, "/repo"); got != "feature/x" {
		t.Errorf("GetCurrentBranchSync() = %q, want %q", got, "feature/x")
	}
}

func TestGetCurrentBranchSync_ReturnsEmptyWhenDetached(t *testing.T) {
	// A detached HEAD makes git echo the literal "HEAD", which is not a branch.
	runner := &fakeRunner{output: map[string]string{
		"git rev-parse --abbrev-ref HEAD": "HEAD",
	}}

	if got := GetCurrentBranchSync(runner, "/repo"); got != "" {
		t.Errorf("GetCurrentBranchSync() = %q, want empty for a detached HEAD", got)
	}
}

func TestGetCurrentBranchSync_ReturnsEmptyOnError(t *testing.T) {
	runner := &fakeRunner{err: map[string]error{
		"git rev-parse --abbrev-ref HEAD": errors.New("not a repository"),
	}}

	if got := GetCurrentBranchSync(runner, "/repo"); got != "" {
		t.Errorf("GetCurrentBranchSync() = %q, want empty outside a repository", got)
	}
}

func TestNewExecRunner_ReportsAnAbsentBinary(t *testing.T) {
	// Available is the one execRunner method with no side effect beyond a PATH
	// lookup, so the default tier can check its negative case.
	if NewExecRunner().Available("definitely-not-a-real-binary-xyz") {
		t.Error("Available() must be false for a binary that is not on PATH")
	}
}
