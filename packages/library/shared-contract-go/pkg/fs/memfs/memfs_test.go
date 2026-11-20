package memfs_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
)

func TestReadFile(t *testing.T) {
	fsys := memfs.New(map[string]string{"a/b.txt": "hello", "./c.txt": "c"})
	got, err := fsys.ReadFile("a/b.txt")
	if err != nil || string(got) != "hello" {
		t.Fatalf("ReadFile = %q, %v", got, err)
	}
	if got, err := fsys.ReadFile("c.txt"); err != nil || string(got) != "c" {
		t.Fatalf("ReadFile with leading dot = %q, %v", got, err)
	}
	if _, err := fsys.ReadFile("missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing err = %v", err)
	}
	if _, err := fsys.ReadFile("a"); err == nil {
		t.Fatal("reading a directory must fail")
	}
}

func TestReadFileReturnsCopy(t *testing.T) {
	fsys := memfs.New(map[string]string{"f": "abc"})
	first, _ := fsys.ReadFile("f")
	first[0] = 'z'
	second, _ := fsys.ReadFile("f")
	if string(second) != "abc" {
		t.Fatalf("stored content changed: %q", second)
	}
}

func TestWriteFile(t *testing.T) {
	fsys := memfs.New(nil)
	if err := fsys.WriteFile("x/y/z.txt", []byte("1")); err != nil {
		t.Fatal(err)
	}
	info, err := fsys.Stat("x/y")
	if err != nil || !info.IsDir {
		t.Fatalf("parent dir = %+v, %v", info, err)
	}
	if err := fsys.WriteFile("x/y", []byte("2")); err == nil {
		t.Fatal("writing over a directory must fail")
	}
}

func TestReadDir(t *testing.T) {
	fsys := memfs.New(map[string]string{"r/a.txt": "", "r/sub/b.txt": "", "r/sub/deep/c.txt": "", "top.txt": ""})
	got, err := fsys.ReadDir("r")
	if err != nil {
		t.Fatal(err)
	}
	want := []fs.Entry{{Name: "a.txt"}, {Name: "sub", IsDir: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadDir(r) = %+v; want %+v", got, want)
	}
	root, err := fsys.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	if want := []fs.Entry{{Name: "r", IsDir: true}, {Name: "top.txt"}}; !reflect.DeepEqual(root, want) {
		t.Fatalf("ReadDir(.) = %+v; want %+v", root, want)
	}
	if _, err := fsys.ReadDir("top.txt"); !errors.Is(err, fs.ErrNotDir) {
		t.Fatalf("file err = %v", err)
	}
	if _, err := fsys.ReadDir("nope"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing err = %v", err)
	}
}

func TestAbsoluteRoot(t *testing.T) {
	fsys := memfs.New(map[string]string{"/repo/a.txt": "x"})
	got, err := fsys.ReadDir("/")
	if err != nil || len(got) != 1 || got[0].Name != "repo" || !got[0].IsDir {
		t.Fatalf("ReadDir(/) = %+v, %v", got, err)
	}
}

func TestStat(t *testing.T) {
	fsys := memfs.New(map[string]string{"d/f": "abcd"})
	if info, err := fsys.Stat("d/f"); err != nil || info.Size != 4 || info.IsDir {
		t.Fatalf("file stat = %+v, %v", info, err)
	}
	if info, err := fsys.Stat("d"); err != nil || !info.IsDir {
		t.Fatalf("dir stat = %+v, %v", info, err)
	}
	if _, err := fsys.Stat("nope"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing err = %v", err)
	}
}

func TestMkdirAll(t *testing.T) {
	fsys := memfs.New(map[string]string{"f": ""})
	if err := fsys.MkdirAll("a/b/c"); err != nil {
		t.Fatal(err)
	}
	if info, err := fsys.Stat("a/b"); err != nil || !info.IsDir {
		t.Fatalf("parent = %+v, %v", info, err)
	}
	if err := fsys.MkdirAll("f"); !errors.Is(err, fs.ErrNotDir) {
		t.Fatalf("mkdir over file err = %v", err)
	}
}

func TestRemove(t *testing.T) {
	fsys := memfs.New(map[string]string{"d/f": "", "d/sub/g": ""})
	if err := fsys.Remove("d"); err == nil {
		t.Fatal("removing a directory with files must fail")
	}
	if err := fsys.Remove("d/f"); err != nil {
		t.Fatal(err)
	}
	if err := fsys.Remove("d"); err == nil {
		t.Fatal("removing a directory with a subdirectory must fail")
	}
	if err := fsys.Remove("d/sub/g"); err != nil {
		t.Fatal(err)
	}
	if err := fsys.Remove("d/sub"); err != nil {
		t.Fatal(err)
	}
	if err := fsys.Remove("d"); err != nil {
		t.Fatal(err)
	}
	if err := fsys.Remove("d"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("second remove err = %v", err)
	}
}
