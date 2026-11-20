package sources_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, "plugins")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, sources.Check{}, ctx)
}

const skill = "---\nname: using-x\ndescription: d\n---\nbody"

const catalog = `{"version":1,"sources":[{"id":"a","kind":"web","title":"T","urls":["https://e.test"],"reviewedAt":"2026-01-01"}]}`

func TestIdentity(t *testing.T) {
	c := sources.Check{}
	if c.Name() != "sources" || c.Summary() == "" || c.DefaultPaths()[0] != "packages/runtime/plugin" {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "plugins"), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
}

func TestCatalogs(t *testing.T) {
	files := map[string]string{
		"plugins/p/skills/using-x/SKILL.md":     skill,
		"plugins/p/skills/using-x/sources.json": catalog,
	}
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("a sound catalog has no finding: %v", got)
	}
	files["plugins/p/skills/using-x/sources.json"] = `{"version":2}`
	got := run(t, files)
	f, ok := checktest.Find(got, "sources.schema")
	if !ok || f.Path != "plugins/p/skills/using-x/sources.json" || f.Check != "sources" {
		t.Fatalf("findings = %v", got)
	}
	files["plugins/p/skills/using-x/sources.json"] = `{"version":1,"sources":[` +
		`{"id":"a","kind":"web","title":"T","urls":["https://e.test"],"reviewedAt":"2026-01-01"},` +
		`{"id":"a","kind":"web","title":"T","urls":["https://e.test"],"reviewedAt":"2026-01-01"}]}`
	if got := run(t, files); !checktest.Has(got, "sources.duplicate-id") {
		t.Fatalf("codes = %v", checktest.Codes(got))
	}
}

func TestMissingCatalog(t *testing.T) {
	files := map[string]string{"plugins/p/skills/using-x/SKILL.md": skill}
	got := run(t, files)
	if f, ok := checktest.Find(got, "sources.missing"); !ok || f.Path != "plugins/p/skills/using-x" {
		t.Fatalf("findings = %v", got)
	}
	if got := run(t, files, map[string]any{"required": false}); len(got) != 0 {
		t.Fatalf("an optional catalog may be missing: %v", got)
	}
}
