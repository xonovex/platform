package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveContainedOptionalPathReturnsContainedConfig(t *testing.T) {
	home := t.TempDir()
	config := filepath.Join(home, ".claude")
	if err := os.Mkdir(config, 0700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	resolved, exists, err := ResolveContainedOptionalPath(home, ".claude", "user config path")

	if err != nil {
		t.Fatalf("ResolveContainedOptionalPath() error = %v", err)
	}
	if !exists || resolved != config {
		t.Fatalf("resolved = %q, exists = %v, want %q, true", resolved, exists, config)
	}
}

func TestResolveContainedOptionalPathRejectsEscapingSymlink(t *testing.T) {
	home := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".claude")); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	_, _, err := ResolveContainedOptionalPath(home, ".claude", "user config path")

	if err == nil || !strings.Contains(err.Error(), "outside home") {
		t.Fatalf("ResolveContainedOptionalPath() error = %v, want containment error", err)
	}
}

func TestResolveContainedOptionalPathReportsMissingConfig(t *testing.T) {
	home := t.TempDir()

	resolved, exists, err := ResolveContainedOptionalPath(home, ".claude", "user config path")

	if err != nil || exists || resolved != "" {
		t.Fatalf("resolved = %q, exists = %v, error = %v, want empty, false, nil", resolved, exists, err)
	}
}
