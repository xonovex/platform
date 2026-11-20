package discovery_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/discovery"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
)

func tree() *memfs.FileSystem {
	return memfs.New(map[string]string{
		"repo/a/SKILL.md":                   "",
		"repo/a/references/r.md":            "",
		"repo/b/SKILL.md":                   "",
		"repo/b/node_modules/x/SKILL.md":    "",
		"repo/b/dist/SKILL.md":              "",
		"repo/b/eval-results/run/SKILL.md":  "",
		"repo/.git/objects/SKILL.md":        "",
		"repo/a/deeper/node_modules/y/z.md": "",
		"repo/a/deeper/keep.md":             "",
	})
}

func TestWalkSkipsDefaultDirectories(t *testing.T) {
	got, err := discovery.Walk(tree(), "repo", discovery.Options{Skip: discovery.DefaultSkip()})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"repo/a/SKILL.md", "repo/a/deeper/keep.md", "repo/a/references/r.md", "repo/b/SKILL.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Walk = %v; want %v", got, want)
	}
}

func TestWalkWithoutSkipEntersEverything(t *testing.T) {
	got, err := discovery.Walk(tree(), "repo", discovery.Options{Match: discovery.Named("SKILL.md")})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 6 {
		t.Fatalf("Walk found %d SKILL.md files: %v", len(got), got)
	}
}

func TestMatchers(t *testing.T) {
	got, err := discovery.Walk(tree(), "repo", discovery.Options{Skip: discovery.DefaultSkip(), Match: discovery.WithExtension(".md")})
	if err != nil || len(got) != 4 {
		t.Fatalf("Walk = %v, %v", got, err)
	}
	if discovery.Named("x.md")("a/b/y.md") || !discovery.Named("y.md")("a/b/y.md") {
		t.Fatal("Named matches the base name only")
	}
}

func TestWalkMissingRoot(t *testing.T) {
	if _, err := discovery.Walk(tree(), "nope", discovery.Options{}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

type failingNested struct{ fs.FileSystem }

func (f failingNested) ReadDir(p string) ([]fs.Entry, error) {
	if p == "repo/b" {
		return nil, errors.New("denied")
	}
	return f.FileSystem.ReadDir(p)
}

func TestWalkPropagatesNestedError(t *testing.T) {
	if _, err := discovery.Walk(failingNested{tree()}, "repo", discovery.Options{}); err == nil {
		t.Fatal("a failing nested directory must fail the walk")
	}
}
