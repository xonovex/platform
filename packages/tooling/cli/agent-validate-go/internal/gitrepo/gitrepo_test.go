package gitrepo_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/gitrepo"
)

func TestRepository(t *testing.T) {
	var calls [][]string
	run := func(dir string, args ...string) (string, error) {
		calls = append(calls, append([]string{dir}, args...))
		return "a.go\r\n\nb/c.go\n", nil
	}
	repo := gitrepo.Repository{Dir: "/repo", Run: run}
	tracked, err := repo.TrackedFiles()
	if err != nil || !reflect.DeepEqual(tracked, []string{"a.go", "b/c.go"}) {
		t.Fatalf("tracked = %v err = %v", tracked, err)
	}
	ignored, err := repo.IgnoredPaths()
	if err != nil || len(ignored) != 2 {
		t.Fatalf("ignored = %v err = %v", ignored, err)
	}
	want := [][]string{
		{"/repo", "ls-files"},
		{"/repo", "ls-files", "--others", "--ignored", "--exclude-standard", "--directory"},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v", calls)
	}
}

func TestFault(t *testing.T) {
	repo := gitrepo.Repository{Run: func(string, ...string) (string, error) { return "", errors.New("not a repository") }}
	for _, call := range []func() ([]string, error){repo.TrackedFiles, repo.IgnoredPaths} {
		if _, err := call(); err == nil || !strings.Contains(err.Error(), "not a repository") || !strings.Contains(err.Error(), "git ls-files") {
			t.Fatalf("err = %v", err)
		}
	}
}
