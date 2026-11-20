package git

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

const (
	repoDir   = "/src/repo"
	sharedDir = "/src/repo/.git"
	target    = "/src/checkout"
)

// commonDirLine is the git query whose answer commonGitDir resolves; it asks the
// same question of two directories, so each recording carries its directory.
func commonDirLine(dir string) string {
	return systemtest.InDir(dir, "git rev-parse --git-common-dir")
}

func branchLine(dir string) string {
	return systemtest.InDir(dir, "git rev-parse --abbrev-ref HEAD")
}

func verifyLine(branch string) string {
	return systemtest.InDir(repoDir, "git rev-parse --verify refs/heads/"+branch)
}

// repoHost is a repository at repoDir whose git directory is sharedDir.
func repoHost() *systemtest.Fake {
	host := systemtest.New().AddDir(sharedDir)
	host.Outcomes[commonDirLine(repoDir)] = systemtest.Outcome{Stdout: ".git\n"}
	return host
}

// worktreeAt writes the .git pointer file that marks dir as a git worktree whose
// common git directory is gitDir.
func worktreeAt(host *systemtest.Fake, dir, gitDir string) {
	host.AddDir(gitDir).AddFile(dir+"/.git", "gitdir: "+gitDir+"/worktrees/x\n", 0o644)
	host.Outcomes[commonDirLine(dir)] = systemtest.Outcome{Stdout: gitDir + "\n"}
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

func TestSetup_ReusesAWorktreeOnTheExpectedBranch(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		host := repoHost()
		worktreeAt(host, target, sharedDir)
		host.Outcomes[branchLine(target)] = systemtest.Outcome{Stdout: "feature/x\n"}

		dir, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, verbose)

		if err != nil {
			t.Fatalf("Setup() error = %v", err)
		}
		if dir != target {
			t.Errorf("Setup() = %q, want %q", dir, target)
		}
		if got := streamed(host); len(got) != 0 {
			t.Errorf("Setup() ran %v, want nothing for a reusable worktree", got)
		}
	}
}

func TestSetup_ReusesAWorktreeThroughASymlinkedGitDirectory(t *testing.T) {
	host := repoHost().AddLink("/links/git", sharedDir)
	worktreeAt(host, target, "/links/git")
	host.Outcomes[branchLine(target)] = systemtest.Outcome{Stdout: "feature/x\n"}

	if _, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, false); err != nil {
		t.Fatalf("Setup() error = %v, want the symlinked git directory to match", err)
	}
}

func TestSetup_RejectsADirectoryThatIsNotAWorktree(t *testing.T) {
	tests := []struct {
		name  string
		setup func(host *systemtest.Fake)
	}{
		{"plain directory", func(host *systemtest.Fake) { host.AddDir(target) }},
		{"git directory", func(host *systemtest.Fake) { host.AddDir(target + "/.git") }},
		{"other .git file", func(host *systemtest.Fake) { host.AddFile(target+"/.git", "not a pointer\n", 0o644) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := repoHost()
			tt.setup(host)

			_, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, false)

			if err == nil || !strings.Contains(err.Error(), "not a worktree") {
				t.Errorf("Setup() error = %v, want a not-a-worktree error", err)
			}
		})
	}
}

func TestSetup_RejectsAWorktreeFromADifferentRepository(t *testing.T) {
	tests := []struct {
		name  string
		setup func(host *systemtest.Fake)
	}{
		{"other repository", func(host *systemtest.Fake) { worktreeAt(host, target, "/src/other/.git") }},
		{"worktree query fails", func(host *systemtest.Fake) {
			worktreeAt(host, target, sharedDir)
			host.Outcomes[commonDirLine(target)] = systemtest.Outcome{Err: errors.New("not a repository")}
		}},
		{"repository query fails", func(host *systemtest.Fake) {
			worktreeAt(host, target, sharedDir)
			host.Outcomes[commonDirLine(repoDir)] = systemtest.Outcome{Err: errors.New("not a repository")}
		}},
		{"git directory is missing", func(host *systemtest.Fake) {
			worktreeAt(host, target, sharedDir)
			host.Outcomes[commonDirLine(target)] = systemtest.Outcome{Stdout: "/src/missing/.git\n"}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host := repoHost()
			tt.setup(host)

			_, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, false)

			if err == nil || !strings.Contains(err.Error(), "different repository") {
				t.Errorf("Setup() error = %v, want a different-repository error", err)
			}
		})
	}
}

func TestSetup_RejectsAWorktreeOnTheWrongBranch(t *testing.T) {
	host := repoHost()
	worktreeAt(host, target, sharedDir)
	host.Outcomes[branchLine(target)] = systemtest.Outcome{Stdout: "main\n"}

	_, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/x", Dir: target}, repoDir, false)

	if err == nil || !strings.Contains(err.Error(), "wrong branch") {
		t.Errorf("Setup() error = %v, want a wrong-branch error", err)
	}
}

