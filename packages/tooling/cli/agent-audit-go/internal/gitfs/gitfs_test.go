package gitfs_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/gitfs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process/processtest"
)

func tree(fake *processtest.Fake) gitfs.FS { return gitfs.FS{Runner: fake, Dir: "cat", Ref: "main"} }

func TestReadFile(t *testing.T) {
	fake := processtest.New().
		On("git show main:./sources.json", processtest.Reply{Stdout: "{}"}).
		On("git show main:./sources.d/a.json", processtest.Reply{ExitCode: 128})
	got, err := tree(fake).ReadFile("./sources.json")
	if err != nil || string(got) != "{}" || fake.Calls[0].Dir != "cat" {
		t.Fatalf("ReadFile = %q, %v, call %+v", got, err, fake.Calls[0])
	}
	if _, err := tree(fake).ReadFile("sources.d/a.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

func TestStat(t *testing.T) {
	fake := processtest.New().
		On("git cat-file -t main:./dir", processtest.Reply{Stdout: "tree\n"}).
		On("git cat-file -t main:./file", processtest.Reply{Stdout: "blob\n"}).
		On("git cat-file -s main:./file", processtest.Reply{Stdout: "42\n"}).
		On("git cat-file -t main:./gone", processtest.Reply{ExitCode: 128})
	if info, err := tree(fake).Stat("dir"); err != nil || !info.IsDir {
		t.Errorf("dir = %+v, %v", info, err)
	}
	if info, err := tree(fake).Stat("file"); err != nil || info.IsDir || info.Size != 42 {
		t.Errorf("file = %+v, %v", info, err)
	}
	if _, err := tree(fake).Stat("gone"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("gone: err = %v", err)
	}
}

func TestReadDir(t *testing.T) {
	fake := processtest.New().
		On("git cat-file -t main:./sources.d", processtest.Reply{Stdout: "tree\n"}).
		On("git ls-tree -z main:./sources.d", processtest.Reply{Stdout: "100644 blob aaa\tb.json\x00040000 tree bbb\tsub\x00100644 blob ccc\ta.json\x00"}).
		On("git cat-file -t main:./file", processtest.Reply{Stdout: "blob\n"}).
		On("git cat-file -s main:./file", processtest.Reply{Stdout: "1\n"}).
		On("git cat-file -t main:./gone", processtest.Reply{ExitCode: 128})
	got, err := tree(fake).ReadDir("sources.d")
	want := []fs.Entry{{Name: "a.json"}, {Name: "b.json"}, {Name: "sub", IsDir: true}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadDir = %v, %v", got, err)
	}
	if _, err := tree(fake).ReadDir("file"); !errors.Is(err, fs.ErrNotDir) {
		t.Errorf("file: err = %v", err)
	}
	if _, err := tree(fake).ReadDir("gone"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("gone: err = %v", err)
	}
}

func TestReadDirIgnoresMalformedRecords(t *testing.T) {
	fake := processtest.New().
		On("git cat-file -t main:./d", processtest.Reply{Stdout: "tree\n"}).
		On("git ls-tree -z main:./d", processtest.Reply{Stdout: "garbage\x00\x00100644\tx\x00"})
	got, err := tree(fake).ReadDir("d")
	if err != nil || !reflect.DeepEqual(got, []fs.Entry{{Name: "x"}}) {
		t.Fatalf("ReadDir = %v, %v", got, err)
	}
}

func TestFaultsOfTheRunnerPropagate(t *testing.T) {
	boom := errors.New("no git")
	cases := map[string]*processtest.Fake{
		"show": processtest.New().On("git show main:./f", processtest.Reply{Err: boom}),
		"type": processtest.New().On("git cat-file -t main:./f", processtest.Reply{Err: boom}),
		"size": processtest.New().On("git cat-file -t main:./f", processtest.Reply{Stdout: "blob"}).On("git cat-file -s main:./f", processtest.Reply{Err: boom}),
		"ls":   processtest.New().On("git cat-file -t main:./f", processtest.Reply{Stdout: "tree"}).On("git ls-tree -z main:./f", processtest.Reply{Err: boom}),
	}
	if _, err := tree(cases["show"]).ReadFile("f"); !errors.Is(err, boom) {
		t.Errorf("show: %v", err)
	}
	if _, err := tree(cases["type"]).Stat("f"); !errors.Is(err, boom) {
		t.Errorf("type: %v", err)
	}
	if _, err := tree(cases["size"]).Stat("f"); !errors.Is(err, boom) {
		t.Errorf("size: %v", err)
	}
	if _, err := tree(cases["ls"]).ReadDir("f"); !errors.Is(err, boom) {
		t.Errorf("ls: %v", err)
	}
	if _, err := tree(cases["type"]).ReadDir("f"); !errors.Is(err, boom) {
		t.Errorf("readdir stat: %v", err)
	}
}

func TestTheTreeIsReadOnly(t *testing.T) {
	f := tree(processtest.New())
	for name, err := range map[string]error{"write": f.WriteFile("a", nil), "mkdir": f.MkdirAll("a"), "remove": f.Remove("a")} {
		if !errors.Is(err, gitfs.ErrReadOnly) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
