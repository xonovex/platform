package skill_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/skill"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

const dir = "plugin/skills/using-git"

const goodDescription = `Use when working with git history. Triggers on rebase, merge or cherry-pick, even when the user doesn't say 'git'.`

func skillFile(description, body string) string {
	return "---\nname: using-git\ndescription: \"" + description + "\"\n---\n" + body
}

const goodBody = "# Using git\n\n## Gotchas\n\n- One fact.\n\n```bash\ngit status\n```\n"

func queries() string {
	var parts []string
	for i := 0; i < 8; i++ {
		for _, split := range []string{"train", "validation"} {
			parts = append(parts,
				fmt.Sprintf(`{"id":"p-%s-%d","prompt":"positive %s %d","shouldTrigger":true,"split":"%s","rationale":"r"}`, split, i, split, i, split),
				fmt.Sprintf(`{"id":"n-%s-%d","prompt":"negative %s %d","shouldTrigger":false,"split":"%s","rationale":"near miss owned by owner-%d"}`, split, i, split, i, split, i%4))
		}
	}
	return triggerDoc(strings.Join(parts, ","))
}

func triggerDoc(queries string) string {
	return `{"version":1,"skill":"using-git","queries":[` + queries + `]}`
}

func outputCase(id string) string {
	return `{"id":"` + id + `","prompt":"p","gate":"required","expect":[{"kind":"llm","criterion":"x"}]}`
}

var goodEvals = `{"version":1,"skill":"using-git","tier":"moderate","cases":[` + outputCase("a") + "," + outputCase("b") + "," + outputCase("c") + `]}`

func base() map[string]string {
	return map[string]string{
		dir + "/SKILL.md":          skillFile(goodDescription, goodBody),
		dir + "/eval-output.json":  goodEvals,
		dir + "/eval-trigger.json": queries(),
	}
}

func run(t *testing.T, files map[string]string, options any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "plugin")
	if options != nil {
		ctx = checktest.WithOptions(t, ctx, options)
	}
	return checktest.Run(t, skill.Check{}, ctx)
}

func with(files map[string]string, changes map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range files {
		out[k] = v
	}
	for k, v := range changes {
		if v == "\x00" {
			delete(out, k)
		} else {
			out[k] = v
		}
	}
	return out
}

func expect(t *testing.T, got []findings.Finding, codes ...string) {
	t.Helper()
	for _, code := range codes {
		if !checktest.Has(got, code) {
			t.Errorf("missing %s in %v", code, checktest.Codes(got))
		}
	}
}

func TestCleanSkillHasNoFinding(t *testing.T) {
	if got := run(t, base(), nil); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestMetadata(t *testing.T) {
	c := skill.Check{}
	if c.Name() != "skill" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, base(), "plugin"), map[string]any{"nope": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("an unknown option is an error")
	}
}

