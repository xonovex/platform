package catalog_test

import (
	"reflect"
	"strings"
	"testing"

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

func TestReadQueriesList(t *testing.T) {
	tr := repo(map[string]string{
		"s/eval-queries.json": `[{"query":"a","should_trigger":true,"split":"train","rationale":"r"},{"query":"b","should_trigger":false}]`,
	})
	got, ok, err := catalog.ReadQueries(tr, "s")
	if err != nil || !ok || len(got) != 2 || got[0].Rationale != "r" || got[1].ShouldTrigger || got[1].Split != "" {
		t.Fatalf("queries = %+v, %v, %v", got, ok, err)
	}
	if _, ok, err := catalog.ReadQueries(tr, "none"); ok || err != nil {
		t.Fatal("a missing file is not an error")
	}
}

func TestReadQueriesObject(t *testing.T) {
	tr := repo(map[string]string{
		"s/eval-queries.json": `{"version":1,"skill":"x","queries":[{"id":"q","prompt":"a","shouldTrigger":true,"split":"train"},{"id":"r","promptFile":"p.md","shouldTrigger":false,"split":"validation"}]}`,
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
		"empty list":          `[]`,
		"bad entry":           `["x"]`,
		"no should":           `[{"query":"a"}]`,
		"empty query":         `[{"query":"","should_trigger":true}]`,
		"bad split":           `[{"query":"a","should_trigger":true,"split":"dev"}]`,
		"empty object":        `{"queries":[]}`,
		"object no should":    `{"queries":[{"prompt":"a"}]}`,
		"missing prompt file": `{"queries":[{"promptFile":"gone.md","shouldTrigger":true}]}`,
	}
	for name, body := range cases {
		tr := repo(map[string]string{"s/eval-queries.json": body})
		if _, ok, err := catalog.ReadQueries(tr, "s"); err == nil || !ok || !strings.Contains(err.Error(), "s/eval-queries.json") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}
