//go:build integration

package osfiles_test

import (
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/osfiles"
)

func TestWriteFileReplacesTheFileWhole(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "agentic", "trust.json")
	files := osfiles.Files{}
	for _, content := range []string{"first", "second"} {
		if err := files.WriteFile(path, []byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	data, err := files.ReadFile(path)
	if err != nil || string(data) != "second" {
		t.Fatalf("ReadFile = %q, %v", data, err)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("a temporary file stayed behind: %v", entries)
	}
	info, err := files.Stat(path)
	if err != nil || !info.Mode.IsRegular() || info.Mode.Perm()&0o077 != 0 {
		t.Fatalf("Stat = %v, %v", info.Mode, err)
	}
	if info, err := files.Stat(filepath.Dir(path)); err != nil || !info.Mode.IsDir() {
		t.Fatalf("Stat dir = %v, %v", info.Mode, err)
	}
	if _, err := files.Stat(path + ".missing"); !errors.Is(err, iofs.ErrNotExist) {
		t.Fatalf("Stat missing = %v", err)
	}
	if names, err := files.ReadDir(filepath.Dir(path)); err != nil || len(names) != 1 || names[0] != "trust.json" {
		t.Fatalf("ReadDir = %v, %v", names, err)
	}
	if _, err := files.ReadDir(path + ".missing"); !errors.Is(err, iofs.ErrNotExist) {
		t.Fatalf("ReadDir missing = %v", err)
	}
}

func TestPlacesFollowTheXDGVariables(t *testing.T) {
	config, state := t.TempDir(), t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)
	t.Setenv("XDG_STATE_HOME", state)
	places := osfiles.Places{}
	if dir, err := places.UserDir(); err != nil || dir != filepath.Join(config, "agentic") {
		t.Fatalf("UserDir = %s, %v", dir, err)
	}
	if dir, err := places.StateDir(); err != nil || dir != filepath.Join(state, "agentic") {
		t.Fatalf("StateDir = %s, %v", dir, err)
	}
	if places.ManagedDir() == "" {
		t.Fatal("ManagedDir is empty")
	}
	if wd, err := places.Getwd(); err != nil || wd == "" {
		t.Fatalf("Getwd = %s, %v", wd, err)
	}
}

func TestPrompterReadsLines(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin := os.Stdin
	os.Stdin = read
	t.Cleanup(func() { os.Stdin = stdin })
	_, _ = write.WriteString("yes\r\nlast")
	_ = write.Close()
	prompter := osfiles.NewPrompter()
	if prompter.IsTerminal() {
		t.Fatal("a pipe is not a terminal")
	}
	for _, want := range []string{"yes", "last"} {
		if line, err := prompter.ReadLine(); err != nil || line != want {
			t.Fatalf("ReadLine = %q, %v", line, err)
		}
	}
	if _, err := prompter.ReadLine(); err == nil {
		t.Fatal("the end of the input must surface")
	}
}
