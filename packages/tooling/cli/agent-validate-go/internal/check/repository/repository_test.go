package repository_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/repository"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

const goodWorkflow = `name: Release
on:
  pull_request:
  workflow_dispatch:
jobs:
  release:
    steps:
      - uses: actions/checkout@v4
      - name: Publish
        if: github.event_name == 'pull_request'
        run: publish
      - name: Publish (dry run)
        if: github.event_name == 'workflow_dispatch'
        run: dry
      - uses: other
`

func base() map[string]string {
	return map[string]string{
		"AGENTS.md":                        "# Root\n",
		"CLAUDE.md":                        repository.ClaudePointer,
		"README.md":                        "# Platform\n",
		".github/workflows/release.yml":    goodWorkflow,
		"packages/AGENTS.md":               "# P\n",
		"packages/CLAUDE.md":               "See @AGENTS.md",
		"packages/tool/group/AGENTS.md":    "# G\n",
		"packages/tool/group/CLAUDE.md":    "See @AGENTS.md",
		"packages/tool/group/pkg/moon.yml": "x",
		"packages/tool/AGENTS.md":          "# T\n",
		"packages/tool/CLAUDE.md":          "See @AGENTS.md",
	}
}

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, repository.Check{}, ctx)
}

func with(changes map[string]string) map[string]string {
	files := base()
	for k, v := range changes {
		if v == "\x00" {
			delete(files, k)
		} else {
			files[k] = v
		}
	}
	return files
}

func TestIdentityAndCleanRun(t *testing.T) {
	c := repository.Check{}
	if c.Name() != "repository" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if got := run(t, base()); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}

func TestInstructionPairs(t *testing.T) {
	cases := map[string]struct {
		changes map[string]string
		code    string
	}{
		"no claude":  {map[string]string{"packages/tool/CLAUDE.md": "\x00"}, "repository.claude-missing"},
		"no pointer": {map[string]string{"CLAUDE.md": "Nothing"}, "repository.claude-pointer"},
		"group":      {map[string]string{"packages/tool/AGENTS.md": "\x00", "packages/tool/CLAUDE.md": "\x00"}, "repository.group-undocumented"},
		"nested":     {map[string]string{"packages/tool/group/pkg/skills/x/AGENTS.md": "x"}, "repository.claude-missing"},
		"custom":     {map[string]string{"packages/tool/group/pkg/evals/AGENTS.md": "x"}, ""},
	}
	for name, tc := range cases {
		got := run(t, with(tc.changes))
		if tc.code != "" && !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
		if tc.code == "" && len(got) != 0 {
			t.Errorf("%s: fixtures hold test data: %v", name, got)
		}
	}
	fixture := with(map[string]string{"packages/tool/group/pkg/data/AGENTS.md": "x"})
	if got := run(t, fixture, map[string]any{"fixtures": []string{"data"}}); len(got) != 0 {
		t.Fatalf("custom fixtures: %v", got)
	}
	if got := run(t, with(map[string]string{"packages/tool/AGENTS.md": "\x00", "packages/tool/CLAUDE.md": "\x00"}), map[string]any{"packages": ""}); len(got) != 0 {
		t.Fatalf("the group rule is optional: %v", got)
	}
	if got := run(t, base(), map[string]any{"packages": "absent"}); len(got) != 0 {
		t.Fatalf("a missing package tree has no group: %v", got)
	}
}

func TestReadmeClaims(t *testing.T) {
	got := run(t, with(map[string]string{"README.md": "# P\n\nOur skills provide enforcement of rules.\n"}))
	f, ok := checktest.Find(got, "repository.readme-claim")
	if !ok || f.Line != 3 {
		t.Fatalf("findings = %v", got)
	}
	if got := run(t, with(map[string]string{"README.md": "\x00"})); !checktest.Has(got, "repository.readme-missing") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
	if got := run(t, with(map[string]string{"README.md": "banned phrase"}), map[string]any{"forbiddenClaims": []string{"banned"}}); !checktest.Has(got, "repository.readme-claim") {
		t.Fatal("custom claims apply")
	}
	if _, err := (repository.Check{}).Run(checktest.WithOptions(t, checktest.Context(t, base(), "."), map[string]any{"forbiddenClaims": []string{"("}})); err == nil {
		t.Fatal("a bad pattern is an error")
	}
	if got := run(t, with(map[string]string{"README.md": "\x00"}), map[string]any{"readme": ""}); len(got) != 0 {
		t.Fatalf("the claim rule is optional: %v", got)
	}
}

func TestReleaseWorkflow(t *testing.T) {
	cases := map[string]struct {
		workflow string
		code     string
	}{
		"push":      {"on:\n  push:\n" + goodWorkflow, "repository.workflow-tags"},
		"tags":      {goodWorkflow + "  refs/tags/x\n", "repository.workflow-tags"},
		"input":     {goodWorkflow + "if: inputs.dry_run != true\n", "repository.workflow-dispatch-input"},
		"publish":   {"steps:\n      - name: Publish\n        run: x\n      - uses: y\n", "repository.workflow-publish"},
		"no step":   {"steps:\n      - uses: y\n", "repository.workflow-publish"},
		"dry run":   {"steps:\n      - name: Publish (dry run)\n        run: x\n      - uses: y\n", "repository.workflow-dry-run"},
		"last step": {"steps:\n      - name: Publish\n        if: github.event_name == 'pull_request'\n", ""},
	}
	for name, tc := range cases {
		got := run(t, with(map[string]string{".github/workflows/release.yml": tc.workflow}))
		if tc.code != "" && !checktest.Has(got, tc.code) {
			t.Errorf("%s: codes = %v", name, checktest.Codes(got))
		}
		if tc.code == "" && checktest.Has(got, "repository.workflow-publish") {
			t.Errorf("%s: a final step has a body: %v", name, checktest.Codes(got))
		}
	}
	if got := run(t, with(map[string]string{".github/workflows/release.yml": "\x00"})); !checktest.Has(got, "repository.workflow-missing") {
		t.Fatal("a missing workflow is reported")
	}
	if got := run(t, with(map[string]string{".github/workflows/release.yml": "\x00"}), map[string]any{"workflow": ""}); len(got) != 0 {
		t.Fatalf("the workflow rule is optional: %v", got)
	}
}