func TestFrontmatterNameRules(t *testing.T) {
	cases := map[string]struct {
		front string
		code  string
	}{
		"missing":   {"---\ndescription: \"x\"\n---\n# T\n", "skill.name-missing"},
		"type":      {"---\nname: 3\ndescription: \"x\"\n---\n# T\n", "skill.name-type"},
		"length":    {"---\nname: " + strings.Repeat("a", 70) + "\ndescription: \"x\"\n---\n# T\n", "skill.name-length"},
		"kebab":     {"---\nname: Using_Git\ndescription: \"x\"\n---\n# T\n", "skill.name-kebab"},
		"gerund":    {"---\nname: git-usage\ndescription: \"x\"\n---\n# T\n", "skill.name-gerund"},
		"directory": {"---\nname: using-svn\ndescription: \"x\"\n---\n# T\n", "skill.name-directory"},
		"reserved":  {"---\nname: using-claude\ndescription: \"x\"\n---\n# T\n", "skill.name-reserved"},
	}
	for name, tc := range cases {
		got := run(t, with(base(), map[string]string{dir + "/SKILL.md": tc.front}), nil)
		if !checktest.Has(got, tc.code) {
			t.Errorf("%s: want %s, got %v", name, tc.code, checktest.Codes(got))
		}
	}
	vendor := with(base(), map[string]string{
		"plugin/skills/configuring-claude-code/SKILL.md":          strings.Replace(skillFile(goodDescription, goodBody), "using-git", "configuring-claude-code", 1),
		"plugin/skills/configuring-claude-code/eval-output.json":  strings.Replace(goodEvals, "using-git", "configuring-claude-code", 1),
		"plugin/skills/configuring-claude-code/eval-trigger.json": strings.Replace(queries(), "using-git", "configuring-claude-code", 1),
	})
	if got := run(t, vendor, nil); !checktest.Has(got, "skill.name-reserved") {
		t.Fatal("a vendor name is reserved unless the option exempts it")
	}
	if got := run(t, vendor, map[string]any{"vendorAdapters": []string{"configuring-claude-code"}}); checktest.Has(got, "skill.name-reserved") {
		t.Fatal("a vendor adapter may carry the vendor")
	}
	xml := with(base(), map[string]string{dir + "/SKILL.md": strings.Replace(skillFile(goodDescription, goodBody), "name: using-git", "name: using-<git>", 1)})
	expect(t, run(t, xml, nil), "skill.name-xml")
}

func TestFrontmatterDescriptionRules(t *testing.T) {
	cases := map[string]string{
		"skill.description-length":      strings.Repeat("a", 830),
		"skill.description-use-when":    "Handles git. Triggers on rebase, even when the user doesn't say 'git'.",
		"skill.description-triggers":    "Use when rebasing, even when the user doesn't say 'git'.",
		"skill.description-even-when":   "Use when rebasing. Triggers on rebase.",
		"skill.description-negative":    "Use when rebasing. Do not use for merges. Triggers on rebase, even when the user doesn't say 'git'.",
		"skill.description-names-skill": "Use when rebasing. Triggers on rebase, see writing-python, even when the user doesn't say 'git'.",
		"skill.description-markup":      "Use when rebasing <b>. Triggers on rebase, even when the user doesn't say 'git'.",
	}
	for code, description := range cases {
		got := run(t, with(base(), map[string]string{dir + "/SKILL.md": skillFile(description, goodBody)}), nil)
		if !checktest.Has(got, code) {
			t.Errorf("want %s, got %v", code, checktest.Codes(got))
		}
	}
	short := run(t, base(), map[string]any{"descriptionLimit": 10})
	expect(t, short, "skill.description-length")
	own := skillFile("Use when using-git applies. Triggers on rebase, even when the user doesn't say 'git'.", goodBody)
	if got := run(t, with(base(), map[string]string{dir + "/SKILL.md": own}), nil); checktest.Has(got, "skill.description-names-skill") {
		t.Fatal("a skill may name itself")
	}
	expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": "---\nname: using-git\n---\n" + goodBody}), nil), "skill.description-missing")
	expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": "---\nname: using-git\ndescription: 4\n---\n" + goodBody}), nil), "skill.description-type")
}

