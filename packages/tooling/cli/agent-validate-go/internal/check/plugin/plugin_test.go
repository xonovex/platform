package plugin_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/plugin"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func manifest(name string, deps ...string) string {
	quoted := make([]string, len(deps))
	for i, d := range deps {
		quoted[i] = fmt.Sprintf("%q", d)
	}
	return fmt.Sprintf(`{"name":%q,"version":"1.0.0","description":"d","skills":"./skills/","dependencies":[%s]}`, name, strings.Join(quoted, ","))
}

func pkgJSON(name, version string, deps map[string]string) string {
	parts := []string{}
	for k, v := range deps {
		parts = append(parts, fmt.Sprintf("%q:%q", k, v))
	}
	return fmt.Sprintf(`{"name":%q,"version":%q,"description":"d","dependencies":{%s}}`, name, version, strings.Join(parts, ","))
}

func skill(name, body string) map[string]string {
	return map[string]string{"skills/" + name + "/SKILL.md": "---\nname: " + name + "\n---\n" + body}
}

func addPlugin(files map[string]string, short string, deps []string, pins map[string]string, extra map[string]string) {
	dir := "root/plugin-" + short + "/"
	name := "xonovex-" + short
	files[dir+".claude-plugin/plugin.json"] = manifest(name, deps...)
	files[dir+".codex-plugin/plugin.json"] = manifest(name, deps...)
	files[dir+"package.json"] = pkgJSON("@xonovex/plugin-"+short, "1.0.0", pins)
	files[dir+"README.md"] = "# R\n\n## Skills\n\n| Skill | Use |\n| --- | --- |\n| [using-" + short + "](skills/using-" + short + "/SKILL.md) | x |\n"
	for k, v := range skill("using-"+short, "# T\n") {
		files[dir+k] = v
	}
	for k, v := range extra {
		files[dir+k] = v
	}
}

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "root")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, plugin.Check{}, ctx)
}

