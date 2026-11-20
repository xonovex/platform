package markdown_test

import (
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

func TestLines(t *testing.T) {
	cases := map[string][]string{
		"":         nil,
		"a":        {"a"},
		"a\n":      {"a"},
		"a\r\nb\r": {"a", "b"},
		"a\n\n":    {"a", ""},
	}
	for in, want := range cases {
		if got := markdown.Lines(in); !reflect.DeepEqual(got, want) {
			t.Errorf("Lines(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLineOf(t *testing.T) {
	if markdown.LineOf("a\nb\nc", 4) != 3 || markdown.LineOf("a", 99) != 1 {
		t.Fatal("LineOf counts newlines before the offset and clamps the offset")
	}
}

func TestMaskBlanksFencesAndSpans(t *testing.T) {
	text := "see `[a](b)` and\n```go\n[c](d)\n```\n[e](f)\n"
	masked := markdown.Mask(text, false)
	if len(masked) != len(text) {
		t.Fatal("masking keeps the length")
	}
	links := markdown.Links(text)
	if len(links) != 1 || links[0].Target != "f" || links[0].Line != 5 {
		t.Fatalf("links = %+v", links)
	}
}

func TestMaskLeavesUnterminatedFenceAndSpan(t *testing.T) {
	text := "```\n[a](b)\n`open\n[c](d)\n"
	if got := markdown.Links(text); len(got) != 2 {
		t.Fatalf("an unterminated fence and span mask nothing: %+v", got)
	}
}

func TestMaskSpanAcrossLines(t *testing.T) {
	text := "`a\nb` [x](y)"
	if masked := markdown.Mask(text, true); masked[0] != ' ' || masked[2] != '\n' || masked[3] != ' ' {
		t.Fatal("a span may cross lines when asked")
	}
	if markdown.Mask(text, false)[:2] != "`a" {
		t.Fatal("a span ends on its line by default")
	}
}

func TestFenceClosingNeedsTheSameFence(t *testing.T) {
	text := "````\n```\n[a](b)\n````\n[c](d)"
	links := markdown.Links(text)
	if len(links) != 1 || links[0].Target != "d" {
		t.Fatalf("links = %+v", links)
	}
	tilde := "~~~\n[a](b)\n~~~\n[c](d)"
	if got := markdown.Links(tilde); len(got) != 1 {
		t.Fatalf("tilde fence: %+v", got)
	}
	indented := "    ```\n[a](b)"
	if got := markdown.Links(indented); len(got) != 1 {
		t.Fatalf("four spaces do not open a fence: %+v", got)
	}
}

func TestStripCode(t *testing.T) {
	text := "keep `drop` this\n```\nfenced\n```\nand **x**"
	got := markdown.StripCode(text)
	if got != "keep  this\nand **x**" {
		t.Fatalf("StripCode = %q", got)
	}
}

func TestProseWords(t *testing.T) {
	text := "one two\n```\nthree four five\n```\nsix"
	if got := markdown.ProseWords(text); got != 3 {
		t.Fatalf("ProseWords = %d", got)
	}
}

func TestLinksWithTitleAndImage(t *testing.T) {
	links := markdown.Links(`![alt](img.png "title") and [t](a.md#x)`)
	if len(links) != 2 || links[0].Target != "img.png" || links[1].Target != "a.md#x" {
		t.Fatalf("links = %+v", links)
	}
	file, fragment := markdown.Split("a.md#x")
	if file != "a.md" || fragment != "x" {
		t.Fatal("Split cuts at the first hash")
	}
}

func TestExternalAndPlaceholder(t *testing.T) {
	for target, want := range map[string]bool{
		"https://x": true, "mailto:a@b": true, "/abs": true, "a/b.md": false, "c:/x": true, "1a:b": false, "#frag": false,
	} {
		if markdown.IsExternal(target) != want {
			t.Errorf("IsExternal(%q) != %v", target, want)
		}
	}
	if !markdown.Placeholder("a/<x>.md") || !markdown.Placeholder("a/{x}.md") || !markdown.Placeholder("a\u2026b") || markdown.Placeholder("a/b.md") {
		t.Fatal("Placeholder")
	}
}

func TestHeadingsAndAnchors(t *testing.T) {
	text := "# Title\n## A `b` c\n## A `b` c\n```\n# not\n```\n### Done ###\n"
	headings := markdown.Headings(text)
	if len(headings) != 4 || headings[3].Text != "Done" || headings[3].Level != 3 {
		t.Fatalf("headings = %+v", headings)
	}
	anchors := markdown.Anchors(text)
	for _, want := range []string{"title", "a-b-c", "a-b-c-1", "done"} {
		if !anchors[want] {
			t.Errorf("missing anchor %q in %v", want, anchors)
		}
	}
	if markdown.Slug("Hello, World!") != "hello-world" {
		t.Fatal("Slug drops punctuation")
	}
	if markdown.DecodeFragment("A%20B") != "a b" || markdown.DecodeFragment("%zz") != "%zz" {
		t.Fatal("DecodeFragment")
	}
	if markdown.DecodePath("a%20b") != "a b" || markdown.DecodePath("%zz") != "%zz" {
		t.Fatal("DecodePath")
	}
}

func TestAnchorsCountTenRepeats(t *testing.T) {
	text := ""
	for i := 0; i < 12; i++ {
		text += "# same\n"
	}
	if !markdown.Anchors(text)["same-10"] {
		t.Fatal("suffixes run past nine")
	}
}

func TestShingles(t *testing.T) {
	set := markdown.Shingles("One two three.\n```\nx y z\n```\nfour five six", 3)
	if len(set) != 4 {
		t.Fatalf("shingles = %v", set)
	}
	if _, ok := set["one two three"]; !ok {
		t.Fatal("lower-cased words")
	}
	if len(markdown.Shingles("a b", 5)) != 0 {
		t.Fatal("a short text has no shingle")
	}
}

func TestSkillNames(t *testing.T) {
	if !markdown.IsSkillName("using-git") || markdown.IsSkillName("git") || markdown.IsSkillName("using") || markdown.IsSkillName("Using-git") {
		t.Fatal("IsSkillName")
	}
	if !markdown.IsKebabCase("a-b1") || markdown.IsKebabCase("a_b") || markdown.IsKebabCase("-a") {
		t.Fatal("IsKebabCase")
	}
	got := markdown.SkillNameMentions("Use using-git, then Writing-Python; not x-using-git or using-git- ok")
	if !reflect.DeepEqual(got, []string{"using-git", "writing-python"}) {
		t.Fatalf("mentions = %v", got)
	}
	bold := markdown.BoldSkillNames("see **using-git** and **writing-python** operation, `**in-code-using**` ```\n**fenced-using-x**\n```")
	if !reflect.DeepEqual(bold, []string{"using-git"}) {
		t.Fatalf("bold = %v", bold)
	}
	if !reflect.DeepEqual(markdown.Sorted([]string{"b", "a", "b"}), []string{"a", "b"}) {
		t.Fatal("Sorted")
	}
}