func TestFrontmatterOptionalFields(t *testing.T) {
	front := "---\nname: using-git\ndescription: \"" + goodDescription + "\"\ncompatibility: 3\nlicense: [a]\nmetadata:\n  a: 1\nallowed-tools: [x]\nextra: 1\nother: 2\n---\n" + goodBody
	got := run(t, with(base(), map[string]string{dir + "/SKILL.md": front}), nil)
	expect(t, got, "skill.compatibility-type", "skill.license-type", "skill.metadata-type", "skill.allowed-tools-type", "skill.unknown-field")
	long := "---\nname: using-git\ndescription: \"" + goodDescription + "\"\ncompatibility: " + strings.Repeat("a", 501) + "\nallowed-tools: \"Bash(git:*) Read $$\"\n---\n" + goodBody
	expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": long}), nil), "skill.compatibility-length", "skill.allowed-tools-characters")
	if f, _ := checktest.Find(got, "skill.unknown-field"); !strings.Contains(f.Message, "extra, other") {
		t.Fatal("unknown fields are sorted")
	}
	okTools := "---\nname: using-git\ndescription: \"" + goodDescription + "\"\nallowed-tools: Bash(git:*) Read\nmetadata:\n  a: b\nlicense: MIT\ncompatibility: node\n---\n" + goodBody
	if got := run(t, with(base(), map[string]string{dir + "/SKILL.md": okTools}), nil); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestFrontmatterErrors(t *testing.T) {
	got := run(t, with(base(), map[string]string{dir + "/SKILL.md": "---\nname: [a\n---\nbody"}), nil)
	expect(t, got, "skill.frontmatter-invalid")
	none := run(t, with(base(), map[string]string{dir + "/SKILL.md": "# Only a body\n"}), nil)
	expect(t, none, "skill.name-missing", "skill.description-missing")
}

func TestLoaderQuoting(t *testing.T) {
	cases := map[string]string{
		"skill.description-quoting": "---\nname: using-git\ndescription: plain text\n---\n" + goodBody,
		"skill.quoting-escape":      "---\nname: using-git\ndescription: \"x\"\nlicense: 'it''s'\n---\n" + goodBody,
		"skill.quoting-inner":       "---\nname: using-git\ndescription: \"x\"\nlicense: \"a \\\"b\\\" c\"\n---\n" + goodBody,
	}
	for code, content := range cases {
		expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": content}), nil), code)
	}
	multiline := "---\nname: using-git\ndescription: \"" + goodDescription + "\"\nlicense: 'open\n  continues'\n---\n" + goodBody
	if got := run(t, with(base(), map[string]string{dir + "/SKILL.md": multiline}), nil); checktest.Has(got, "skill.quoting-inner") {
		t.Fatal("a multiline scalar is not judged")
	}
	if got := run(t, with(base(), map[string]string{dir + "/SKILL.md": "---\nname: using-git\ndescription: \"" + goodDescription + "\"\nlicense: 'x'\n---\n" + goodBody}), nil); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestBody(t *testing.T) {
	long := goodBody + strings.Repeat("line\n", 500)
	expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, long)}), nil), "skill.body-length")
	warn := goodBody + strings.Repeat("line\n", 410)
	got := run(t, with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, warn)}), nil)
	if f, _ := checktest.Find(got, "skill.body-length"); f.Level != findings.Warn {
		t.Fatalf("level = %v", f.Level)
	}
	bare := "# T\n\n## Gotchas\n\n```\ncode\n```\n```\nmore\n```\n"
	expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, bare)}), nil), "skill.body-fence-language")
	expect(t, run(t, with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, "text only")}), nil), "skill.body-headings", "skill.body-gotchas")
}