func TestIdentity(t *testing.T) {
	c := plugin.Check{}
	if c.Name() != "plugin" || c.Summary() == "" || c.DefaultPaths()[0] != "packages/runtime/plugin" {
		t.Fatal("identity")
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "root"), map[string]any{"x": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}

func TestCleanPackage(t *testing.T) {
	files := map[string]string{}
	addPlugin(files, "git", nil, nil, nil)
	files["root/not-a-plugin/readme.txt"] = "x"
	files["root/stray.txt"] = "x"
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestManifestProblems(t *testing.T) {
	cases := map[string]struct {
		change map[string]string
		code   string
	}{
		"invalid":       {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": "{"}, "plugin.manifest-invalid"},
		"no name":       {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": `{"skills":"./skills/"}`}, "plugin.manifest-name"},
		"bad package":   {map[string]string{"root/plugin-git/package.json": "{"}, "plugin.package-invalid"},
		"wrong name":    {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": manifest("xonovex-other"), "root/plugin-git/.claude-plugin/plugin.json": manifest("xonovex-other")}, "plugin.name"},
		"names differ":  {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": manifest("xonovex-other")}, "plugin.names-differ"},
		"deps differ":   {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": manifest("xonovex-git", "xonovex-x")}, "plugin.dependencies-differ"},
		"skills differ": {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": `{"name":"xonovex-git","skills":["./skills/","./more/"]}`}, "plugin.skills-differ"},
		"skills path":   {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": `{"name":"xonovex-git","skills":"./other/"}`, "root/plugin-git/.claude-plugin/plugin.json": `{"name":"xonovex-git","skills":"./other/"}`}, "plugin.skills-path"},
		"skills list":   {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": `{"name":"xonovex-git","skills":["./skills/"]}`, "root/plugin-git/.claude-plugin/plugin.json": `{"name":"xonovex-git","skills":["./skills/"]}`}, "plugin.skills-form"},
		"commands":      {map[string]string{"root/plugin-git/commands/x.md": "x"}, "plugin.commands-present"},
		"no codex":      {map[string]string{"root/plugin-git/.codex-plugin/plugin.json": "\x00"}, "plugin.manifest-missing"},
		"no claude":     {map[string]string{"root/plugin-git/.claude-plugin/plugin.json": "\x00"}, "plugin.manifest-missing"},
	}
	for name, tc := range cases {
		files := map[string]string{}
		addPlugin(files, "git", nil, nil, nil)
		for k, v := range tc.change {
			if v == "\x00" {
				delete(files, k)
			} else {
				files[k] = v
			}
		}
		if got := run(t, files); !checktest.Has(got, tc.code) {
			t.Errorf("%s: want %s, got %v", name, tc.code, checktest.Codes(got))
		}
	}
	empty := map[string]string{}
	addPlugin(empty, "git", nil, nil, nil)
	delete(empty, "root/plugin-git/skills/using-git/SKILL.md")
	empty["root/plugin-git/skills/.keep"] = ""
	if got := run(t, empty); !checktest.Has(got, "plugin.no-skills") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestReadmeCatalog(t *testing.T) {
	files := map[string]string{}
	addPlugin(files, "git", nil, nil, nil)
	files["root/plugin-git/README.md"] = "# R\n\n| [using-other](skills/using-other/SKILL.md) | outside the section |\n\n## Skills\n\n| [using-ghost](skills/using-ghost/SKILL.md) | x |\n| [using-git](skills/other/SKILL.md) | mismatched |\n\n## Other\n\n| [using-git](skills/using-git/SKILL.md) | after the section |\n"
	got := run(t, files)
	if !checktest.Has(got, "plugin.readme-skill-missing") || !checktest.Has(got, "plugin.readme-skill-unlisted") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	files["root/plugin-git/README.md"] = "# No skills heading\n"
	if got := run(t, files); !checktest.Has(got, "plugin.readme-skill-unlisted") {
		t.Fatal("a README without the heading lists no skill")
	}
	delete(files, "root/plugin-git/README.md")
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("a README is optional: %v", got)
	}
}

func TestDependencies(t *testing.T) {
	files := map[string]string{}
	addPlugin(files, "base", nil, nil, nil)
	addPlugin(files, "app", []string{"xonovex-base"}, map[string]string{"@xonovex/plugin-base": "1.0.0"},
		map[string]string{"skills/using-app/SKILL.md": "---\nname: using-app\n---\nHand to **using-base**.\n"})
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}

	unpinned := map[string]string{}
	addPlugin(unpinned, "base", nil, nil, nil)
	addPlugin(unpinned, "app", []string{"xonovex-base"}, nil, map[string]string{"skills/using-app/SKILL.md": "---\nname: using-app\n---\n**using-base**\n"})
	if got := run(t, unpinned); !checktest.Has(got, "plugin.dependency-unpinned") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}

	wrong := map[string]string{}
	addPlugin(wrong, "base", nil, nil, nil)
	addPlugin(wrong, "app", []string{"xonovex-base"}, map[string]string{"@xonovex/plugin-base": "0.9.0"}, map[string]string{"skills/using-app/SKILL.md": "---\nname: using-app\n---\n**using-base**\n"})
	if got := run(t, wrong); !checktest.Has(got, "plugin.dependency-pin") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}

	missing := map[string]string{}
	addPlugin(missing, "app", []string{"xonovex-gone"}, map[string]string{"@xonovex/plugin-gone": "1.0.0"}, nil)
	got := run(t, missing)
	if !checktest.Has(got, "plugin.dependency-missing") || !checktest.Has(got, "plugin.dependency-unused") || !checktest.Has(got, "plugin.package-dependency-unused") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}

	undeclared := map[string]string{}
	addPlugin(undeclared, "base", nil, nil, nil)
	addPlugin(undeclared, "app", nil, map[string]string{"@xonovex/plugin-base": "1.0.0"}, nil)
	if got := run(t, undeclared); !checktest.Has(got, "plugin.dependency-undeclared") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}

	noVersion := map[string]string{}
	addPlugin(noVersion, "base", nil, nil, nil)
	noVersion["root/plugin-base/package.json"] = `{"name":"@xonovex/plugin-base"}`
	addPlugin(noVersion, "app", []string{"xonovex-base"}, map[string]string{"@xonovex/plugin-base": "1.0.0"}, map[string]string{"skills/using-app/SKILL.md": "---\nname: using-app\n---\n**using-base**\n"})
	if got := run(t, noVersion); !checktest.Has(got, "plugin.dependency-version") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestCycles(t *testing.T) {
	files := map[string]string{}
	addPlugin(files, "a", []string{"xonovex-b"}, map[string]string{"@xonovex/plugin-b": "1.0.0"}, map[string]string{"skills/using-a/SKILL.md": "---\nname: using-a\n---\n**using-b**\n"})
	addPlugin(files, "b", []string{"xonovex-c"}, map[string]string{"@xonovex/plugin-c": "1.0.0"}, map[string]string{"skills/using-b/SKILL.md": "---\nname: using-b\n---\n**using-c**\n"})
	addPlugin(files, "c", []string{"xonovex-a"}, map[string]string{"@xonovex/plugin-a": "1.0.0"}, map[string]string{"skills/using-c/SKILL.md": "---\nname: using-c\n---\n**using-a**\n"})
	got := run(t, files)
	count := 0
	for _, f := range got {
		if f.Code == "plugin.dependency-cycle" {
			count++
			if !strings.Contains(f.Message, "xonovex-a -> xonovex-b -> xonovex-c -> xonovex-a") {
				t.Fatalf("message = %s", f.Message)
			}
		}
	}
	if count != 1 {
		t.Fatalf("one cycle expected: %v", got)
	}
}

func TestReferenceOverlap(t *testing.T) {
	text := strings.Repeat("alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima mike november oscar papa quebec romeo sierra tango uniform victor whiskey xray yankee zulu ", 2)
	files := map[string]string{}
	addPlugin(files, "base", nil, nil, map[string]string{"skills/using-base/references/ref.md": text})
	addPlugin(files, "app", []string{"xonovex-base"}, map[string]string{"@xonovex/plugin-base": "1.0.0"}, map[string]string{
		"skills/using-app/SKILL.md":         "---\nname: using-app\n---\n**using-base**\n",
		"skills/using-app/references/c.md":  text,
		"skills/using-app/references/s.md":  "short text only",
		"skills/using-app/references/d.txt": text,
	})
	got := run(t, files)
	if !checktest.Has(got, "plugin.reference-overlap") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	files["root/plugin-app/skills/using-app/references/c.md"] = "# Different words entirely here, " + strings.Repeat("unrelated sentence content number one two three four five six seven eight nine ten ", 4)
	if got := run(t, files); checktest.Has(got, "plugin.reference-overlap") {
		t.Fatal("different text is not a copy")
	}
}
