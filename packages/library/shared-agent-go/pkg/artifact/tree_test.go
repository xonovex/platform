package artifact

import (
	"strings"
	"testing"
)

func file(path, content string) File { return File{Path: path, Content: []byte(content)} }

func TestNewTreeSortsAndGets(t *testing.T) {
	tree, err := NewTree([]File{file("b", "2"), file("a/x", "1"), file("a-b", "3")})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, f := range tree {
		paths = append(paths, f.Path)
	}
	if got := strings.Join(paths, ","); got != "a-b,a/x,b" {
		t.Errorf("paths = %s", got)
	}
	if f, ok := tree.Get("a/x"); !ok || string(f.Content) != "1" {
		t.Errorf("Get(a/x) = %v, %v", f, ok)
	}
	if _, ok := tree.Get("missing"); ok {
		t.Error("Get found a missing path")
	}
}

func TestNewTreeFailsClosed(t *testing.T) {
	for name, files := range map[string][]File{
		"escape":         {file("../x", "")},
		"duplicate":      {file("a", ""), file("a", "")},
		"case collision": {file("a", ""), file("A", "")},
		"file and dir":   {file("a", ""), file("a-b", ""), file("a/b", "")},
	} {
		if _, err := NewTree(files); err == nil {
			t.Errorf("%s: NewTree passed", name)
		}
	}
}

func TestTreeUnder(t *testing.T) {
	tree, _ := NewTree([]File{file("a/x", "1"), file("a/y/z", "2"), file("ab/q", "3")})
	under := tree.Under("a")
	if len(under) != 2 || under[0].Path != "x" || under[1].Path != "y/z" {
		t.Errorf("Under(a) = %v", under)
	}
}

func TestDigestCoversPathModeAndContent(t *testing.T) {
	base, _ := NewTree([]File{file("a", "1"), file("b", "2")})
	same, _ := NewTree([]File{file("b", "2"), file("a", "1")})
	if base.Digest() != same.Digest() {
		t.Error("the digest depends on the order of files")
	}
	for name, files := range map[string][]File{
		"content": {file("a", "1"), file("b", "3")},
		"path":    {file("a", "1"), file("c", "2")},
		"mode":    {file("a", "1"), {Path: "b", Content: []byte("2"), Executable: true}},
		"extra":   {file("a", "1"), file("b", "2"), file("c", "")},
	} {
		other, _ := NewTree(files)
		if other.Digest() == base.Digest() {
			t.Errorf("a change of %s kept the digest", name)
		}
	}
	if !ValidDigest(base.Digest()) {
		t.Errorf("Digest() %q is not a valid digest", base.Digest())
	}
}

func TestValidDigest(t *testing.T) {
	good := "sha256:" + strings.Repeat("a", 64)
	for digest, want := range map[string]bool{
		good:                                true,
		"sha256:" + strings.Repeat("A", 64): false,
		"sha256:" + strings.Repeat("a", 63): false,
		"sha256:" + strings.Repeat("g", 64): false,
		"md5:" + strings.Repeat("a", 64):    false,
		"":                                  false,
	} {
		if got := ValidDigest(digest); got != want {
			t.Errorf("ValidDigest(%q) = %v", digest, got)
		}
	}
}
