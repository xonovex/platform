package tree_test

import (
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

func newTree(root string, files map[string]string) tree.Tree {
	return tree.Tree{FS: memfs.New(files), Root: root}
}

func TestPaths(t *testing.T) {
	rooted := newTree("/repo", nil)
	if got := rooted.Abs("a/b"); got != "/repo/a/b" {
		t.Fatalf("Abs = %s", got)
	}
	if got := rooted.Rel("/repo/a/b"); got != "a/b" {
		t.Fatalf("Rel = %s", got)
	}
	if got := rooted.Rel("/repo"); got != "." {
		t.Fatalf("Rel of the root = %s", got)
	}
	if got := rooted.Rel("/elsewhere/x"); got != "/elsewhere/x" {
		t.Fatalf("a foreign path stays = %s", got)
	}
	for _, root := range []string{"", "."} {
		plain := newTree(root, nil)
		if plain.Abs("./a/../b") != "b" || plain.Rel("a//b") != "a/b" {
			t.Fatalf("root %q: %s %s", root, plain.Abs("./a/../b"), plain.Rel("a//b"))
		}
	}
}

func TestReading(t *testing.T) {
	tr := newTree("/repo", map[string]string{"/repo/a.txt": "text", "/repo/d/b.txt": "b"})
	if text, err := tr.Read("a.txt"); err != nil || text != "text" {
		t.Fatalf("Read = %q %v", text, err)
	}
	if _, err := tr.Read("missing"); err == nil {
		t.Fatal("a missing file is an error")
	}
	if text, ok, err := tr.Optional("a.txt"); err != nil || !ok || text != "text" {
		t.Fatalf("Optional = %q %v %v", text, ok, err)
	}
	if text, ok, err := tr.Optional("missing"); err != nil || ok || text != "" {
		t.Fatalf("a missing file is not an error: %q %v %v", text, ok, err)
	}
	if !tr.IsFile("a.txt") || tr.IsFile("d") || tr.IsFile("missing") {
		t.Fatal("IsFile")
	}
	if !tr.IsDir("d") || tr.IsDir("a.txt") || tr.IsDir("missing") {
		t.Fatal("IsDir")
	}
	if !tr.Exists("a.txt") || !tr.Exists("d") || tr.Exists("missing") {
		t.Fatal("Exists")
	}
}

func TestList(t *testing.T) {
	tr := newTree(".", map[string]string{"d/b.txt": "b", "d/sub/c.txt": "c"})
	entries, err := tr.List("d")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name)
	}
	if !reflect.DeepEqual(names, []string{"b.txt", "sub"}) {
		t.Fatalf("names = %v", names)
	}
	if got, err := tr.List("missing"); err != nil || got != nil {
		t.Fatalf("a missing directory lists nothing: %v %v", got, err)
	}
	if _, err := tr.List("d/b.txt"); err == nil {
		t.Fatal("a file cannot be listed")
	}
}

func TestWalk(t *testing.T) {
	files := map[string]string{
		"a.md": "a", "d/b.md": "b", "d/c.txt": "c", "node_modules/x.md": "x", "vendor/y.md": "y", "d/dist/z.md": "z",
	}
	tr := newTree(".", files)
	isMarkdown := func(p string) bool { return len(p) > 3 && p[len(p)-3:] == ".md" }
	got, err := tr.Walk(".", isMarkdown)
	if err != nil || !reflect.DeepEqual(got, []string{"a.md", "d/b.md", "vendor/y.md"}) {
		t.Fatalf("Walk = %v %v", got, err)
	}
	tr.Skip = []string{"vendor"}
	got, _ = tr.Walk(".", isMarkdown)
	if !reflect.DeepEqual(got, []string{"a.md", "d/b.md"}) {
		t.Fatalf("a skip name is not entered: %v", got)
	}
	if got, _ := tr.Walk("d", nil); !reflect.DeepEqual(got, []string{"d/b.md", "d/c.txt"}) {
		t.Fatalf("a nil match keeps every file: %v", got)
	}
	if got, _ := tr.Walk("a.md", isMarkdown); !reflect.DeepEqual(got, []string{"a.md"}) {
		t.Fatalf("a file path walks to itself: %v", got)
	}
	if got, _ := tr.Walk("a.md", func(string) bool { return false }); len(got) != 0 {
		t.Fatalf("a file that does not match is dropped: %v", got)
	}
	if got, err := tr.Walk("missing", nil); err != nil || len(got) != 0 {
		t.Fatalf("a missing path walks to nothing: %v %v", got, err)
	}
}

func TestWalkAll(t *testing.T) {
	tr := newTree(".", map[string]string{"a/x.md": "x", "b/y.md": "y", "a/z.md": "z"})
	got, err := tr.WalkAll([]string{"b", "a", "a/x.md"}, nil)
	if err != nil || !reflect.DeepEqual(got, []string{"a/x.md", "a/z.md", "b/y.md"}) {
		t.Fatalf("WalkAll sorts and drops repeats: %v %v", got, err)
	}
}
