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

func TestResolveHomeDirUsesExplicitDirectory(t *testing.T) {
	home := t.TempDir()

	resolved, err := ResolveHomeDir(home)

	if err != nil || resolved != home {
		t.Fatalf("ResolveHomeDir() = %q, %v, want %q, nil", resolved, err, home)
	}
}

func TestResolveDirectoryRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if _, err := ResolveDirectory(path, "fixture"); err == nil {
		t.Fatal("ResolveDirectory() error = nil, want file rejection")
	}
}

func TestResolveExistingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	resolved, err := ResolveExistingPath(path, "fixture")

	if err != nil || resolved != path {
		t.Fatalf("ResolveExistingPath() = %q, %v, want %q, nil", resolved, err, path)
	}
	if _, err := ResolveExistingPath(path+"-missing", "fixture"); err == nil {
		t.Fatal("ResolveExistingPath() missing error = nil, want error")
	}
}