func TestReferences(t *testing.T) {
	body := "# T\n\n## Gotchas\n\n## Progressive Disclosure\n\n- references/ok.md: Load when asked\n- references/missing.md: details\n- references/deep/inner.md\n- references/Bad_Name.md\n- references/<placeholder>.md\n\n## End\nUse @references/ok.md\n"
	files := with(base(), map[string]string{
		dir + "/SKILL.md":                 skillFile(goodDescription, body),
		dir + "/references/ok.md":         "# Ok\n",
		dir + "/references/deep/inner.md": "# Inner\n",
		dir + "/references/Bad_Name.md":   "# Bad\n",
	})
	got := run(t, files, nil)
	expect(t, got, "skill.reference-broken", "skill.reference-nested", "skill.reference-name", "skill.reference-eager", "skill.reference-load-when")
	count := 0
	for _, f := range got {
		if f.Code == "skill.reference-broken" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("one link is broken, got %d", count)
	}
	clean := with(files, map[string]string{
		dir + "/SKILL.md": skillFile(goodDescription, "# T\n\n## Gotchas\n\n## Progressive Disclosure\n\n- references/ok.md: Load when asked\n"),
	})
	for _, extra := range []string{"deep/inner.md", "Bad_Name.md"} {
		clean[dir+"/references/"+extra] = "# x\n"
	}
	if got := run(t, clean, nil); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestReferenceTocAndLinks(t *testing.T) {
	long := "# Long\n" + strings.Repeat("line\n", 210)
	withToc := "# Long\n\n## Contents\n\n" + strings.Repeat("line\n", 210)
	files := with(base(), map[string]string{
		dir + "/references/long.md":   long,
		dir + "/references/toc.md":    withToc,
		dir + "/references/links.md":  "[ok](long.md#a) [bad](gone.md) [ext](https://x.y) [ph](<x>.md) [frag](#a) `[code](nope.md)`\n",
		dir + "/references/deep/n.md": "[skip](../gone.md)\n",
	})
	got := run(t, files, nil)
	expect(t, got, "skill.reference-toc", "skill.reference-link-broken")
	var broken, toc int
	for _, f := range got {
		switch f.Code {
		case "skill.reference-link-broken":
			broken++
		case "skill.reference-toc":
			toc++
		}
	}
	if broken != 1 || toc != 1 {
		t.Fatalf("broken=%d toc=%d: %v", broken, toc, got)
	}
}

func TestSourcePlacementHarnessAndCredentials(t *testing.T) {
	body := "# T\n\n## Gotchas\n\n## External References\n\nUse TodoWrite and CLAUDE.md here.\n\n```bash\nexport GH_TOKEN=abc123\n```\n"
	files := with(base(), map[string]string{
		dir + "/SKILL.md":        skillFile(goodDescription, body),
		dir + "/references/r.md": "## External References\n\necho $API_TOKEN | docker login\n",
	})
	got := run(t, files, nil)
	expect(t, got, "skill.source-placement", "skill.harness-neutrality", "skill.credentials")
	var placement int
	for _, f := range got {
		if f.Code == "skill.source-placement" {
			placement++
		}
	}
	if placement != 2 {
		t.Fatalf("placement findings = %d", placement)
	}
	safe := with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, "# T\n\n## Gotchas\n\nexport GH_TOKEN=\"$X\"\n")})
	if got := run(t, safe, nil); checktest.Has(got, "skill.credentials") {
		t.Fatal("a variable reference is not a secret")
	}
	tokenFile := with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, "# T\n\n## Gotchas\n\ncmd < token.txt\n")})
	expect(t, run(t, tokenFile, nil), "skill.credentials")
	long := strings.Repeat("x", 100)
	wide := with(base(), map[string]string{dir + "/SKILL.md": skillFile(goodDescription, "# T\n\n## Gotchas\n\nTodoWrite "+long+"\n")})
	if f, _ := checktest.Find(run(t, wide, nil), "skill.harness-neutrality"); len(f.Message) > 130 {
		t.Fatal("the excerpt is capped")
	}
}

func TestImplementationFiles(t *testing.T) {
	files := with(base(), map[string]string{
		dir + "/helper.ts":       "x",
		dir + "/assets/tool.mjs": "x",
		dir + "/.hidden/skip.ts": "x",
		dir + "/scripts/run.sh":  "x",
	})
	got := run(t, files, nil)
	expect(t, got, "skill.implementation-files", "skill.scripts-compatibility", "skill.scripts-allowed-tools")
	f, _ := checktest.Find(got, "skill.implementation-files")
	if !strings.Contains(f.Message, "helper.ts") || !strings.Contains(f.Message, "assets/tool.mjs") || strings.Contains(f.Message, "skip.ts") {
		t.Fatalf("message = %s", f.Message)
	}
	declared := "---\nname: using-git\ndescription: \"" + goodDescription + "\"\ncompatibility: bash\nallowed-tools: Bash(git:*)\n---\n" + goodBody
	ok := with(base(), map[string]string{dir + "/SKILL.md": declared, dir + "/scripts/run.sh": "x", dir + "/scripts/.cache/x": "x"})
	if got := run(t, ok, nil); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	only := with(base(), map[string]string{dir + "/scripts/.hidden": "x"})
	if got := run(t, only, nil); len(got) != 0 {
		t.Fatalf("a hidden file is not a script: %v", got)
	}
	custom := run(t, with(base(), map[string]string{dir + "/x.py": "x"}), map[string]any{"forbiddenExtensions": []string{".py"}})
	expect(t, custom, "skill.implementation-files")
}

