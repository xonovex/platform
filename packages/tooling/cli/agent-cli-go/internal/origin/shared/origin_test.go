package shared

import (
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

type namedOrigin struct{ accepts bool }

func (o namedOrigin) Accepts(string) bool            { return o.accepts }
func (o namedOrigin) Fetch(string) (Checkout, error) { return Checkout{}, nil }

func TestRegistrySelectsTheOnlyOriginThatAccepts(t *testing.T) {
	registry := Registry{"a": namedOrigin{false}, "b": namedOrigin{true}}
	origin, err := registry.Select("x")
	if err != nil || origin != registry["b"] {
		t.Errorf("Select() = %v, %v", origin, err)
	}
}

func TestRegistrySelectFailsClosed(t *testing.T) {
	if _, err := (Registry{"a": namedOrigin{false}}).Select("x"); err == nil || !strings.Contains(err.Error(), "neither an existing directory") {
		t.Errorf("no origin: %v", err)
	}
	if _, err := (Registry{"b": namedOrigin{true}, "a": namedOrigin{true}}).Select("x"); err == nil || !strings.Contains(err.Error(), "a, b all accept") {
		t.Errorf("two origins: %v", err)
	}
}

func TestReadTreeReadsRegularFilesAndSkipsGit(t *testing.T) {
	host := systemtest.New().
		AddFile("/p/skills/a/SKILL.md", "skill", 0o644).
		AddFile("/p/skills/a/run.sh", "#!/bin/sh", 0o755).
		AddFile("/p/.git/config", "git", 0o644).
		AddFile("/p/sub/.git", "nested", 0o644).
		AddDir("/p/empty")
	tree, err := ReadTree(host, "/p")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, file := range tree {
		got = append(got, file.Path)
		if file.Path == "skills/a/run.sh" && !file.Executable {
			t.Error("the executable bit was lost")
		}
	}
	if strings.Join(got, ",") != "skills/a/SKILL.md,skills/a/run.sh,sub/.git" {
		t.Errorf("paths = %v", got)
	}
}

func TestReadTreeFailsClosed(t *testing.T) {
	link := systemtest.New().AddFile("/p/a", "x", 0o644).AddLink("/p/b", "/p/a")
	if _, err := ReadTree(link, "/p"); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Errorf("symlink: %v", err)
	}
	if _, err := ReadTree(systemtest.New(), "/missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing dir: %v", err)
	}
	if _, err := ReadTree(systemtest.New().AddFile("/p/A", "x", 0o644).AddFile("/p/a", "y", 0o644), "/p"); err == nil {
		t.Error("colliding names passed")
	}
}

func TestReadTreeEnforcesTheLimits(t *testing.T) {
	big := systemtest.New().AddFile("/p/big", strings.Repeat("x", MaxFileBytes+1), 0o644)
	if _, err := ReadTree(big, "/p"); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Errorf("big file: %v", err)
	}
	many := systemtest.New()
	for index := range MaxFiles + 1 {
		many.AddFile("/p/f"+strings.Repeat("a", index%7)+"/"+itoa(index), "", 0o644)
	}
	if _, err := ReadTree(many, "/p"); err == nil || !strings.Contains(err.Error(), "more than 4096 files") {
		t.Errorf("many files: %v", err)
	}
	heavy := systemtest.New()
	chunk := strings.Repeat("x", MaxFileBytes)
	for index := range MaxTreeBytes/MaxFileBytes + 1 {
		heavy.AddFile("/p/f"+itoa(index), chunk, 0o644)
	}
	if _, err := ReadTree(heavy, "/p"); err == nil || !strings.Contains(err.Error(), "bytes") {
		t.Errorf("heavy tree: %v", err)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}

