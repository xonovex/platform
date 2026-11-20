package git_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process/processtest"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/git"
)

func source() sources.Source {
	return sources.Source{
		Kind: sources.Git, Checkout: "../up", Version: "1.0.0", Commit: "abc1234",
		Watch: []sources.Watch{
			{Path: "src/components", Feeds: []string{"references/components.md", "references/navigation.md"}},
			{Path: "src/hooks", Feeds: []string{"references/hooks.md"}},
		},
	}
}

func repo() fs.FileSystem { return memfs.New(map[string]string{"cat/../up/.git/HEAD": "ref"}) }

func observeWith(t *testing.T, fsys fs.FileSystem, fake *processtest.Fake, s sources.Source, pull bool) (observe.Observation, error) {
	t.Helper()
	return git.Transport{FS: fsys, Runner: fake}.Observe(context.Background(), observe.Request{Source: s, Dir: "cat", Pull: pull})
}

func scripted() *processtest.Fake {
	return processtest.New().
		On("git tag --list", processtest.Reply{Stdout: "v1.0.0\nv1.2.0\nv1.10.0\nnightly\nv2.0.0-rc1\n"}).
		On("git rev-parse --verify --quiet abc1234^{commit}", processtest.Reply{}).
		On("git log --oneline abc1234..v1.10.0 -- src/components src/hooks", processtest.Reply{Stdout: "c3 third\nc2 second\nc1 first\n"}).
		On("git diff --name-only abc1234..v1.10.0 -- src/components src/hooks", processtest.Reply{Stdout: "src/components/a.ts\nsrc/hooks/b.ts\nsrc/components/c.ts\n"})
}

func TestObserveReportsTheDriftSinceThePinnedCommit(t *testing.T) {
	got, err := observeWith(t, repo(), scripted(), source(), false)
	if err != nil {
		t.Fatal(err)
	}
	g := got.Git
	if got.Transport != git.Name || g == nil || !g.Resolved || g.Checkout != "../up" {
		t.Fatalf("observation = %+v", got)
	}
	if g.LatestVersion != "v1.10.0" || g.PinnedVersion != "1.0.0" || !g.Behind {
		t.Errorf("versions = pinned %q latest %q behind %v", g.PinnedVersion, g.LatestVersion, g.Behind)
	}
	if g.Ref != "abc1234" || g.CommitCount != 3 || len(g.Commits) != 3 {
		t.Errorf("commits = %d %v", g.CommitCount, g.Commits)
	}
	wantFeeds := "references/components.md references/hooks.md references/navigation.md"
	if strings.Join(g.ReviewFeeds, " ") != wantFeeds {
		t.Errorf("review feeds = %v", g.ReviewFeeds)
	}
	if len(g.ChangedFiles) != 3 {
		t.Errorf("changed files = %v", g.ChangedFiles)
	}
}

func TestObserveStaysCurrentWhenTheLatestReleaseIsPinned(t *testing.T) {
	s := source()
	s.Version = "v1.10.0"
	got, err := observeWith(t, repo(), scripted(), s, false)
	if err != nil || got.Git.Behind {
		t.Fatalf("behind = %v, err = %v", got.Git.Behind, err)
	}
}

