package budget_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// flaky fails the reads and directory listings whose path ends with a listed suffix.
type flaky struct {
	*memfs.FileSystem
	failRead, failDir string
}

var errDisk = errors.New("disk fault")

func (f flaky) ReadFile(path string) ([]byte, error) {
	if f.failRead != "" && strings.HasSuffix(path, f.failRead) {
		return nil, errDisk
	}
	return f.FileSystem.ReadFile(path)
}

func (f flaky) ReadDir(path string) ([]fs.Entry, error) {
	if f.failDir != "" && strings.HasSuffix(path, f.failDir) {
		return nil, errDisk
	}
	return f.FileSystem.ReadDir(path)
}

func TestCollectReportsADiskFault(t *testing.T) {
	files := map[string]string{
		"c/p/skills/using-a/SKILL.md":        "one\n",
		"c/p/skills/using-a/references/x.md": "two\n",
	}
	cases := map[string]flaky{
		"skills":     {FileSystem: memfs.New(files), failDir: "c"},
		"SKILL.md":   {FileSystem: memfs.New(files), failRead: "using-a/SKILL.md"},
		"references": {FileSystem: memfs.New(files), failDir: "using-a/references"},
		"reference":  {FileSystem: memfs.New(files), failRead: "references/x.md"},
	}
	for name, disk := range cases {
		if _, err := budget.Collect(tree.Tree{FS: disk, Root: "."}, []string{"c"}); !errors.Is(err, errDisk) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func repo(files map[string]string) tree.Tree { return tree.Tree{FS: memfs.New(files), Root: "."} }

func TestCollectReadsSkillAndReferences(t *testing.T) {
	files, err := budget.Collect(repo(map[string]string{
		"c/p/skills/using-a/SKILL.md":            "---\nname: using-a\n---\none two three\n",
		"c/p/skills/using-a/references/x.md":     "four five\n```go\nignored words here\n```\n",
		"c/p/skills/using-a/references/y.txt":    "not markdown",
		"c/p/skills/using-a/references/sub/z.md": "nested",
		"c/p/skills/using-b/SKILL.md":            "x\n",
	}), []string{"c"})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range files {
		got[f.Path] = f.Kind
	}
	want := map[string]string{
		"c/p/skills/using-a/SKILL.md":        budget.KindSkill,
		"c/p/skills/using-a/references/x.md": budget.KindReference,
		"c/p/skills/using-b/SKILL.md":        budget.KindSkill,
	}
	if len(got) != len(want) {
		t.Fatalf("files = %v", got)
	}
	for path, kind := range want {
		if got[path] != kind {
			t.Errorf("%s: kind %q, want %q", path, got[path], kind)
		}
	}
}

func TestSeedRecordsProseWordsOnly(t *testing.T) {
	files := []budget.File{
		{Path: "a/SKILL.md", Kind: budget.KindSkill, Text: "one two three"},
		{Path: "a/references/code.md", Kind: budget.KindReference, Text: "```go\nx y z\n```\n"},
	}
	seeded := budget.Seed(files)
	if len(seeded) != 1 || seeded["a/SKILL.md"] != 3 {
		t.Fatalf("seeded = %v", seeded)
	}
}

func TestMarshalOrdersKeysByCodePoint(t *testing.T) {
	data, err := budget.Marshal(map[string]int{"b": 2, "B": 1, "a": 3})
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"B\": 1,\n  \"a\": 3,\n  \"b\": 2\n}\n"
	if string(data) != want {
		t.Fatalf("manifest = %q", data)
	}
	empty, err := budget.Marshal(nil)
	if err != nil || string(empty) != "{}\n" {
		t.Fatalf("empty = %q, %v", empty, err)
	}
}
