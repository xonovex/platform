package catalog_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// layout is the plugin layout of the trees below.
var layout = types.PluginLayout{Manifest: ".claude-plugin/plugin.json", SkillsDir: "skills"}

func TestLoadNeedsAPluginLayout(t *testing.T) {
	for _, bad := range []types.PluginLayout{{SkillsDir: "skills"}, {Manifest: "plugin.json"}} {
		if _, err := catalog.Load(memfs.New(tree()), bad, []string{"/p"}); err == nil {
			t.Errorf("Load(%+v) must fail", bad)
		}
	}
}

func TestLoadFollowsTheLayout(t *testing.T) {
	files := map[string]string{"/p/x/meta.json": `{"name":"x"}`, "/p/x/lib/one/SKILL.md": "o", "/p/y/.claude-plugin/plugin.json": `{"name":"y"}`}
	cat, err := catalog.Load(memfs.New(files), types.PluginLayout{Manifest: "meta.json", SkillsDir: "lib"}, []string{"/p"})
	if err != nil || len(cat.Plugins) != 1 || cat.Plugins[0].Name != "x" || len(cat.Plugins[0].Skills) != 1 || cat.Plugins[0].Skills[0].Dir != "/p/x/lib/one" {
		t.Fatalf("Load() = %+v, %v", cat, err)
	}
}

func manifest(name string, deps ...string) string {
	d := ""
	if len(deps) > 0 {
		d = `,"dependencies":["` + strings.Join(deps, `","`) + `"]`
	}
	return `{"name":"` + name + `","skills":"./skills/"` + d + `}`
}

func tree() map[string]string {
	return map[string]string{
		"/p/core/.claude-plugin/plugin.json":     manifest("core"),
		"/p/core/skills/alpha/SKILL.md":          "a",
		"/p/core/skills/alpha/eval-trigger.json": "{}",
		"/p/core/skills/alpha/eval-output.json":  "{}",
		"/p/core/skills/beta/SKILL.md":           "b",
		"/p/core/skills/notes/readme.md":         "not a skill",
		"/p/lang/.claude-plugin/plugin.json":     manifest("lang", "core"),
		"/p/lang/skills/gamma/SKILL.md":          "g",
		"/p/stray.txt":                           "x",
		"/p/empty/.claude-plugin/plugin.json":    manifest("empty"),
	}
}

func TestLoadFindsPluginsBelowARootAndAPluginRoot(t *testing.T) {
	cat, err := catalog.Load(memfs.New(tree()), layout, []string{"/p", "/p/lang"})
	if err == nil {
		t.Fatalf("lang is found twice: %+v", cat)
	}
	cat, err = catalog.Load(memfs.New(tree()), layout, []string{"/p"})
	if err != nil || len(cat.Plugins) != 3 || cat.Plugins[0].Name != "core" {
		t.Fatalf("%v %+v", err, cat)
	}
	direct, err := catalog.Load(memfs.New(tree()), layout, []string{"/p/lang"})
	if err != nil || len(direct.Plugins) != 1 || len(direct.Plugins[0].Skills) != 1 {
		t.Fatalf("%v %+v", err, direct)
	}
	core, _ := cat.Plugin("core")
	if len(core.Skills) != 2 || core.Skills[0].TriggerFile != "/p/core/skills/alpha/eval-trigger.json" ||
		core.Skills[0].OutputFile == "" || core.Skills[1].TriggerFile != "" {
		t.Fatalf("%+v", core.Skills)
	}
	if _, ok := cat.Plugin("nope"); ok {
		t.Fatal("unknown plugin")
	}
	if empty, _ := cat.Plugin("empty"); len(empty.Skills) != 0 {
		t.Fatal("a plugin without skills is empty")
	}
}

