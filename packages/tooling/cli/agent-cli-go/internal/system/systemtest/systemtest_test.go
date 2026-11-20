package systemtest

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
)

func TestNewDescribesAUser(t *testing.T) {
	f := New()

	home, err := f.UserHomeDir()
	wd, _ := f.Getwd()
	if err != nil || home != "/home/user" || wd != "/work" || f.Getuid() != 1000 || f.Getgid() != 1000 || f.NumCPU() != 4 {
		t.Errorf("identity = %q %v %q %d %d %d", home, err, wd, f.Getuid(), f.Getgid(), f.NumCPU())
	}
	f.HomeErr = errors.New("no home")
	if _, err := f.UserHomeDir(); err == nil {
		t.Error("UserHomeDir() error = nil, want HomeErr")
	}
}

func TestEnvironment(t *testing.T) {
	f := New()
	f.Env["B"] = "2"
	f.Env["A"] = "1"

	if f.Getenv("A") != "1" || f.Getenv("C") != "" {
		t.Errorf("Getenv = %q %q", f.Getenv("A"), f.Getenv("C"))
	}
	if got := f.Environ(); !slices.Equal(got, []string{"A=1", "B=2"}) {
		t.Errorf("Environ() = %v, want sorted entries", got)
	}
}

func TestFilesAndDirectories(t *testing.T) {
	f := New().AddFile("/a/b/file", "data", 0o644).AddDir("/c")

	info, err := f.Stat("/a/b/file")
	if err != nil || info.Name() != "file" || info.Size() != 4 || info.IsDir() || info.Mode() != 0o644 || info.Sys() != nil || !info.ModTime().IsZero() {
		t.Fatalf("Stat(file) = %v, %v", info, err)
	}
	dir, err := f.Stat("/a/b")
	if err != nil || !dir.IsDir() || dir.Mode() != fs.ModeDir|0o755 {
		t.Fatalf("Stat(dir) = %v, %v", dir, err)
	}
	if content, ok := f.File("/a/b/file"); !ok || content != "data" {
		t.Errorf("File() = %q, %v", content, ok)
	}
	if _, ok := f.File("/a/b"); ok {
		t.Error("File(dir) reports a file")
	}
	if !f.Exists("/c") || f.Exists("/d") {
		t.Error("Exists() does not follow the tree")
	}
	data, err := f.ReadFile("/a/b/file")
	if err != nil || string(data) != "data" {
		t.Errorf("ReadFile() = %q, %v", data, err)
	}
	if _, err := f.ReadFile("/a/b"); err == nil {
		t.Error("ReadFile(dir) error = nil")
	}
	if _, err := f.ReadFile("/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadFile(missing) error = %v, want ErrNotExist", err)
	}
	if _, err := f.Stat("/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(missing) error = %v, want ErrNotExist", err)
	}
}

