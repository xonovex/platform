//go:build integration

package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.test", "-c", "commit.gpgsign=false"}, args...)...)
	command.Dir = dir
	command.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestBinaryComparesWithABaseRef(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "--quiet", "--initial-branch=main")
	write(t, root, "skill/references/a.md", "# A\n")
	write(t, root, "skill/sources.json", catalog("2099-01-01", ""))
	git(t, root, "add", ".")
	git(t, root, "commit", "--quiet", "-m", "base")

	// Nothing changed since the base.
	if same := run(t, root, "sources", "skill", "--base", "HEAD"); same.code != 0 {
		t.Fatalf("unchanged: %+v", same)
	}

	// A fed file changes without a new review of the source.
	write(t, root, "skill/references/a.md", "# A changed\n")
	edited := run(t, root, "sources", "skill", "--base", "HEAD")
	if edited.code != 1 || !strings.Contains(edited.stdout, "audit.git-file-changed") {
		t.Fatalf("fed file: %+v", edited)
	}

	// A pin changes without a new review.
	write(t, root, "skill/references/a.md", "# A\n")
	write(t, root, "skill/sources.json", catalog("2099-01-01", `,"version":"2.0.0"`))
	pinned := run(t, root, "sources", "skill", "--base", "HEAD")
	if pinned.code != 1 || !strings.Contains(pinned.stdout, "audit.metadata-changed") {
		t.Fatalf("pin: %+v", pinned)
	}

	// A new review date explains the change.
	write(t, root, "skill/sources.json", catalog("2099-02-01", `,"version":"2.0.0"`))
	if reviewed := run(t, root, "sources", "skill", "--base", "HEAD"); reviewed.code != 0 {
		t.Fatalf("reviewed: %+v", reviewed)
	}

	// A ref the repository does not have is a fault.
	if unknown := run(t, root, "sources", "skill", "--base", "no-such-ref"); unknown.code != 2 {
		t.Fatalf("unknown ref: %+v", unknown)
	}
}

func TestBinaryAuditsAGitCheckoutSource(t *testing.T) {
	root := t.TempDir()
	upstream := root + "/upstream"
	git(t, root, "init", "--quiet", "--initial-branch=main", "upstream")
	write(t, root, "upstream/src/cli.go", "package cli\n")
	git(t, upstream, "add", ".")
	git(t, upstream, "commit", "--quiet", "-m", "one")
	write(t, root, "upstream/src/cli.go", "package cli // v2\n")
	git(t, upstream, "commit", "--quiet", "-am", "two")

	out, err := exec.Command("git", "-C", upstream, "rev-parse", "--short=12", "HEAD~1").Output()
	if err != nil {
		t.Fatal(err)
	}
	pin := strings.TrimSpace(string(out))

	write(t, root, "skill/references/cli.md", "# CLI\n")
	write(t, root, "skill/sources.json", `{"version":1,"sources":[{"id":"up","kind":"git","title":"Upstream","checkout":"../upstream","commit":"`+pin+`","watch":[{"path":"src","feeds":["references/cli.md"]}],"reviewedAt":"2099-01-01","feeds":["references/cli.md"]}]}`)
	got := run(t, root, "sources", "skill")
	if got.code != 1 || !strings.Contains(got.stdout, "audit.git-file-changed") {
		t.Fatalf("drift: %+v", got)
	}
}
