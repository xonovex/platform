package drift_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/drift"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func words(n int) string { return strings.Repeat("word ", n) }

func skills(files map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range files {
		out[k] = v
	}
	return out
}

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "c")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, drift.Check{}, ctx)
}

func TestIdentity(t *testing.T) {
	c := drift.Check{}
	if c.Name() != "drift" || c.Summary() == "" || len(c.DefaultPaths()) != 1 {
		t.Fatal("identity")
	}
}

func TestCapsApplyToUnlistedFiles(t *testing.T) {
	files := skills(map[string]string{
		"c/p/skills/using-a/SKILL.md":             words(901),
		"c/p/skills/using-a/references/big.md":    words(651),
		"c/p/skills/using-a/references/ok.md":     words(650),
		"c/p/skills/using-a/references/notes.txt": words(2000),
		"c/p/skills/using-a/references/dir/x.md":  words(2000),
	})
	got := run(t, files)
	if n := len(checktest.Codes(got)); n != 2 {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	for _, f := range got {
		if f.Level != findings.Drift {
			t.Fatalf("level = %v", f.Level)
		}
	}
	custom := run(t, files, map[string]any{"skillCap": 2000, "referenceCap": 700})
	if len(custom) != 0 {
		t.Fatalf("findings = %v", custom)
	}
}

func TestBudgetManifest(t *testing.T) {
	files := skills(map[string]string{
		"c/p/skills/using-a/SKILL.md":        words(120),
		"c/p/skills/using-a/references/r.md": words(10),
		"budgets.json":                       `{"c/p/skills/using-a/SKILL.md":100,"c/p/skills/using-a/references/r.md":10}`,
	})
	got := run(t, files)
	if len(got) != 1 || got[0].Code != "drift.budget" || !strings.Contains(got[0].Message, "over its 100-word budget") {
		t.Fatalf("findings = %v", got)
	}
	fenced := skills(map[string]string{
		"c/p/skills/using-a/SKILL.md": "intro\n```\n" + words(2000) + "\n```\n",
	})
	if got := run(t, fenced); len(got) != 0 {
		t.Fatalf("fenced code is not counted: %v", got)
	}
}

func TestBrokenManifest(t *testing.T) {
	base := map[string]string{"c/p/skills/using-a/SKILL.md": words(5000)}
	cases := map[string]struct{ body, code string }{
		"json":     {`{`, "drift.budget-json"},
		"negative": {`{"a":-1}`, "drift.budget-entry"},
		"text":     {`{"a":"x"}`, "drift.budget-entry"},
		"fraction": {`{"a":1.5}`, "drift.budget-entry"},
	}
	for name, tc := range cases {
		files := skills(base)
		files["budgets.json"] = tc.body
		got := run(t, files)
		if !checktest.Has(got, tc.code) || checktest.Has(got, "drift.budget") {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
		if f, _ := checktest.Find(got, tc.code); f.Level != findings.Fail {
			t.Errorf("%s: a broken manifest fails", name)
		}
	}
	unreadable := skills(base)
	unreadable["budgets.json/inner"] = "x"
	if got := run(t, unreadable); !checktest.Has(got, "drift.budget-unreadable") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestDuplicationAcrossFiles(t *testing.T) {
	rule := "Agents must never commit secrets to the repository history under any circumstances."
	variant := "Agents must never commit secrets to the repository history under any circumstance."
	files := skills(map[string]string{
		"c/p/skills/using-a/SKILL.md": "# A\n\n" + rule + "\n",
		"c/p/skills/using-b/SKILL.md": "# B\n\n" + variant + " More text here.\n",
		"c/p/skills/using-c/SKILL.md": "# C\n\n```\n" + rule + "\n```\n" + rule + "\n",
	})
	got := run(t, files)
	if len(got) != 1 || got[0].Code != "drift.duplication" || !strings.Contains(got[0].Message, "3 files") {
		t.Fatalf("findings = %v", got)
	}
	two := skills(files)
	delete(two, "c/p/skills/using-c/SKILL.md")
	if got := run(t, two); len(got) != 0 {
		t.Fatalf("two files are below the limit: %v", got)
	}
	if got := run(t, two, map[string]any{"minFiles": 2}); len(got) != 1 {
		t.Fatalf("a lower limit finds the pair: %v", got)
	}
}

func TestShortAndNonNormativeSentencesAreIgnored(t *testing.T) {
	files := skills(map[string]string{
		"c/p/skills/using-a/SKILL.md": "Never do it.\n\nThe weather is mild and pleasant in this region today.\n",
		"c/p/skills/using-b/SKILL.md": "Never do it.\n\nThe weather is mild and pleasant in this region today.\n",
		"c/p/skills/using-c/SKILL.md": "Never do it.\n\nThe weather is mild and pleasant in this region today.\n",
	})
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestRejectsUnknownOptions(t *testing.T) {
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "c"), map[string]any{"nope": 1})
	if _, err := (drift.Check{}).Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}
