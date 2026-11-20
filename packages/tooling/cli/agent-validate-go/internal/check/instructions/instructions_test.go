package instructions_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/instructions"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

const agents = "# Title\n\nIntro.\n\n## Rules\n\n- Always run the gates.\n- Never push.\n\n## Other\n\n- Keep notes.\n"

const predicate = `{"kind": "session", "event": "command", "match": "git push"}`

func corpus(cases string) string {
	return `{"version": 1, "instruction": "AGENTS.md", "cases": [` + cases + `]}`
}

func caseOf(id, anchor string) string {
	return `{"id": "` + id + `", "prompt": "p", "anchor": ` + anchor + `, "expect": {"present": [` + predicate + `]}}`
}

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, instructions.Check{}, ctx)
}

func TestIdentity(t *testing.T) {
	c := instructions.Check{}
	if c.Name() != "instructions" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}

func TestCorpus(t *testing.T) {
	cases := map[string]struct {
		corpus string
		code   string
	}{
		"schema":    {`{"version": 1}`, "instructions.schema"},
		"heading":   {corpus(caseOf("a", `{"heading": "Missing"}`)), "instructions.anchor-heading"},
		"text":      {corpus(caseOf("a", `{"text": "Not in the file"}`)), "instructions.anchor-text"},
		"duplicate": {corpus(caseOf("a", `{"heading": "Rules"}`) + "," + caseOf("a", `{"heading": "Other"}`)), "instructions.duplicate-case"},
		"target":    {strings.Replace(corpus(caseOf("a", `{"heading": "Rules"}`)), "AGENTS.md", "OTHER.md", 1), "instructions.target-missing"},
	}
	for name, tc := range cases {
		got := run(t, map[string]string{"AGENTS.md": agents, "instructions.eval.json": tc.corpus})
		if !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
	}
	got := run(t, map[string]string{"AGENTS.md": agents, "instructions.eval.json": corpus(caseOf("a", `{"heading": "rules"}`) + "," + caseOf("b", `{"text": "Keep notes."}`))})
	if len(got) != 0 {
		t.Fatalf("a covered file is clean: %v", got)
	}
}

func TestCoverage(t *testing.T) {
	files := map[string]string{"AGENTS.md": agents, "instructions.eval.json": corpus(caseOf("a", `{"text": "Never push."}`))}
	got := run(t, files)
	var lines []int
	for _, f := range got {
		if f.Code != "instructions.unguarded-rule" || f.Level != findings.Warn || f.Path != "AGENTS.md" {
			t.Fatalf("finding = %v", f)
		}
		lines = append(lines, f.Line)
	}
	if len(lines) != 2 || lines[0] != 7 || lines[1] != 12 {
		t.Fatalf("lines = %v", lines)
	}
	if got := run(t, files, map[string]any{"coverage": false}); len(got) != 0 {
		t.Fatalf("coverage off: %v", got)
	}
}

func TestRequire(t *testing.T) {
	files := map[string]string{"AGENTS.md": agents, "pkg/AGENTS.md": agents}
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("a corpus is optional: %v", got)
	}
	got := run(t, files, map[string]any{"require": true})
	if len(got) != 2 || got[0].Code != "instructions.corpus-missing" {
		t.Fatalf("findings = %v", got)
	}
	files["instructions.eval.json"] = corpus(caseOf("a", `{"heading": "Rules"}`))
	if got := run(t, files, map[string]any{"require": true, "coverage": false}); len(got) != 1 || got[0].Path != "pkg/AGENTS.md" {
		t.Fatalf("findings = %v", got)
	}
}

func TestFencedBulletsAreNotRules(t *testing.T) {
	source := "## Rules\n\n```\n- not a rule\n```\n"
	got := run(t, map[string]string{"AGENTS.md": source, "instructions.eval.json": corpus(caseOf("a", `{"heading": "Rules"}`))})
	if len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}
