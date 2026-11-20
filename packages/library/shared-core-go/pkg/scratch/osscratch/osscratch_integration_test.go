//go:build integration

package osscratch_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/scratch"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/scratch/osscratch"
)

func TestMakeAndRemoveARealDirectory(t *testing.T) {
	var port scratch.Scratch = osscratch.New()
	dir, err := port.Make("shared-core-")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(dir) || !strings.HasPrefix(filepath.Base(dir), "shared-core-") {
		t.Fatalf("Make = %q", dir)
	}
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := port.Remove(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("Remove left %q: %v", dir, err)
	}
}