func TestObserveDiffsToHeadWithoutAReleaseTagAndWithoutWatchedPaths(t *testing.T) {
	s := sources.Source{Kind: sources.Git, Checkout: "../up", Commit: "abc1234"}
	fake := processtest.New().
		On("git tag --list", processtest.Reply{Stdout: "nightly\n"}).
		On("git rev-parse --verify --quiet abc1234^{commit}", processtest.Reply{}).
		On("git log --oneline abc1234..HEAD", processtest.Reply{Stdout: "c1 first\n"}).
		On("git diff --name-only abc1234..HEAD", processtest.Reply{Stdout: "x.go\n"})
	got, err := observeWith(t, repo(), fake, s, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Git.LatestVersion != "" || got.Git.Behind || got.Git.CommitCount != 1 || len(got.Git.ReviewFeeds) != 0 {
		t.Errorf("git = %+v", got.Git)
	}
}

func TestObserveFallsBackToTheVersionWhenNoCommitIsPinned(t *testing.T) {
	s := sources.Source{Kind: sources.Git, Checkout: "../up", Version: "1.0.0"}
	fake := processtest.New().
		On("git tag --list", processtest.Reply{Stdout: "1.0.0\n"}).
		On("git rev-parse --verify --quiet 1.0.0^{commit}", processtest.Reply{}).
		On("git log --oneline 1.0.0..1.0.0", processtest.Reply{}).
		On("git diff --name-only 1.0.0..1.0.0", processtest.Reply{})
	got, err := observeWith(t, repo(), fake, s, false)
	if err != nil || got.Git.Ref != "1.0.0" || got.Git.CommitCount != 0 || got.Git.Behind {
		t.Fatalf("git = %+v, err = %v", got.Git, err)
	}
}

func TestObserveCapsTheListedCommits(t *testing.T) {
	var many strings.Builder
	for i := 0; i < 30; i++ {
		fmt.Fprintf(&many, "c%d subject\n", i)
	}
	s := sources.Source{Kind: sources.Git, Checkout: "../up", Commit: "abc1234"}
	fake := processtest.New().
		On("git tag --list", processtest.Reply{}).
		On("git rev-parse --verify --quiet abc1234^{commit}", processtest.Reply{}).
		On("git log --oneline abc1234..HEAD", processtest.Reply{Stdout: many.String()}).
		On("git diff --name-only abc1234..HEAD", processtest.Reply{})
	got, err := observeWith(t, repo(), fake, s, false)
	if err != nil || got.Git.CommitCount != 30 || len(got.Git.Commits) != 20 {
		t.Fatalf("count = %d, listed = %d, err = %v", got.Git.CommitCount, len(got.Git.Commits), err)
	}
}

func TestObserveReportsAnUnusableCheckoutAsState(t *testing.T) {
	t.Run("no repository", func(t *testing.T) {
		got, err := observeWith(t, memfs.New(nil), processtest.New(), source(), false)
		if err != nil || got.Git.Resolved {
			t.Fatalf("git = %+v, err = %v", got.Git, err)
		}
	})
	t.Run("pull failed", func(t *testing.T) {
		fake := processtest.New().On("git fetch --tags --quiet", processtest.Reply{ExitCode: 128})
		got, err := observeWith(t, repo(), fake, source(), true)
		if err != nil || !got.Git.PullFailed || len(fake.Calls) != 1 {
			t.Fatalf("git = %+v, err = %v", got.Git, err)
		}
	})
	t.Run("pull succeeded", func(t *testing.T) {
		fake := scripted().On("git fetch --tags --quiet", processtest.Reply{})
		got, err := observeWith(t, repo(), fake, source(), true)
		if err != nil || got.Git.PullFailed || got.Git.CommitCount != 3 || fake.Calls[0].Dir != "cat/../up" && fake.Calls[0].Dir != "up" {
			t.Fatalf("git = %+v, err = %v, first call %+v", got.Git, err, fake.Calls[0])
		}
	})
	t.Run("nothing pinned", func(t *testing.T) {
		s := sources.Source{Kind: sources.Git, Checkout: "../up"}
		fake := processtest.New().On("git tag --list", processtest.Reply{Stdout: "v1.0.0\n"})
		got, err := observeWith(t, repo(), fake, s, false)
		if err != nil || !got.Git.NoRef || got.Git.LatestVersion != "v1.0.0" || got.Git.Behind {
			t.Fatalf("git = %+v, err = %v", got.Git, err)
		}
	})
	t.Run("pinned ref missing", func(t *testing.T) {
		fake := processtest.New().
			On("git tag --list", processtest.Reply{}).
			On("git rev-parse --verify --quiet abc1234^{commit}", processtest.Reply{ExitCode: 1})
		got, err := observeWith(t, repo(), fake, source(), false)
		if err != nil || !got.Git.RefMissing {
			t.Fatalf("git = %+v, err = %v", got.Git, err)
		}
	})
	t.Run("a failing tag listing reads as no release", func(t *testing.T) {
		fake := processtest.New().
			On("git tag --list", processtest.Reply{ExitCode: 128}).
			On("git rev-parse --verify --quiet abc1234^{commit}", processtest.Reply{ExitCode: 1})
		got, err := observeWith(t, repo(), fake, source(), false)
		if err != nil || got.Git.LatestVersion != "" {
			t.Fatalf("git = %+v, err = %v", got.Git, err)
		}
	})
	t.Run("a failing log reads as no commits", func(t *testing.T) {
		fake := processtest.New().
			On("git tag --list", processtest.Reply{}).
			On("git rev-parse --verify --quiet abc1234^{commit}", processtest.Reply{}).
			On("git log --oneline abc1234..HEAD -- src/components src/hooks", processtest.Reply{ExitCode: 128}).
			On("git diff --name-only abc1234..HEAD -- src/components src/hooks", processtest.Reply{ExitCode: 128})
		got, err := observeWith(t, repo(), fake, source(), false)
		if err != nil || got.Git.CommitCount != 0 || len(got.Git.ChangedFiles) != 0 {
			t.Fatalf("git = %+v, err = %v", got.Git, err)
		}
	})
}

func TestObservePropagatesAFaultOfTheInfrastructure(t *testing.T) {
	boom := errors.New("git is not installed")
	commands := []string{
		"git fetch --tags --quiet",
		"git tag --list",
		"git rev-parse --verify --quiet abc1234^{commit}",
		"git log --oneline abc1234..v1.10.0 -- src/components src/hooks",
		"git diff --name-only abc1234..v1.10.0 -- src/components src/hooks",
	}
	for _, failing := range commands {
		t.Run(failing, func(t *testing.T) {
			fake := scripted().On("git fetch --tags --quiet", processtest.Reply{}).On(failing, processtest.Reply{Err: boom})
			if _, err := observeWith(t, repo(), fake, source(), true); !errors.Is(err, boom) {
				t.Fatalf("err = %v; want %v", err, boom)
			}
		})
	}
	t.Run("stat", func(t *testing.T) {
		_, err := observeWith(t, statFails{repo(), boom}, processtest.New(), source(), false)
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v", err)
		}
	})
}

type statFails struct {
	fs.FileSystem
	err error
}

func (s statFails) Stat(string) (fs.Info, error) { return fs.Info{}, s.err }

func TestTransportFacts(t *testing.T) {
	tr := git.Transport{}
	if tr.Name() != "git" || tr.Remote() {
		t.Error("git reads a local checkout")
	}
	if !tr.Applies(sources.Source{Checkout: "c"}) || tr.Applies(sources.Source{URLs: []string{"https://e.test"}}) {
		t.Error("Applies must follow the checkout")
	}
}
