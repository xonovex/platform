package migrate_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/migrate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

const legacyOutput = `{"skill_name":"alpha","tier":"conservative","evals":[
 {"id":7,"prompt":"do it","assertions":["is good","is short"],"files":["evals/files/seven/notes/a.txt","other.md"]},
 {"id":"x y","prompt":"again","assertions":["fine"],"files":[]},
 {"prompt":"no id","assertions":["ok"]}]}`

const legacyTrigger = `[{"query":"load it","should_trigger":true,"split":"train","rationale":"why"},
 {"query":"leave it","should_trigger":false,"split":"validation","rationale":"because"}]`

func TestOutputKeepsEverything(t *testing.T) {
	doc, err := migrate.Output([]byte(legacyOutput))
	if err != nil || doc.Skill != "alpha" || doc.Tier != model.Conservative || len(doc.Cases) != 3 {
		t.Fatalf("%v %+v", err, doc)
	}
	c := doc.Cases[0]
	if c.ID != "case-7" || c.Prompt.Prompt != "do it" || c.Gate != model.Required || len(c.Expect) != 2 ||
		c.Expect[1].Kind != model.KindLLM || c.Expect[1].Criterion != "is short" {
		t.Fatalf("%+v", c)
	}
	if len(c.Fixtures) != 2 || c.Fixtures[0].Path != "notes/a.txt" || c.Fixtures[0].ContentFile != "evals/files/seven/notes/a.txt" ||
		c.Fixtures[1].Path != "other.md" {
		t.Fatalf("%+v", c.Fixtures)
	}
	if doc.Cases[1].ID != "case-x-y" || doc.Cases[2].ID != "case-3" {
		t.Fatalf("%s %s", doc.Cases[1].ID, doc.Cases[2].ID)
	}
}

func TestTriggerKeepsEverything(t *testing.T) {
	doc, err := migrate.Trigger([]byte(legacyTrigger), "alpha")
	if err != nil || doc.Skill != "alpha" || len(doc.Queries) != 2 {
		t.Fatalf("%v %+v", err, doc)
	}
	q := doc.Queries[0]
	if q.ID != "query-0001" || q.Prompt.Prompt != "load it" || !q.ShouldTrigger || q.Split != model.Train || q.Rationale != "why" {
		t.Fatalf("%+v", q)
	}
}

func TestACaseWithoutAssertionsIsRefused(t *testing.T) {
	if _, err := migrate.Output([]byte(`{"skill_name":"a","evals":[{"id":1,"prompt":"p","assertions":[]}]}`)); err == nil {
		t.Fatal("want an error")
	}
}

func TestMalformedLegacyFiles(t *testing.T) {
	if _, err := migrate.Output([]byte("[")); err == nil {
		t.Fatal("output")
	}
	if _, err := migrate.Trigger([]byte("{"), "a"); err == nil {
		t.Fatal("trigger")
	}
}

func converter(t *testing.T, files map[string]string, remove bool) (migrate.Converter, fs.FileSystem) {
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	fsys := memfs.New(files)
	return migrate.Converter{FS: fsys, Reader: document.Reader{FS: fsys, Validator: v}, RemoveLegacy: remove}, fsys
}

func TestSkillWritesValidDocumentsThatReadBack(t *testing.T) {
	c, fsys := converter(t, map[string]string{
		"/s/alpha/evals.json":                    strings.Replace(legacyOutput, `"id":"x y"`, `"id":8`, 1),
		"/s/alpha/eval-queries.json":             legacyTrigger,
		"/s/alpha/evals/files/seven/notes/a.txt": "n",
		"/s/alpha/other.md":                      "o",
	}, false)
	res, err := c.Skill("/s/alpha")
	if err != nil || len(res.Written) != 2 || len(res.Removed) != 0 || res.Skill != "alpha" {
		t.Fatalf("%v %+v", err, res)
	}
	out, err := c.Reader.Output("/s/alpha/eval-output.json")
	if err != nil || len(out.Cases) != 3 || out.Cases[0].Fixtures[0].Text != "n" {
		t.Fatalf("%v %+v", err, out)
	}
	if _, err := c.Reader.Trigger("/s/alpha/eval-trigger.json"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := fs.Exists(fsys, "/s/alpha/evals.json"); !ok {
		t.Fatal("the legacy file stays unless asked")
	}
}

func TestSkillRemovesLegacyFilesOnRequestAndIgnoresPlainDirectories(t *testing.T) {
	c, fsys := converter(t, map[string]string{
		"/s/beta/eval-queries.json": legacyTrigger,
		"/s/gamma/SKILL.md":         "g",
	}, true)
	res, err := c.Skill("/s/beta")
	if err != nil || len(res.Removed) != 1 {
		t.Fatalf("%v %+v", err, res)
	}
	if ok, _ := fs.Exists(fsys, "/s/beta/eval-queries.json"); ok {
		t.Fatal("legacy file removed")
	}
	if res, err := c.Skill("/s/gamma"); err != nil || len(res.Written) != 0 {
		t.Fatalf("%v %+v", err, res)
	}
}

func TestSkillRejectsAnInvalidConversion(t *testing.T) {
	c, _ := converter(t, map[string]string{"/s/a/eval-queries.json": `[{"query":"q","should_trigger":true,"split":"bogus"}]`}, false)
	if _, err := c.Skill("/s/a"); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("%v", err)
	}
	c, _ = converter(t, map[string]string{"/s/a/evals.json": "["}, false)
	if _, err := c.Skill("/s/a"); err == nil {
		t.Fatal("malformed legacy file")
	}
}
