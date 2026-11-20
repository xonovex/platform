package text

import (
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

var (
	listMarker    = regexp.MustCompile(`^\s*(?:[-*+]|\d{1,9}[.)])\s`)
	headingLine   = regexp.MustCompile(`^\s{0,3}#{1,6}(?:\s|$)`)
	thematicBreak = regexp.MustCompile(`^\s{0,3}(?:(?:-\s*){3,}|(?:\*\s*){3,}|(?:_\s*){3,}|=+)\s*$`)
	definition    = regexp.MustCompile(`^\s{0,3}\[[^\]]+\]:\s`)
	quotePrefix   = regexp.MustCompile(`^\s{0,3}>\s?`)
	hardBreak     = regexp.MustCompile(`(?:\s{2,}|\\|<br\s*/?>)$`)
)

// frontmatterEnd returns the index of the line after the frontmatter, or 0
// when the file has none. A folded scalar there may wrap by design.
func frontmatterEnd(lines []string) int {
	if len(lines) == 0 || lines[0] != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return i + 1
		}
	}
	return 0
}

// starts reports whether a line opens a block of its own, so that it cannot
// continue the paragraph above it.
func starts(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" ||
		headingLine.MatchString(line) ||
		listMarker.MatchString(line) ||
		thematicBreak.MatchString(line) ||
		definition.MatchString(line) ||
		strings.HasPrefix(trimmed, "|") ||
		strings.HasPrefix(trimmed, "<") ||
		strings.HasPrefix(trimmed, "```") ||
		strings.HasPrefix(trimmed, "~~~")
}

// prose reports whether a line is a paragraph or list line that a following
// line can continue.
func prose(line string) bool {
	return !starts(line) || listMarker.MatchString(line)
}

// A line that is followed by a line of the same paragraph breaks the rule: a
// reader and a diff see one paragraph as several lines. A trailing hard break
// is deliberate and exempt, a masked fenced block reads as blank lines, and
// one finding names each paragraph.
func hardWraps(report *check.Reporter, file, source string) {
	lines := strings.Split(maskComments(markdown.MaskFences(source)), "\n")
	flagged := false
	for i := frontmatterEnd(lines); i+1 < len(lines); i++ {
		current, next := lines[i], lines[i+1]
		if quotePrefix.MatchString(current) && quotePrefix.MatchString(next) {
			current, next = quotePrefix.ReplaceAllString(current, ""), quotePrefix.ReplaceAllString(next, "")
		}
		if !prose(current) || starts(next) || hardBreak.MatchString(current) {
			flagged = false
			continue
		}
		if !flagged {
			report.Fail("text.hard-wrap", file, i+1, "hard-wrapped Markdown: put one paragraph or list item on one line")
			flagged = true
		}
	}
}

var (
	comment    = regexp.MustCompile(`(?s)<!--.*?-->`)
	nonNewline = regexp.MustCompile(`[^\n]`)
)

// maskComments blanks HTML comments, which a reader never sees as prose, and
// keeps the line structure.
func maskComments(text string) string {
	return comment.ReplaceAllStringFunc(text, func(span string) string {
		return nonNewline.ReplaceAllString(span, " ")
	})
}