func TestSymlinks(t *testing.T) {
	f := New().
		AddFile("/real/dir/file", "x", 0o600).
		AddLink("/abs", "/real/dir").
		AddLink("/real/rel", "dir/file").
		AddLink("/chain", "/abs").
		AddLink("/loop", "/loop")

	tests := map[string]string{
		"/abs/file":  "/real/dir/file",
		"/real/rel":  "/real/dir/file",
		"/chain":     "/real/dir",
		"/real/dir/": "/real/dir",
	}
	for in, want := range tests {
		if got, err := f.EvalSymlinks(in); err != nil || got != want {
			t.Errorf("EvalSymlinks(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if info, err := f.Stat("/abs/file"); err != nil || info.Name() != "file" {
		t.Errorf("Stat(through link) = %v, %v", info, err)
	}
	if _, err := f.EvalSymlinks("/loop"); err == nil || !strings.Contains(err.Error(), "too many levels") {
		t.Errorf("EvalSymlinks(loop) error = %v", err)
	}
	if _, err := f.EvalSymlinks("/real/dir/file/below"); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("EvalSymlinks(below a file) error = %v", err)
	}
	if _, err := f.EvalSymlinks("relative"); err == nil || !strings.Contains(err.Error(), "relative path") {
		t.Errorf("EvalSymlinks(relative) error = %v", err)
	}
	if _, err := f.EvalSymlinks("/abs/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("EvalSymlinks(missing) error = %v", err)
	}
	if _, err := f.EvalSymlinks("/dangling"); err == nil {
		t.Error("EvalSymlinks(dangling) error = nil")
	}
	f.AddLink("/broken", "/nowhere")
	if _, err := f.Stat("/broken"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(broken link) error = %v", err)
	}
}

func TestWriteCreateAndRemove(t *testing.T) {
	f := New().AddDir("/data").AddLink("/link", "/data")

	if err := f.WriteFile("/link/out", []byte("v1"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if content, ok := f.File("/data/out"); !ok || content != "v1" {
		t.Errorf("File(/data/out) = %q, %v; want the write through the link", content, ok)
	}
	if err := f.WriteFile("/missing/out", nil, 0o600); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("WriteFile(missing parent) error = %v", err)
	}
	f.AddFile("/data/file", "", 0o600)
	if err := f.WriteFile("/data/file/out", nil, 0o600); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("WriteFile(file parent) error = %v", err)
	}

	first, _ := f.CreateTemp("script-*.sh")
	second, _ := f.CreateTemp("plain")
	if first != "/tmp/script-1.sh" || second != "/tmp/plain" {
		t.Errorf("CreateTemp() = %q %q", first, second)
	}
	if info, err := f.Stat(first); err != nil || info.Mode() != 0o600 {
		t.Errorf("Stat(temp) = %v, %v; want a private file", info, err)
	}

	if err := f.Remove(first); err != nil || f.Exists(first) {
		t.Errorf("Remove() error = %v, exists = %v", err, f.Exists(first))
	}
	if err := f.Remove(first); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Remove(missing) error = %v", err)
	}
	if err := f.MkdirAll("/x/y/z", 0o755); err != nil || !f.Exists("/x/y") {
		t.Errorf("MkdirAll() error = %v", err)
	}
}

func TestLookPath(t *testing.T) {
	f := New().
		AddExecutable("/usr/bin/tool").
		AddExecutable("/opt/bin/tool").
		AddFile("/usr/bin/data", "", 0o644).
		AddDir("/usr/bin/dir")
	f.Env["PATH"] = ":/missing:/usr/bin:/opt/bin"

	if got, err := f.LookPath("tool"); err != nil || got != "/usr/bin/tool" {
		t.Errorf("LookPath(tool) = %q, %v; want the first PATH entry", got, err)
	}
	for _, name := range []string{"data", "dir", "absent"} {
		if _, err := f.LookPath(name); err == nil {
			t.Errorf("LookPath(%q) error = nil", name)
		}
	}
	if got, err := f.LookPath("/opt/bin/tool"); err != nil || got != "/opt/bin/tool" {
		t.Errorf("LookPath(path) = %q, %v", got, err)
	}
	if _, err := f.LookPath("/usr/bin/data"); err == nil {
		t.Error("LookPath(non-executable path) error = nil")
	}
}

func TestProcesses(t *testing.T) {
	f := New()
	f.Outcomes[CommandLine("git", "status")] = Outcome{Stdout: " generic \n"}
	f.Outcomes[InDir("/repo", CommandLine("git", "status"))] = Outcome{Stdout: "in repo"}
	f.Outcomes[CommandLine("make")] = Outcome{ExitCode: 2}
	f.Outcomes[CommandLine("broken")] = Outcome{Err: errors.New("failed")}

	if out, err := f.Capture("git", []string{"status"}, "/repo"); err != nil || out != "in repo" {
		t.Errorf("Capture(in dir) = %q, %v", out, err)
	}
	if out, err := f.Capture("git", []string{"status"}, "/other"); err != nil || out != "generic" {
		t.Errorf("Capture(generic) = %q, %v", out, err)
	}
	if code, err := f.Run("make", nil, []string{"A=1"}, "/w"); err != nil || code != 2 {
		t.Errorf("Run() = %d, %v", code, err)
	}
	if code, err := f.Run("broken", nil, nil, ""); err == nil || code != 1 {
		t.Errorf("Run(broken) = %d, %v", code, err)
	}
	if _, err := f.Capture("broken", nil, ""); err == nil {
		t.Error("Capture(broken) error = nil")
	}
	if err := f.Stream("make", nil, ""); err != nil {
		t.Errorf("Stream() error = %v", err)
	}
	if err := f.Stream("unknown", []string{"x"}, "/d"); err == nil || !strings.Contains(err.Error(), `no outcome for "unknown x" in "/d"`) {
		t.Errorf("Stream(unknown) error = %v", err)
	}

	kinds := make([]string, len(f.Calls))
	for i, call := range f.Calls {
		kinds[i] = call.Kind
	}
	if !slices.Equal(kinds, []string{"capture", "capture", "run", "run", "capture", "stream", "stream"}) {
		t.Errorf("Calls kinds = %v", kinds)
	}
	if call := f.Calls[2]; call.Line() != "make" || call.Dir != "/w" || !slices.Equal(call.Env, []string{"A=1"}) {
		t.Errorf("Calls[2] = %+v", call)
	}
}

