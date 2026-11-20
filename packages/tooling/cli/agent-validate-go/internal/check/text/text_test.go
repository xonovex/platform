package text_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check/text"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/checktest"
)

// esc spells a unicode escape without writing the escaped character.
func esc(hex string) string { return "\\u" + hex }

func run(t *testing.T, files map[string]string, options ...any) []findings.Finding {
	t.Helper()
	ctx := checktest.Context(t, files, ".")
	if len(options) > 0 {
		ctx = checktest.WithOptions(t, ctx, options[0])
	}
	return checktest.Run(t, text.Check{}, ctx)
}

func TestIdentity(t *testing.T) {
	c := text.Check{}
	if c.Name() != "text" || c.Summary() == "" || c.DefaultPaths()[0] != "." {
		t.Fatal("identity")
	}
	ctx := checktest.WithOptions(t, checktest.Context(t, nil, "."), map[string]any{"x": 1})
	if _, err := c.Run(ctx); err == nil {
		t.Fatal("unknown option")
	}
}

func TestPunctuation(t *testing.T) {
	cases := map[string]string{
		"text.em-dash":                    "a — b",
		"text.escaped-em-dash":            `{"a": "` + esc("2014") + `"}`,
		"text.ellipsis":                   "wait…",
		"text.escaped-ellipsis":           `"` + esc("2026") + `"`,
		"text.smart-single-quote":         "it’s",
		"text.escaped-smart-single-quote": `"` + esc("2018") + `"`,
		"text.smart-double-quote":         "“quoted”",
		"text.escaped-smart-double-quote": `"` + esc("201C") + `"`,
	}
	for code, body := range cases {
		got := run(t, map[string]string{"notes.txt": "fine\n" + body + "\n"})
		f, ok := checktest.Find(got, code)
		if !ok || f.Line != 2 || f.Path != "notes.txt" {
			t.Errorf("%s: findings = %v", code, got)
		}
	}
	long := "a — " + strings.Repeat("x", 200)
	got := run(t, map[string]string{"a.md": long})
	if f, ok := checktest.Find(got, "text.em-dash"); !ok || strings.Contains(f.Message, strings.Repeat("x", 100)) {
		t.Fatalf("the excerpt is capped: %v", got)
	}
	if got := run(t, map[string]string{"a.md": "short — line"}); !strings.Contains(got[0].Message, "short") {
		t.Fatalf("message = %v", got)
	}
}

func TestFileSelection(t *testing.T) {
	bad := "a — b"
	files := map[string]string{
		"a.go":                   bad,
		"node_modules/x/a.md":    bad,
		".moon/tasks/a.yml":      bad,
		"pkg/coverage/report.md": bad,
		"pkg/ok.yaml":            "fine",
		"pkg/not.yaml":           bad,
	}
	got := run(t, files)
	if len(got) != 1 || got[0].Path != "pkg/not.yaml" {
		t.Fatalf("findings = %v", got)
	}
	got = run(t, map[string]string{"a.go": bad, "a.md": bad}, map[string]any{"extensions": []string{".go"}, "skip": []string{}})
	if len(got) != 1 || got[0].Path != "a.go" {
		t.Fatalf("custom extensions: %v", got)
	}
}

func wraps(t *testing.T, body string) []findings.Finding {
	t.Helper()
	return run(t, map[string]string{"doc.md": body})
}

func TestHardWrap(t *testing.T) {
	clean := []string{
		"# Title\n\nOne paragraph on one line.\n\nAnother.\n",
		"- item one\n- item two\n  - nested\n\n1. first\n2. second\n",
		"```\nline one\nline two\n```\n",
		"| a | b |\n| - | - |\n| 1 | 2 |\n",
		"Line with a hard break  \nsecond line\n",
		"Line with a backslash\\\nsecond line\n",
		"Line with a tag<br>\nsecond line\n",
		"---\ndescription: >\n  folded\n  scalar\n---\n\nBody.\n",
		"Intro:\n- a\n- b\n",
		"Text\n# Heading\nText\n",
		"[ref]: https://example.com\n[other]: https://example.org\n",
		"<div>\nhtml\n</div>\n",
		"Title\n=====\n",
		"> one quote line\n\nafter\n",
		"text\n***\ntext\n",
		"",
	}
	for i, body := range clean {
		if got := wraps(t, body); len(got) != 0 {
			t.Errorf("clean %d: %v", i, got)
		}
	}
	dirty := map[string]struct {
		body string
		line int
	}{
		"paragraph":   {"# T\n\nfirst line of the\nparagraph that wraps\n", 3},
		"list item":   {"- one item that\n  wraps here\n", 1},
		"quote":       {"> a quote that\n> wraps\n", 1},
		"after fence": {"```\ncode\n```\ntext that\nwraps\n", 4},
		"frontmatter": {"---\na: b\n---\n\nfirst\nsecond\n", 5},
		"numbered":    {"1. numbered item\n   continues\n", 1},
		"indented":    {"text\n    indented lazy continuation\n", 1},
	}
	for name, tc := range dirty {
		got := wraps(t, tc.body)
		f, ok := checktest.Find(got, "text.hard-wrap")
		if !ok || f.Line != tc.line {
			t.Errorf("%s: findings = %v", name, got)
		}
	}
	got := wraps(t, "a\nb\nc\nd\n")
	if len(got) != 1 {
		t.Fatalf("one finding names one paragraph: %v", got)
	}
	got = wraps(t, "a\nb\n\nc\nd\n")
	if len(got) != 2 {
		t.Fatalf("two paragraphs: %v", got)
	}
}

func TestHardWrapSwitches(t *testing.T) {
	files := map[string]string{"a.md": "a\nb\n", "a.txt": "a\nb\n"}
	if got := run(t, files, map[string]any{"wrap": false}); len(got) != 0 {
		t.Fatalf("wrap off: %v", got)
	}
	got := run(t, files)
	if len(got) != 1 || got[0].Path != "a.md" {
		t.Fatalf("only Markdown wraps: %v", got)
	}
	got = run(t, files, map[string]any{"wrapExtensions": []string{".txt"}})
	if len(got) != 1 || got[0].Path != "a.txt" {
		t.Fatalf("custom wrap extensions: %v", got)
	}
}

func TestExclude(t *testing.T) {
	files := map[string]string{"CHANGELOG.md": "a\nb\n", "pkg/CHANGELOG.md": "a\nb\n", "doc.md": "a\nb\n"}
	got := run(t, files)
	if len(got) != 1 || got[0].Path != "doc.md" {
		t.Fatalf("generated changelogs are left out: %v", got)
	}
	got = run(t, files, map[string]any{"exclude": []string{"doc.md"}})
	if len(got) != 2 {
		t.Fatalf("a custom list replaces the default: %v", got)
	}
	if _, err := (text.Check{}).Run(checktest.WithOptions(t, checktest.Context(t, files, "."), map[string]any{"exclude": []string{"a{"}})); err != nil {
		t.Fatalf("a brace is a literal: %v", err)
	}
}

func TestHardWrapIgnoresComments(t *testing.T) {
	body := "Text\n\n<!-- a comment that\n     wraps over lines -->\n\nMore\n"
	if got := wraps(t, body); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}
