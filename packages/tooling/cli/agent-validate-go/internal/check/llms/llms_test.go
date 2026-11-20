package llms_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/llms"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, llms.Check{}, ctx)
}

func TestIdentity(t *testing.T) {
	c := llms.Check{}
	if c.Name() != "llms" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	if _, err := c.Run(checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})); err == nil {
		t.Fatal("unknown option")
	}
}

const index = "\n# Project\n\n> Summary\n\n- [Guide](/guide.md): the guide\n- [Rel](docs/rel.md#part)\n- [Site](https://example.com/x)\n- [Here](#top)\n"

func TestValidIndex(t *testing.T) {
	files := map[string]string{"site/llms.txt": index, "site/guide.md": "x", "site/docs/rel.md": "x"}
	if got := run(t, files); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
	files = map[string]string{"llms.txt": index, "public/guide.md": "x", "docs/rel.md": "x"}
	got := run(t, files, map[string]any{"siteRoot": "public"})
	if len(got) != 0 {
		t.Fatalf("site root: %v", got)
	}
}

func TestInvalidIndex(t *testing.T) {
	cases := map[string]struct {
		text string
		code string
		line int
	}{
		"empty":      {"", "llms.heading", 1},
		"no h1":      {"\nIntro\n- [A](a.md)\n", "llms.heading", 2},
		"h2":         {"## Project\n- [A](a.md)\n", "llms.heading", 1},
		"not md":     {"# P\n- [A](a.html)\n", "llms.target-format", 2},
		"missing":    {"# P\n- [A](gone.md)\n", "llms.target-missing", 2},
		"no links":   {"# P\ntext\n", "llms.links", 1},
		"only other": {"# P\n[x](https://e.com)\n", "", 0},
	}
	for name, tc := range cases {
		got := run(t, map[string]string{"llms.txt": tc.text})
		if tc.code == "" {
			if len(got) != 0 {
				t.Errorf("%s: %v", name, got)
			}
			continue
		}
		f, ok := checktest.Find(got, tc.code)
		if !ok || f.Line != tc.line {
			t.Errorf("%s: findings = %v", name, got)
		}
	}
}
