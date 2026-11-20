package hostlaunch

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

func TestResolveContainedOptionalPathReturnsContainedConfig(t *testing.T) {
	fsys := oshosttest.New().AddDir("/home/user/.claude")

	resolved, exists, err := ResolveContainedOptionalPath(fsys, "/home/user", ".claude", "user config path")

	if err != nil {
		t.Fatalf("ResolveContainedOptionalPath() error = %v", err)
	}
	if !exists || resolved != "/home/user/.claude" {
		t.Fatalf("resolved = %q, exists = %v, want /home/user/.claude, true", resolved, exists)
	}
}

func TestResolveContainedOptionalPathRejectsEscapingSymlink(t *testing.T) {
	fsys := oshosttest.New().AddDir("/home/user").AddDir("/outside").AddLink("/home/user/.claude", "/outside")

	_, _, err := ResolveContainedOptionalPath(fsys, "/home/user", ".claude", "user config path")

	if err == nil || !strings.Contains(err.Error(), "outside home") {
		t.Fatalf("ResolveContainedOptionalPath() error = %v, want containment error", err)
	}
}

func TestResolveContainedOptionalPathReportsMissingConfig(t *testing.T) {
	fsys := oshosttest.New().AddDir("/home/user")

	resolved, exists, err := ResolveContainedOptionalPath(fsys, "/home/user", ".claude", "user config path")

	if err != nil || exists || resolved != "" {
		t.Fatalf("resolved = %q, exists = %v, error = %v, want empty, false, nil", resolved, exists, err)
	}
}

func TestResolveContainedOptionalPathRejectsInvalidRelativePaths(t *testing.T) {
	fsys := oshosttest.New().AddDir("/home/user")
	tests := []struct {
		name     string
		relative string
		phrase   string
	}{
		{"absolute", "/etc/passwd", "must be relative"},
		{"parent", "..", "escapes the home directory"},
		{"traversal", "../other", "escapes the home directory"},
		{"home itself", ".", "escapes the home directory"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := ResolveContainedOptionalPath(fsys, "/home/user", test.relative, "user config path")
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("ResolveContainedOptionalPath(%q) error = %v, want %q", test.relative, err, test.phrase)
			}
		})
	}
}

func TestResolveContainedOptionalPathRequiresTheHomeDirectory(t *testing.T) {
	if _, _, err := ResolveContainedOptionalPath(oshosttest.New(), "/home/missing", ".claude", "user config path"); err == nil {
		t.Fatal("ResolveContainedOptionalPath() error = nil, want missing home error")
	}
}

func TestResolveContainedOptionalPathReportsBrokenResolution(t *testing.T) {
	fsys := oshosttest.New().AddFile("/home/user/file", "x", 0o600).AddLink("/home/user/.claude", "file/inside")

	_, _, err := ResolveContainedOptionalPath(fsys, "/home/user", ".claude", "user config path")

	if err == nil || !strings.Contains(err.Error(), "resolve user config path") {
		t.Fatalf("ResolveContainedOptionalPath() error = %v, want resolution error", err)
	}
}

func TestResolveDirectoryResolvesSymlinks(t *testing.T) {
	fsys := oshosttest.New().AddDir("/data/real").AddLink("/work", "/data/real")

	resolved, err := ResolveDirectory(fsys, "/work/", "work directory")

	if err != nil || resolved != "/data/real" {
		t.Fatalf("ResolveDirectory() = %q, %v, want /data/real, nil", resolved, err)
	}
}

func TestResolveDirectoryRejectsFileMissingAndRelativePaths(t *testing.T) {
	fsys := oshosttest.New().AddFile("/work/file", "content", 0o600)
	tests := []struct {
		path   string
		phrase string
	}{
		{"/work/file", "is not a directory"},
		{"/work/missing", "inspect fixture"},
		{"work", "must be an absolute path"},
	}
	for _, test := range tests {
		if _, err := ResolveDirectory(fsys, test.path, "fixture"); err == nil || !strings.Contains(err.Error(), test.phrase) {
			t.Errorf("ResolveDirectory(%q) error = %v, want %q", test.path, err, test.phrase)
		}
	}
}

func TestResolveExistingPath(t *testing.T) {
	fsys := oshosttest.New().AddFile("/work/file", "content", 0o600)

	resolved, err := ResolveExistingPath(fsys, "/work/./file", "fixture")

	if err != nil || resolved != "/work/file" {
		t.Fatalf("ResolveExistingPath() = %q, %v, want /work/file, nil", resolved, err)
	}
	if _, err := ResolveExistingPath(fsys, "/work/file-missing", "fixture"); err == nil {
		t.Fatal("ResolveExistingPath() missing error = nil, want error")
	}
	if _, err := ResolveExistingPath(fsys, "file", "fixture"); err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("ResolveExistingPath(relative) error = %v, want absolute path error", err)
	}
}

// unresolvable is a FileSystem whose symbolic links never resolve.
type unresolvable struct{ FileSystem }

func (unresolvable) EvalSymlinks(string) (string, error) { return "", errors.New("loop") }

func TestResolveDirectoryReportsAnUnresolvableDirectory(t *testing.T) {
	fsys := unresolvable{oshosttest.New().AddDir("/work")}

	_, err := ResolveDirectory(fsys, "/work", "work directory")

	if err == nil || !strings.Contains(err.Error(), "resolve canonical work directory") {
		t.Fatalf("ResolveDirectory() error = %v, want resolution error", err)
	}
}