func TestSetup_CreatesAWorktreeForAnExistingBranch(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		host := repoHost()
		host.Outcomes[verifyLine("existing")] = systemtest.Outcome{Stdout: "abc123\n"}
		add := systemtest.InDir(repoDir, "git worktree add "+target+" existing")
		host.Outcomes[add] = systemtest.Outcome{}

		dir, err := New(host, testLog).Setup(wsshared.Config{Branch: "existing", Dir: target}, repoDir, verbose)

		if err != nil {
			t.Fatalf("Setup() error = %v", err)
		}
		if dir != target {
			t.Errorf("Setup() = %q, want %q", dir, target)
		}
		if got := streamed(host); !slices.Equal(got, []string{add}) {
			t.Errorf("Setup() ran %v, want one worktree add for the existing branch", got)
		}
	}
}

func newBranchHost(source string) (*systemtest.Fake, string) {
	host := repoHost()
	host.Outcomes[verifyLine("feature/new")] = systemtest.Outcome{Err: errors.New("unknown revision")}
	add := systemtest.InDir(repoDir, "git worktree add "+target+" -b feature/new "+source)
	host.Outcomes[add] = systemtest.Outcome{}
	host.Outcomes[systemtest.InDir(repoDir, "git config branch.feature/new.mergeBackTo "+source)] = systemtest.Outcome{}
	return host, add
}

func TestSetup_CreatesANewBranchWorktreeAndRecordsMergeBackTo(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		host, add := newBranchHost("main")
		host.Outcomes[branchLine(repoDir)] = systemtest.Outcome{Stdout: "main\n"}

		dir, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/new", Dir: target}, repoDir, verbose)

		if err != nil {
			t.Fatalf("Setup() error = %v", err)
		}
		if dir != target {
			t.Errorf("Setup() = %q, want %q", dir, target)
		}
		if got := streamed(host); !slices.Equal(got, []string{add}) {
			t.Errorf("Setup() ran %v, want one worktree add branching from main", got)
		}
		last := host.Calls[len(host.Calls)-1]
		if last.Line() != "git config branch.feature/new.mergeBackTo main" {
			t.Errorf("last call = %q, want the mergeBackTo record", last.Line())
		}
	}
}

func TestSetup_PrefersAnExplicitSourceBranch(t *testing.T) {
	host, add := newBranchHost("release")

	if _, err := New(host, testLog).Setup(
		wsshared.Config{Branch: "feature/new", Dir: target, SourceBranch: "release"}, repoDir, false,
	); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if got := streamed(host); !slices.Equal(got, []string{add}) {
		t.Errorf("Setup() ran %v, want the explicit source branch", got)
	}
}

func TestSetup_RequiresASourceBranch(t *testing.T) {
	host, _ := newBranchHost("main")
	// A detached HEAD names no branch to start the new one from.
	host.Outcomes[branchLine(repoDir)] = systemtest.Outcome{Stdout: "HEAD\n"}

	_, err := New(host, testLog).Setup(wsshared.Config{Branch: "feature/new", Dir: target}, repoDir, false)

	if err == nil || !strings.Contains(err.Error(), "source branch") {
		t.Errorf("Setup() error = %v, want a missing-source-branch error", err)
	}
}

func TestSetup_ReportsAFailedMergeBackRecord(t *testing.T) {
	host, _ := newBranchHost("main")
	host.Outcomes[systemtest.InDir(repoDir, "git config branch.feature/new.mergeBackTo main")] = systemtest.Outcome{Err: errors.New("locked")}

	if _, err := New(host, testLog).Setup(
		wsshared.Config{Branch: "feature/new", Dir: target, SourceBranch: "main"}, repoDir, false,
	); err == nil {
		t.Fatal("Setup() must surface a failed mergeBackTo record")
	}
}

func TestSetup_ReportsAFailedNewBranchWorktreeAdd(t *testing.T) {
	host, add := newBranchHost("main")
	host.Outcomes[add] = systemtest.Outcome{Err: errors.New("exit status 128")}

	if _, err := New(host, testLog).Setup(
		wsshared.Config{Branch: "feature/new", Dir: target, SourceBranch: "main"}, repoDir, false,
	); err == nil {
		t.Fatal("Setup() must surface a failed worktree add")
	}
}

func TestSetup_ResolvesARelativeDirectoryAgainstTheRepository(t *testing.T) {
	host := repoHost()
	host.Outcomes[verifyLine("existing")] = systemtest.Outcome{Stdout: "abc123\n"}
	host.Outcomes[systemtest.InDir(repoDir, "git worktree add checkout existing")] = systemtest.Outcome{}

	dir, err := New(host, testLog).Setup(wsshared.Config{Branch: "existing", Dir: "checkout"}, repoDir, false)

	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if want := repoDir + "/checkout"; dir != want {
		t.Errorf("Setup() = %q, want %q", dir, want)
	}
}

func TestSetup_ReportsAFailedWorktreeAdd(t *testing.T) {
	host := repoHost()
	host.Outcomes[verifyLine("existing")] = systemtest.Outcome{Stdout: "abc123\n"}
	host.Outcomes[systemtest.InDir(repoDir, "git worktree add "+target+" existing")] = systemtest.Outcome{Err: errors.New("exit status 128")}

	if _, err := New(host, testLog).Setup(wsshared.Config{Branch: "existing", Dir: target}, repoDir, false); err == nil {
		t.Fatal("Setup() must surface a failed worktree add")
	}
}

func TestAvailable(t *testing.T) {
	if !New(systemtest.New(), testLog).Available() {
		t.Error("git worktree variant must always report available")
	}
}