func TestReadFilesReadsTheNamedFilesThatExist(t *testing.T) {
	host := systemtest.New().AddFile("/m/a.json", "A", 0o644).AddFile("/m/dir/b.json", "B", 0o644)
	tree, err := ReadFiles(host, "/m", []string{"a.json", "missing.json", "dir/b.json"})
	if err != nil || len(tree) != 2 || tree[0].Path != "a.json" || string(tree[1].Content) != "B" {
		t.Errorf("tree = %v, %v", tree, err)
	}
	if _, err := ReadFiles(host, "/m", []string{"dir"}); err == nil {
		t.Error("a directory passed as a file")
	}
	if _, err := ReadFiles(systemtest.New().AddFile("/m/big", strings.Repeat("x", MaxFileBytes+1), 0o644), "/m", []string{"big"}); err == nil {
		t.Error("a big file passed")
	}
	if tree, err := ReadFiles(host, "/m", []string{"a.json/x"}); err != nil || len(tree) != 0 {
		t.Errorf("a path below a file = %v, %v", tree, err)
	}
}

// flaky fails one operation on one path, and passes the rest to the fake.
type flaky struct {
	*systemtest.Fake
	op, path string
	err      error
}

func (f flaky) Lstat(p string) (fs.FileInfo, error) {
	if f.op == "lstat" && p == f.path {
		return nil, f.err
	}
	return f.Fake.Lstat(p)
}

func (f flaky) ReadDir(p string) ([]fs.DirEntry, error) {
	if f.op == "readdir" && p == f.path {
		return nil, f.err
	}
	return f.Fake.ReadDir(p)
}

func (f flaky) ReadFile(p string) ([]byte, error) {
	if f.op == "read" && p == f.path {
		return nil, f.err
	}
	return f.Fake.ReadFile(p)
}

func TestReadersReportEveryFailureOfTheMachine(t *testing.T) {
	boom := errors.New("boom")
	base := func() *systemtest.Fake {
		return systemtest.New().AddFile("/p/sub/a", "x", 0o644).AddFile("/p/b", "y", 0o644)
	}
	for _, c := range []struct{ op, path string }{{"lstat", "/p/b"}, {"readdir", "/p/sub"}, {"read", "/p/sub/a"}} {
		if _, err := ReadTree(flaky{base(), c.op, c.path, boom}, "/p"); !errors.Is(err, boom) {
			t.Errorf("ReadTree with %s failing: %v", c.op, err)
		}
	}
	for _, c := range []struct{ op, path string }{{"lstat", "/p/b"}, {"read", "/p/b"}} {
		if _, err := ReadFiles(flaky{base(), c.op, c.path, boom}, "/p", []string{"b"}); !errors.Is(err, boom) {
			t.Errorf("ReadFiles with %s failing: %v", c.op, err)
		}
	}
	if err := CheckNoSymlinks(flaky{base(), "lstat", "/p/b", boom}, "/p", "b"); !errors.Is(err, boom) {
		t.Errorf("CheckNoSymlinks: %v", err)
	}
}

type odd struct{ fs.FileInfo }

func (odd) Mode() fs.FileMode { return fs.ModeNamedPipe }
func (odd) IsDir() bool       { return false }

type oddFiles struct{ *systemtest.Fake }

func (o oddFiles) Lstat(p string) (fs.FileInfo, error) {
	info, err := o.Fake.Lstat(p)
	if err == nil && p == "/p/pipe" {
		return odd{info}, nil
	}
	return info, err
}

func TestReadTreeRefusesAFileThatIsNeitherRegularNorADirectory(t *testing.T) {
	if _, err := ReadTree(oddFiles{systemtest.New().AddFile("/p/pipe", "", 0o644)}, "/p"); err == nil || !strings.Contains(err.Error(), "neither a regular file") {
		t.Errorf("pipe: %v", err)
	}
}

func TestCheckNoSymlinks(t *testing.T) {
	host := systemtest.New().AddFile("/c/plugins/a/x", "x", 0o644).AddLink("/c/evil", "/etc")
	if err := CheckNoSymlinks(host, "/c", "plugins/a"); err != nil {
		t.Errorf("a plain path failed: %v", err)
	}
	if err := CheckNoSymlinks(host, "/c", "evil"); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Errorf("a symlink passed: %v", err)
	}
	if err := CheckNoSymlinks(host, "/c", "missing/x"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a missing path: %v", err)
	}
	if err := CheckNoSymlinks(host, "/c", ""); err != nil {
		t.Errorf("the root failed: %v", err)
	}
}
