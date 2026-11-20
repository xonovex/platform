package kit_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/kit"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string) []findings.Finding {
	t.Helper()
	return checktest.Run(t, kit.Check{}, checktest.Context(t, files, "."))
}

func spec(body string) map[string]string { return map[string]string{"kits/a/spec.yaml": body} }

const valid = `schemaVersion: "2"
kind: sandbox
name: demo
sandbox: {}
args:
  token:
    required: true
permissions:
  network:
    allow:
      - api.example.com
credentials:
  - service: demo
    apiKey:
      name: DEMO_KEY
      inject:
        - domain: api.example.com
          format: "Bearer %s"
environment:
  variables:
    TOKEN: ${{ kit.args.token }}
`

func TestIdentity(t *testing.T) {
	c := kit.Check{}
	if c.Name() != "kit" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
}

func TestValid(t *testing.T) {
	if got := run(t, spec(valid)); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	if got := run(t, map[string]string{"kits/a/spec.yml": "schemaVersion: '2'\nkind: mixin\nname: m\n", "kits/a/other.yaml": "["}); len(got) != 0 {
		t.Fatalf("a yml spec is read and other files are not: %v", got)
	}
}

func TestInvalid(t *testing.T) {
	cases := map[string]struct {
		body string
		code string
		line int
	}{
		"yaml":     {"a: [\n", "kit.yaml", 0},
		"yaml any": {"\tbad: x", "kit.yaml", 0},
		"schema":   {"schemaVersion: '1'\nkind: mixin\nname: m\n", "kit.schema", 1},
		"missing":  {"schemaVersion: '2'\nkind: mixin\n", "kit.schema", 1},
		"extra":    {"schemaVersion: '2'\nkind: mixin\nname: m\nextra: 1\n", "kit.schema", 1},
		"domain":   {strings.Replace(valid, "      - api.example.com\n", "      - other.example.com\n", 1), "kit.domain", 0},
		"format":   {strings.Replace(valid, "Bearer %s", "Bearer", 1), "kit.format", 0},
		"argument": {valid + "    OTHER: ${{ kit.args.ghost }} and ${{ kit.args.ghost }}\n", "kit.argument", 0},
	}
	for name, tc := range cases {
		got := run(t, spec(tc.body))
		f, ok := checktest.Find(got, tc.code)
		if !ok {
			t.Errorf("%s: findings = %v", name, got)
			continue
		}
		if tc.line > 0 && f.Line != tc.line {
			t.Errorf("%s: line = %d", name, f.Line)
		}
	}
	got := run(t, spec(valid+"    OTHER: ${{ kit.args.ghost }} and ${{ kit.args.ghost }}\n"))
	if len(got) != 1 {
		t.Fatalf("one finding names one missing argument: %v", got)
	}
	got = run(t, spec("schemaVersion: '2'\nkind: mixin\nname: m\nextra: 1\n"))
	if got[0].Path != "kits/a/spec.yaml" {
		t.Fatalf("path = %s", got[0].Path)
	}
}

func TestSchemaLinesFollowTheKey(t *testing.T) {
	body := "schemaVersion: '2'\nkind: other\nname: m\n"
	f, ok := checktest.Find(run(t, spec(body)), "kit.schema")
	if !ok || f.Line != 2 {
		t.Fatalf("finding = %v", f)
	}
}
