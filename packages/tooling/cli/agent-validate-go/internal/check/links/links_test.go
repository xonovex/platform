package links_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/links"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

func run(t *testing.T, files map[string]string, options any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if options != nil {
		ctx = checktest.WithOptions(t, ctx, options)
	}
	return checktest.Run(t, links.Check{}, ctx)
}

var off = map[string]any{"skillRoots": []string{}}

func TestIdentity(t *testing.T) {
	c := links.Check{}
	if c.Name() != "links" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
}

func TestResolvesFilesDirectoriesAndAnchors(t *testing.T) {
	files := map[string]string{
		"docs/a.md":     "# A\n\n## Setup steps\n\n[b](b.md) [dir](sub) [anchor](b.md#target) [self](#setup-steps) [up](../README.md#top)\n",
		"docs/b.md":     "# Target\n",
		"docs/sub/x.md": "x",
		"README.md":     "# Top\n",
	}
	if got := run(t, files, off); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestReportsBrokenLinksAndAnchors(t *testing.T) {
	files := map[string]string{
		"docs/a.md":   "[gone](gone.md)\n[bad](b.md#nope)\n[self](#nope)\n[enc](b%20c.md)\n[escape](../../outside.md)\n[abs](/abs.md)\n[web](https://x.y/z)\n",
		"docs/b.md":   "# Target\n",
		"docs/b c.md": "# Spaced\n",
	}
	got := run(t, files, off)
	if len(got) != 3 || !checktest.Has(got, "links.missing-target") || !checktest.Has(got, "links.missing-anchor") {
		t.Fatalf("findings = %v", got)
	}
	if got[0].Line != 1 || got[0].Path != "docs/a.md" {
		t.Fatalf("finding = %+v", got[0])
	}
}

func TestAnchorsCanBeSwitchedOff(t *testing.T) {
	files := map[string]string{"a.md": "[x](b.md#nope)", "b.md": "# B\n"}
	got := run(t, files, map[string]any{"skillRoots": []string{}, "anchors": false})
	if len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestSkipsCodePlaceholdersAndNonMarkdownTargets(t *testing.T) {
	files := map[string]string{
		"a.md": "`[x](nope.md)`\n```\n[y](nope.md)\n```\n[p](<x>.md) [t](a.ts#L10)\n",
		"a.ts": "code",
	}
	if got := run(t, files, off); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestHandoffs(t *testing.T) {
	skill := func(body string) string { return "---\nname: x\n---\n" + body }
	files := map[string]string{
		"p/plugin-a/skills/using-a/SKILL.md":        skill("Hand to **using-b** and **using-gone** and **using-a**.\n"),
		"p/plugin-a/skills/using-a/references/r.md": "See **using-ref-gone**.\n",
		"p/plugin-a/README.md":                      "Use **using-b**.\n",
		"p/plugin-a/docs/guide.md":                  "Use **using-readme-gone**.\n",
		"p/plugin-b/skills/using-b/SKILL.md":        skill("# B\n"),
	}
	got := run(t, files, map[string]any{"skillRoots": []string{"p"}})
	if n := len(checktest.Codes(got)); n != 3 {
		t.Fatalf("findings = %v", got)
	}
	for _, f := range got {
		if f.Code != "links.handoff-missing" || f.Line == 0 {
			t.Fatalf("finding = %+v", f)
		}
	}
}

func TestRejectsUnknownOptions(t *testing.T) {
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})
	if _, err := (links.Check{}).Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}
