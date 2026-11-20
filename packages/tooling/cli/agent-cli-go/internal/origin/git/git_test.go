package git

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestAcceptsRepositoryURLs(t *testing.T) {
	origin := New(systemtest.New())
	for _, spec := range []string{
		"https://github.com/org/repo.git", "https://github.com/org/repo#v1.2.3", "ssh://git@host/org/repo.git",
		"file:///srv/repo", "git@github.com:org/repo.git", "git@github.com:org/repo.git#feature/x",
	} {
		if !origin.Accepts(spec) {
			t.Errorf("Accepts(%q) = false", spec)
		}
	}
	for _, spec := range []string{
		"", "http://host/repo", "ext::sh -c id", "-oProxyCommand=x", "--upload-pack=x", "/local/dir", "git@host:", "https://host/repo#", "https://host/repo#-x", "https://host/repo#a b", "ftp://host/x",
	} {
		if origin.Accepts(spec) {
			t.Errorf("Accepts(%q) = true", spec)
		}
	}
}

func cloneLine(url, ref string) string {
	args := []string{"-c", "protocol.ext.allow=never", "clone", "--quiet", "--depth", "1", "--no-tags"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, "--", url, "/tmp/agent-cli-origin-1/checkout")
	return systemtest.CommandLine("git", args...)
}

func TestFetchClonesShallowlyAndResolvesTheCommit(t *testing.T) {
	host := systemtest.New()
	host.Outcomes[cloneLine("https://h/r.git", "v1")] = systemtest.Outcome{}
	host.Outcomes[systemtest.InDir("/tmp/agent-cli-origin-1/checkout", "git rev-parse HEAD")] = systemtest.Outcome{Stdout: "abc123\n"}
	checkout, err := New(host).Fetch("https://h/r.git#v1")
	if err != nil {
		t.Fatal(err)
	}
	if checkout.Dir != "/tmp/agent-cli-origin-1/checkout" || checkout.Source != "https://h/r.git#v1" || checkout.Resolved != "abc123" {
		t.Errorf("checkout = %+v", checkout)
	}
	if !host.Exists("/tmp/agent-cli-origin-1") {
		t.Error("the clone directory is missing")
	}
	checkout.Release()
	if host.Exists("/tmp/agent-cli-origin-1") {
		t.Error("Release left the clone directory")
	}
}

func TestFetchWithoutARefClonesTheDefaultBranch(t *testing.T) {
	host := systemtest.New()
	host.Outcomes[cloneLine("git@h:o/r.git", "")] = systemtest.Outcome{}
	host.Outcomes[systemtest.InDir("/tmp/agent-cli-origin-1/checkout", "git rev-parse HEAD")] = systemtest.Outcome{Stdout: "def"}
	checkout, err := New(host).Fetch("git@h:o/r.git")
	if err != nil || checkout.Source != "git@h:o/r.git" || strings.Contains(strings.Join(host.Calls[0].Args, " "), "--branch") {
		t.Errorf("checkout = %+v, %v, calls = %v", checkout, err, host.Calls)
	}
}

func TestFetchFailuresReleaseTheDirectory(t *testing.T) {
	failure := errors.New("boom")
	host := systemtest.New()
	host.Outcomes[cloneLine("https://h/r.git", "")] = systemtest.Outcome{Err: failure}
	if _, err := New(host).Fetch("https://h/r.git"); !errors.Is(err, failure) || host.Exists("/tmp/agent-cli-origin-1") {
		t.Errorf("clone failure: %v, exists = %v", err, host.Exists("/tmp/agent-cli-origin-1"))
	}

	host = systemtest.New()
	host.Outcomes[cloneLine("https://h/r.git", "")] = systemtest.Outcome{}
	host.Outcomes[systemtest.InDir("/tmp/agent-cli-origin-1/checkout", "git rev-parse HEAD")] = systemtest.Outcome{Err: failure}
	if _, err := New(host).Fetch("https://h/r.git"); !errors.Is(err, failure) || host.Exists("/tmp/agent-cli-origin-1") {
		t.Errorf("rev-parse failure: %v, exists = %v", err, host.Exists("/tmp/agent-cli-origin-1"))
	}

	if _, err := New(systemtest.New()).Fetch("/not/a/url"); err == nil {
		t.Error("a non-URL passed")
	}
}

type noTemp struct{ *systemtest.Fake }

func (noTemp) MkdirTemp(string) (string, error) { return "", errors.New("no space") }

func TestFetchFailsWithoutATemporaryDirectory(t *testing.T) {
	if _, err := New(noTemp{systemtest.New()}).Fetch("https://h/r.git"); err == nil || !strings.Contains(err.Error(), "no space") {
		t.Errorf("error = %v", err)
	}
}