func TestOutputEvals(t *testing.T) {
	cases := map[string]struct {
		body string
		code string
	}{
		"not json": {`{`, "skill.evals-json"},
		"array":    {`[]`, "skill.evals-schema"},
		"no cases": {`{"version":1,"skill":"using-git","cases":[]}`, "skill.evals-schema"},
		"tier":     {`{"version":1,"skill":"using-git","tier":"bold","cases":[` + outputCase("a") + `]}`, "skill.evals-schema"},
		"name":     {`{"version":1,"skill":"other-thing","cases":[` + outputCase("a") + "," + outputCase("b") + "," + outputCase("c") + `]}`, "skill.evals-name"},
		"count":    {`{"version":1,"skill":"using-git","cases":[` + outputCase("a") + `]}`, "skill.evals-count"},
	}
	for name, tc := range cases {
		got := run(t, with(base(), map[string]string{dir + "/eval-output.json": tc.body}), nil)
		if !checktest.Has(got, tc.code) {
			t.Errorf("%s: want %s, got %v", name, tc.code, checktest.Codes(got))
		}
	}
	missing := base()
	delete(missing, dir+"/eval-output.json")
	expect(t, run(t, missing, nil), "skill.evals-missing")
	if got := run(t, missing, map[string]any{"requireEvals": false}); checktest.Has(got, "skill.evals-missing") {
		t.Fatal("evals are optional when not required")
	}
}

func TestTriggerEvals(t *testing.T) {
	entry := func(id, prompt string, should bool, split, rationale string) string {
		out := fmt.Sprintf(`{"id":%q,"prompt":%q,"shouldTrigger":%t,"split":%q`, id, prompt, should, split)
		if rationale != "" {
			out += fmt.Sprintf(`,"rationale":%q`, rationale)
		}
		return out + "}"
	}
	cases := map[string]struct {
		body string
		code string
	}{
		"not json":  {`{`, "skill.evals-json"},
		"scalar":    {`3`, "skill.evals-schema"},
		"no query":  {triggerDoc(``), "skill.evals-schema"},
		"minimum":   {triggerDoc(entry("a", "a", true, "train", "")), "skill.evals-minimum"},
		"splitmin":  {triggerDoc(entry("a", "a", true, "train", "")), "skill.evals-split-minimum"},
		"generic":   {triggerDoc(entry("a", "quick pre-merge sanity check: after the build", true, "train", "")), "skill.evals-generic"},
		"duplicate": {triggerDoc(entry("a", "A", true, "train", "") + "," + entry("b", " a ", true, "train", "")), "skill.evals-duplicate"},
		"owners": {triggerDoc(entry("a", "a", false, "train", "near miss owned by one") + "," + entry("b", "b", false, "train", "near miss owned by one") +
			"," + entry("c", "c", false, "train", "near miss owned by one")), "skill.evals-owners"},
		"unreadable": {triggerDoc(`{"id":"a","promptFile":"gone.md","shouldTrigger":true,"split":"train"}`), "skill.evals-queries"},
	}
	for name, tc := range cases {
		got := run(t, with(base(), map[string]string{dir + "/eval-trigger.json": tc.body}), nil)
		if !checktest.Has(got, tc.code) {
			t.Errorf("%s: want %s, got %v", name, tc.code, checktest.Codes(got))
		}
	}
	missing := base()
	delete(missing, dir+"/eval-trigger.json")
	expect(t, run(t, missing, nil), "skill.evals-missing")
	if got := run(t, base(), nil); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestUnusedCheckImport(t *testing.T) {
	var _ check.Check = skill.Check{}
}
