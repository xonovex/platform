package catalogs_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/catalogs"
)

func TestDiscover(t *testing.T) {
	fsys := memfs.New(map[string]string{
		"root/a/sources.json":                "{}",
		"root/a/sources.d/one.json":          "{}",
		"root/b/sources.d/two.json":          "{}",
		"root/b/sources.d/notes.md":          "x",
		"root/c/data.json":                   "{}",
		"root/d/node_modules/x/sources.json": "{}",
		"root/e/dist/sources.json":           "{}",
		"root/sources.json":                  "{}",
		"root/f/other/sources.json":          "{}",
	})
	got, err := catalogs.Discover(fsys, "root")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"root", "root/a", "root/b", "root/f/other"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Discover = %v; want %v", got, want)
	}
}

func TestDiscoverFindsNothing(t *testing.T) {
	got, err := catalogs.Discover(memfs.New(map[string]string{"root/a.txt": "x"}), "root")
	if err != nil || len(got) != 0 {
		t.Fatalf("Discover = %v, %v", got, err)
	}
}

func TestDiscoverReportsAnUnreadableRoot(t *testing.T) {
	_, err := catalogs.Discover(memfs.New(nil), "nope")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}