func TestLoadRejectsPluginsThatAreNotSkillOnly(t *testing.T) {
	for name, files := range map[string]map[string]string{
		"hooks":       {"/p/x/.claude-plugin/plugin.json": manifest("x"), "/p/x/hooks/h.json": "{}"},
		"settings":    {"/p/x/.claude-plugin/plugin.json": manifest("x"), "/p/x/.claude/settings.json": "{}"},
		"extra field": {"/p/x/.claude-plugin/plugin.json": `{"name":"x","mcpServers":{}}`},
		"bad json":    {"/p/x/.claude-plugin/plugin.json": `{`},
		"no name":     {"/p/x/.claude-plugin/plugin.json": `{"skills":"./skills/"}`},
		"bad deps":    {"/p/x/.claude-plugin/plugin.json": `{"name":"x","dependencies":[1]}`},
	} {
		if _, err := catalog.Load(memfs.New(files), layout, []string{"/p"}); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
	if _, err := catalog.Load(memfs.New(nil), layout, []string{"/missing"}); err == nil {
		t.Fatal("a missing root is an error")
	}
}

func TestSkillLookup(t *testing.T) {
	files := tree()
	files["/p/other/.claude-plugin/plugin.json"] = manifest("other")
	files["/p/other/skills/alpha/SKILL.md"] = "a2"
	cat, err := catalog.Load(memfs.New(files), layout, []string{"/p"})
	if err != nil {
		t.Fatal(err)
	}
	if s, err := cat.Skill("beta"); err != nil || s.Plugin != "core" {
		t.Fatalf("%v %+v", err, s)
	}
	if s, err := cat.Skill("other:alpha"); err != nil || s.Plugin != "other" {
		t.Fatalf("%v %+v", err, s)
	}
	if _, err := cat.Skill("alpha"); err == nil || !strings.Contains(err.Error(), "qualify") {
		t.Fatalf("%v", err)
	}
	if _, err := cat.Skill("zeta"); err == nil {
		t.Fatal("unknown skill")
	}
}

func TestClosureOrdersDependenciesFirst(t *testing.T) {
	cat, _ := catalog.Load(memfs.New(tree()), layout, []string{"/p"})
	got, err := cat.Closure("lang")
	if err != nil || len(got) != 2 || got[0].Name != "core" || got[1].Name != "lang" {
		t.Fatalf("%v %+v", err, got)
	}
	if _, err := cat.Closure("nope"); err == nil {
		t.Fatal("unknown plugin")
	}
	files := tree()
	files["/p/lang/.claude-plugin/plugin.json"] = manifest("lang", "missing")
	broken, _ := catalog.Load(memfs.New(files), layout, []string{"/p"})
	if _, err := broken.Closure("lang"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("%v", err)
	}
	files["/p/lang/.claude-plugin/plugin.json"] = manifest("lang", "core")
	files["/p/core/.claude-plugin/plugin.json"] = manifest("core", "lang")
	cyclic, _ := catalog.Load(memfs.New(files), layout, []string{"/p"})
	if _, err := cyclic.Closure("lang"); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("%v", err)
	}
	files["/p/lang/.claude-plugin/plugin.json"] = manifest("lang", "core", "core")
	files["/p/core/.claude-plugin/plugin.json"] = manifest("core")
	diamond, _ := catalog.Load(memfs.New(files), layout, []string{"/p"})
	if got, err := diamond.Closure("lang"); err != nil || len(got) != 2 {
		t.Fatalf("a repeated dependency is visited once: %v %+v", err, got)
	}
}

func q(text string, should bool, split model.Split) model.TriggerQuery {
	return model.TriggerQuery{Prompt: model.Prompt{Text: text}, ShouldTrigger: should, Split: split}
}

func TestScenarios(t *testing.T) {
	docs := map[string]model.TriggerEval{
		"a": {Queries: []model.TriggerQuery{q("shared", true, model.Train), q("both positive", true, model.Train), q("only a", true, model.Train), q("all negative", false, model.Train)}},
		"b": {Queries: []model.TriggerQuery{q("shared", false, model.Validation), q("both positive", true, model.Train), q("all negative", false, model.Train)}},
		"c": {Queries: []model.TriggerQuery{q("shared", false, model.Train)}},
	}
	got := catalog.Scenarios(docs)
	if len(got) != 1 || got[0].Prompt != "shared" || got[0].Positive != "a" || got[0].Split != model.Train ||
		strings.Join(got[0].Candidates, ",") != "a,b,c" {
		t.Fatalf("%+v", got)
	}
	docs["b"] = model.TriggerEval{Queries: []model.TriggerQuery{q("zed", true, model.Train), q("shared", false, model.Train)}}
	docs["c"] = model.TriggerEval{Queries: []model.TriggerQuery{q("zed", false, model.Train)}}
	ordered := catalog.Scenarios(docs)
	if len(ordered) != 2 || ordered[0].Positive != "a" || ordered[1].Positive != "b" {
		t.Fatalf("%+v", ordered)
	}
}
