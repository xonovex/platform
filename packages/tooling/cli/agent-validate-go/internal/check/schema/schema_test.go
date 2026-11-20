package schema_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/schema"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, schema.Check{}, ctx)
}

const goodProject = `{"version": 1}`

func TestIdentity(t *testing.T) {
	c := schema.Check{}
	if c.Name() != "schema" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
}

func TestRules(t *testing.T) {
	files := map[string]string{
		"conf/a.json":   goodProject,
		"conf/b.json":   `{"version": 2}`,
		"conf/c.json":   `{`,
		"other/d.json":  `{"version": 2}`,
		"conf/notes.md": "x",
	}
	rules := map[string]any{"rules": []map[string]any{{"schema": "validate-project", "files": []string{"conf/*.json"}}}}
	got := run(t, files, rules)
	paths := map[string]string{}
	for _, f := range got {
		paths[f.Path] = f.Code
	}
	if len(got) != 2 || paths["conf/b.json"] != "schema.invalid" || paths["conf/c.json"] != "schema.json" {
		t.Fatalf("findings = %v", got)
	}
	byID := map[string]any{"rules": []map[string]any{
		{"schema": schemas.ValidateProject, "files": []string{"conf/b.json"}},
		{"schema": "validate-project", "files": []string{"conf/b.json"}},
	}}
	if got := run(t, map[string]string{"conf/b.json": `{"version": 2}`}, byID); len(got) != 1 {
		t.Fatalf("one rule per schema validates once: %v", got)
	}
}

func TestOptionErrors(t *testing.T) {
	cases := map[string]any{
		"unknown schema": map[string]any{"rules": []map[string]any{{"schema": "nope", "files": []string{"a"}}}},
		"unknown ignore": map[string]any{"ignore": []string{"nope"}},
	}
	for name, options := range cases {
		if _, err := (schema.Check{}).Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), options)); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestDetection(t *testing.T) {
	declared := func(id, rest string) string { return `{"$schema": "` + id + `", ` + rest + `}` }
	files := map[string]string{
		"a.json": declared(schemas.ValidateProject, `"version": 2`),
		"b.json": declared("validate-project", `"version": 1`),
		"c.json": declared(schemas.Sources, `"version": 2`),
		"d.json": `{"$schema": "https://example.com/other.json"}`,
		"e.json": `[1]`,
		"f.json": `{`,
	}
	got := run(t, files)
	if len(got) != 1 || got[0].Path != "a.json" {
		t.Fatalf("a document that names a schema is checked, except an ignored one: %v", got)
	}
	got = run(t, files, map[string]any{"ignore": []string{}})
	seen := map[string]bool{}
	for _, f := range got {
		seen[f.Path] = true
	}
	if len(seen) != 2 || !seen["c.json"] {
		t.Fatalf("with no ignore list the sources document is checked too: %v", got)
	}
	if got := run(t, files, map[string]any{"detect": false}); len(got) != 0 {
		t.Fatalf("detection off: %v", got)
	}
}
