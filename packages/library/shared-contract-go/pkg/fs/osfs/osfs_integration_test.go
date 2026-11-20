//go:build integration

package osfs_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
)

func TestRoundTripOnDisk(t *testing.T) {
	root := t.TempDir()
	fsys := osfs.New()

	file := root + "/a/b/c.txt"
	if err := fsys.WriteFile(file, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	got, err := fsys.ReadFile(file)
	if err != nil || string(got) != "hello" {
		t.Fatalf("ReadFile = %q, %v", got, err)
	}
	if err := fsys.WriteFile(root+"/a/z.txt", nil); err != nil {
		t.Fatal(err)
	}

	entries, err := fsys.ReadDir(root + "/a")
	if err != nil {
		t.Fatal(err)
	}
	want := []fs.Entry{{Name: "b", IsDir: true}, {Name: "z.txt"}}
	if !reflect.DeepEqual(entries, want) {
		t.Fatalf("ReadDir = %+v; want %+v", entries, want)
	}

	info, err := fsys.Stat(file)
	if err != nil || info.Size != 5 || info.IsDir {
		t.Fatalf("Stat = %+v, %v", info, err)
	}
	if err := fsys.MkdirAll(root + "/m/n"); err != nil {
		t.Fatal(err)
	}
	if info, err := fsys.Stat(root + "/m/n"); err != nil || !info.IsDir {
		t.Fatalf("Stat dir = %+v, %v", info, err)
	}
	if err := fsys.Remove(file); err != nil {
		t.Fatal(err)
	}
	if _, err := fsys.Stat(file); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("after remove err = %v", err)
	}
}
