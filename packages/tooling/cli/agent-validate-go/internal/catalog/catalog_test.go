package catalog_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

func repo(files map[string]string) tree.Tree {
	return tree.Tree{FS: memfs.New(files), Root: ".", Skip: []string{"vendor"}}
}

func TestSkillDirs(t *testing.T) {
	tr := repo(map[string]string{
		"p/plugin-a/skills/using-a/SKILL.md":                "x",
		"p/plugin-a/skills/using-a/references/SKILL.md":     "nested file of the skill",
		"p/plugin-a/skills/using-b/SKILL.md":                "x",
		"p/plugin-a/skills/using-b/node_modules/x/SKILL.md": "skipped",
		"p/plugin-a/.hidden/using-c/SKILL.md":               "skipped",
		"p/plugin-a/vendor/using-d/SKILL.md":                "skipped",
		"p/plugin-b/skills/using-e/SKILL.md":                "x",
		"p/plugin-b/skills/using-e/assets/inner/SKILL.md":   "leaf",
		"q/note.md": "x",
	})
	got, err := catalog.SkillDirs(tr, []string{"p", "p/plugin-a/skills/using-a/SKILL.md", "p/plugin-b/skills/using-e", "q", "missing"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"p/plugin-a/skills/using-a", "p/plugin-a/skills/using-b", "p/plugin-b/skills/using-e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SkillDirs = %v, want %v", got, want)
	}
}

func TestReadQueries(t *testing.T) {
	tr := repo(map[string]string{
		"s/eval-trigger.json": `{"version":1,"skill":"x","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"train"},{"id":"r","promptFile":"p.md","shouldTrigger":false,"split":"validation"}]}`,
		"s/p.md":              "  long prompt \n",
	})
	got, _, err := catalog.ReadQueries(tr, "s")
	if err != nil || len(got) != 2 || got[0].Text != "a" || got[1].Text != "long prompt" || !got[0].ShouldTrigger {
		t.Fatalf("queries = %+v, %v", got, err)
	}
}

func TestReadQueriesErrors(t *testing.T) {
	cases := map[string]string{
		"not json":            `{`,
		"scalar":              `3`,
		"empty object":        `{"queries":[]}`,
		"bad entry":           `{"queries":["x"]}`,
		"object no should":    `{"queries":[{"prompt":"a"}]}`,
		"missing prompt file": `{"queries":[{"promptFile":"gone.md","shouldTrigger":true}]}`,
	}
	for name, body := range cases {
		tr := repo(map[string]string{"s/eval-trigger.json": body})
		if _, ok, err := catalog.ReadQueries(tr, "s"); err == nil || !ok || !strings.Contains(err.Error(), "s/eval-trigger.json") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if _, ok, err := catalog.ReadQueries(repo(nil), "none"); ok || err != nil {
		t.Fatal("a missing file is not an error")
	}
}

// brokenDir fails the listing of every directory whose path ends with a suffix.
type brokenDir struct {
	*memfs.FileSystem
	suffix string
}

var errDisk = errors.New("disk fault")

func (b brokenDir) ReadDir(path string) ([]fs.Entry, error) {
	if strings.HasSuffix(path, b.suffix) {
		return nil, errDisk
	}
	return b.FileSystem.ReadDir(path)
}

func TestSkillDirsReportsADiskFaultAtAnyDepth(t *testing.T) {
	files := map[string]string{"p/plugin-a/skills/using-a/SKILL.md": "x"}
	for _, suffix := range []string{"p", "p/plugin-a"} {
		tr := tree.Tree{FS: brokenDir{FileSystem: memfs.New(files), suffix: suffix}, Root: "."}
		if _, err := catalog.SkillDirs(tr, []string{"p"}); !errors.Is(err, errDisk) {
			t.Errorf("suffix %q: err = %v", suffix, err)
		}
	}
}
