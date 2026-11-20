package routing_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/routing"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func skillMD(name string) string {
	return fmt.Sprintf("---\nname: %s\ndescription: \"x\"\n---\n# T\n", name)
}

func entry(query string, should bool, split, rationale string) string {
	return fmt.Sprintf(`{"id":"q","prompt":%q,"shouldTrigger":%t,"split":%q,"rationale":%q}`, query, should, split, rationale)
}

// queries wraps entries in a trigger eval document.
func queries(entries ...string) string {
	return `{"version":1,"skill":"x","queries":[` + strings.Join(entries, ",") + `]}`
}

func pair() map[string]string {
	return map[string]string{
		"c/a/skills/using-a/SKILL.md":          skillMD("using-a"),
		"c/a/skills/using-a/eval-trigger.json": queries(entry("alpha", true, "validation", ""), entry("beta", false, "validation", "")),
		"c/b/skills/using-b/SKILL.md":          skillMD("using-b"),
		"c/b/skills/using-b/eval-trigger.json": queries(entry("beta", true, "validation", ""), entry("alpha", false, "validation", "")),
	}
}

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "c")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, routing.Check{}, ctx)
}

func TestPairedOwnersAreClean(t *testing.T) {
	if got := run(t, pair()); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	c := routing.Check{}
	if c.Name() != "routing" || c.Summary() == "" || c.DefaultPaths()[0] != "packages/runtime/plugin" {
		t.Fatal("identity")
	}
}

func TestSkillWithoutValidationScenario(t *testing.T) {
	files := pair()
	files["c/b/skills/using-b/eval-trigger.json"] = queries(entry("beta", true, "train", ""), entry("alpha", false, "validation", ""))
	got := run(t, files)
	if !checktest.Has(got, "routing.no-owner") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestConflictingOwners(t *testing.T) {
	files := pair()
	files["c/b/skills/using-b/eval-trigger.json"] = queries(entry("Alpha.", true, "validation", ""), entry("beta", true, "validation", ""))
	got := run(t, files)
	f, ok := checktest.Find(got, "routing.conflict")
	if !ok || f.Message == "" {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestUnresolvedOperation(t *testing.T) {
	files := pair()
	files["c/a/skills/using-a/eval-trigger.json"] = queries(entry("alpha", true, "validation", "uses the `retired-op` operation and the `known` operation and a plain-compound operation"), entry("beta", false, "validation", ""))
	files["c/b/skills/using-b/references/known.md"] = "# k\n"
	files["c/b/skills/using-b/references/sub/x.md"] = "# nested\n"
	got := run(t, files)
	count := 0
	for _, f := range got {
		if f.Code == "routing.operation-unresolved" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestSkillsWithoutQueriesAreIgnoredAndBrokenQueriesReported(t *testing.T) {
	files := pair()
	files["c/c/skills/using-c/SKILL.md"] = skillMD("using-c")
	files["c/d/skills/using-d/SKILL.md"] = skillMD("using-d")
	files["c/d/skills/using-d/eval-trigger.json"] = "{"
	got := run(t, files)
	if !checktest.Has(got, "routing.queries-invalid") || len(checktest.Codes(got)) != 1 {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestMissingSkillName(t *testing.T) {
	files := pair()
	files["c/e/skills/using-e/SKILL.md"] = "# no frontmatter\n"
	files["c/e/skills/using-e/eval-trigger.json"] = queries(entry("q", true, "train", ""))
	if got := run(t, files); !checktest.Has(got, "routing.name-missing") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	files["c/e/skills/using-e/SKILL.md"] = "---\nname: [a\n---\n"
	if got := run(t, files); !checktest.Has(got, "routing.name-missing") {
		t.Fatal("a broken frontmatter has no name")
	}
}

func TestRejectsUnknownOptions(t *testing.T) {
	ctx := checktest.WithOptions(t, checktest.Context(t, pair(), "c"), map[string]any{"x": 1})
	if _, err := (routing.Check{}).Run(ctx); err == nil {
		t.Fatal("an unknown option is an error")
	}
}