func TestLstatDoesNotFollowAFinalLink(t *testing.T) {
	f := New().AddFile("/data/file", "x", 0o644).AddLink("/link", "/data/file").AddLink("/dir", "/data")
	info, err := f.Lstat("/link")
	if err != nil || info.Mode()&fs.ModeSymlink == 0 || info.Name() != "link" {
		t.Errorf("Lstat(link) = %v, %v; want the link itself", info, err)
	}
	if info, err := f.Lstat("/dir/file"); err != nil || !info.Mode().IsRegular() {
		t.Errorf("Lstat(dir/file) = %v, %v; want the file behind the link directory", info, err)
	}
	if _, err := f.Lstat("/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Lstat(missing) error = %v", err)
	}
	if _, err := f.Lstat("/missing/child"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Lstat(missing/child) error = %v", err)
	}
}

func TestReadDirListsDirectChildrenSortedByName(t *testing.T) {
	f := New().AddFile("/d/b", "", 0o644).AddFile("/d/a/inner", "", 0o644).AddLink("/d/c", "/d/b").AddFile("/other/x", "", 0o644).AddLink("/alias", "/d")
	entries, err := f.ReadDir("/alias")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, entry := range entries {
		info, _ := entry.Info()
		got = append(got, entry.Name()+":"+strings.Join([]string{boolText(entry.IsDir()), entry.Type().String(), info.Name()}, ":"))
	}
	if want := "a:true:d---------:a,b:false:----------:b,c:false:L---------:c"; strings.Join(got, ",") != want {
		t.Errorf("entries = %v, want %s", got, want)
	}
	if _, err := f.ReadDir("/d/b"); err == nil {
		t.Error("ReadDir(file) error = nil")
	}
	if _, err := f.ReadDir("/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("ReadDir(missing) error = %v", err)
	}
	if entries, err := f.ReadDir("/"); err != nil || len(entries) != 3 {
		t.Errorf("ReadDir(/) = %d entries, %v", len(entries), err)
	}
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func TestMkdirTempRenameAndRemoveAll(t *testing.T) {
	f := New()
	first, _ := f.MkdirTemp("overlay-*")
	second, _ := f.MkdirTemp("plain")
	if first != "/tmp/overlay-1" || second != "/tmp/plain2" {
		t.Errorf("MkdirTemp() = %q %q", first, second)
	}
	if info, err := f.Stat(first); err != nil || !info.IsDir() {
		t.Errorf("Stat(temp) = %v, %v", info, err)
	}

	f.AddFile(first+"/a/b", "content", 0o644).AddFile(first+"-sibling", "x", 0o644).AddDir("/dest")
	if err := f.Rename(first, "/dest/moved"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if content, ok := f.File("/dest/moved/a/b"); !ok || content != "content" || f.Exists(first) || !f.Exists(first+"-sibling") {
		t.Errorf("Rename() moved %q, %v; old exists = %v", content, ok, f.Exists(first))
	}
	if err := f.Rename("/missing", "/dest/x"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Rename(missing) error = %v", err)
	}
	if err := f.Rename("/dest/moved", "/nowhere/x"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Rename(missing parent) error = %v", err)
	}
	f.AddFile("/file", "", 0o644)
	if err := f.Rename("/dest/moved", "/file/x"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Rename(file parent) error = %v", err)
	}

	if err := f.RemoveAll("/dest"); err != nil || f.Exists("/dest") || f.Exists("/dest/moved/a/b") || !f.Exists(first+"-sibling") {
		t.Errorf("RemoveAll() error = %v", err)
	}
	if err := f.RemoveAll("/dest"); err != nil {
		t.Errorf("RemoveAll(missing) error = %v", err)
	}
}
