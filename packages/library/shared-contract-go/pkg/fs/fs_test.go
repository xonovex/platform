package fs_test

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
)

type failingStat struct{ fs.FileSystem }

var errBoom = errors.New("boom")

func (failingStat) Stat(string) (fs.Info, error) { return fs.Info{}, errBoom }

func TestExists(t *testing.T) {
	fsys := memfs.New(map[string]string{"a/b.txt": "x"})
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"file", "a/b.txt", true},
		{"directory", "a", true},
		{"missing", "a/c.txt", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := fs.Exists(fsys, tc.path)
			if err != nil || got != tc.want {
				t.Fatalf("Exists(%q) = %v, %v; want %v", tc.path, got, err, tc.want)
			}
		})
	}
}

func TestExistsReturnsOtherErrors(t *testing.T) {
	if _, err := fs.Exists(failingStat{}, "x"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v; want boom", err)
	}
}
