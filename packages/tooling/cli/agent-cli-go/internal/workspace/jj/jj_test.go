package jj

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

const (
	repoDir = "/src/repo"
	target  = "/src/workspace"
)

// onBranch is a host with jj on PATH whose repository at repoDir is on branch.
func onBranch(branch string) *systemtest.Fake {
	host := systemtest.New().AddExecutable("/usr/bin/jj")
	host.Env["PATH"] = "/usr/bin"
	host.Outcomes[systemtest.InDir(repoDir, "git rev-parse --abbrev-ref HEAD")] = systemtest.Outcome{Stdout: branch + "\n"}
	return host
}

func addLine(dir, revision string) string {
	return systemtest.InDir(repoDir, "jj workspace add "+dir+" --revision "+revision)
}

func streamed(host *systemtest.Fake) []string {
	var lines []string
	for _, call := range host.Calls {
		if call.Kind == "stream" {
			lines = append(lines, systemtest.InDir(call.Dir, call.Line()))
		}
	}
	return lines
}

func TestSetup_CreatesAWorkspaceFromTheCurrentBranch(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		host := onBranch("main")
		host.Outcomes[addLine(target, "main")] = systemtest.Outcome{}

		dir, err := New(host, testLog).Setup(wsshared.Config{Dir: target}, repoDir, verbose)

		if err != nil {
			t.Fatalf("Setup() error = %v", err)
		}
		if dir != target {
			t.Errorf("Setup() = %q, want %q", dir, target)
		}
		if got := streamed(host); !slices.Equal(got, []string{addLine(target, "main")}) {
			t.Errorf("Setup() ran %v, want a jj workspace add at revision main", got)
		}
	}
}

func TestSetup_ResolvesARelativeDirectoryAgainstTheRepository(t *testing.T) {
	host := onBranch("main")
	host.Outcomes[addLine(repoDir+"/checkout", "main")] = systemtest.Outcome{}

	dir, err := New(host, testLog).Setup(wsshared.Config{Dir: "checkout"}, repoDir, false)

	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if want := repoDir + "/checkout"; dir != want {
		t.Errorf("Setup() = %q, want %q", dir, want)
	}
}

func TestSetup_ReusesAnExistingWorkspaceWithoutRunningJJ(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		host := onBranch("main").AddDir(target)

		dir, err := New(host, testLog).Setup(wsshared.Config{Dir: target}, repoDir, verbose)

		if err != nil {
			t.Fatalf("Setup() error = %v", err)
		}
		if dir != target {
			t.Errorf("Setup() = %q, want the existing directory %q", dir, target)
		}
		if len(host.Calls) != 0 {
			t.Errorf("Setup() ran %v, want nothing for an existing workspace", host.Calls)
		}
	}
}

func TestSetup_PrefersAnExplicitSourceBranch(t *testing.T) {
	host := onBranch("main")
	host.Outcomes[addLine(target, "release")] = systemtest.Outcome{}

	if _, err := New(host, testLog).Setup(wsshared.Config{Dir: target, SourceBranch: "release"}, repoDir, false); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if got := streamed(host); !slices.Equal(got, []string{addLine(target, "release")}) {
		t.Errorf("Setup() ran %v, want the explicit source branch", got)
	}
}

func TestSetup_RequiresJJOnPath(t *testing.T) {
	host := onBranch("main")
	host.Env["PATH"] = "/opt/bin"

	_, err := New(host, testLog).Setup(wsshared.Config{Dir: target}, repoDir, false)

	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("Setup() error = %v, want an unavailable-jj error", err)
	}
}

func TestSetup_RequiresASourceRevision(t *testing.T) {
	// A detached HEAD names no branch, so there is nothing to branch the
	// workspace from.
	host := onBranch("HEAD")

	_, err := New(host, testLog).Setup(wsshared.Config{Dir: target}, repoDir, false)

	if err == nil || !strings.Contains(err.Error(), "source revision") {
		t.Errorf("Setup() error = %v, want a missing-source-revision error", err)
	}
}

func TestSetup_ReportsAFailedWorkspaceAdd(t *testing.T) {
	host := onBranch("main")
	host.Outcomes[addLine(target, "main")] = systemtest.Outcome{Err: errors.New("exit status 1")}

	_, err := New(host, testLog).Setup(wsshared.Config{Dir: target}, repoDir, false)

	if err == nil || !strings.Contains(err.Error(), "jj workspace add failed") {
		t.Errorf("Setup() error = %v, want a wrapped workspace-add failure", err)
	}
}

func TestAvailable(t *testing.T) {
	if !New(onBranch("main"), testLog).Available() {
		t.Error("Available() must be true when jj resolves on PATH")
	}
	if New(systemtest.New(), testLog).Available() {
		t.Error("Available() must be false when jj is absent")
	}
}
