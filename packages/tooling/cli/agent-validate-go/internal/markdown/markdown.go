// Package markdown holds the pure text functions that the checks share: lines,
// fenced code, links, headings and anchors, skill-name notation and the
// measures of prose. It reads no file.
package markdown

import (
	"net/url"
	"regexp"
	"strings"
)

// Lines splits text on every common newline and omits a terminal empty line.
func Lines(text string) []string {
	if text == "" {
		return nil
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	parts := strings.Split(text, "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// LineOf returns the 1-based line of a byte offset.
func LineOf(text string, offset int) int {
	if offset > len(text) {
		offset = len(text)
	}
	return strings.Count(text[:offset], "\n") + 1
}

// fenceOpen matches the opening line of a fenced block.
var fenceOpen = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")

// fenceSpans returns the line index ranges [start, end] of every complete
// fenced block, fences included. An unterminated fence opens no block.
func fenceSpans(lines []string) [][2]int {
	var spans [][2]int
	for i := 0; i < len(lines); i++ {
		m := fenceOpen.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		fence := m[1]
		for j := i + 1; j < len(lines); j++ {
			closing := strings.TrimLeft(lines[j], " ")
			if len(lines[j])-len(closing) <= 3 && strings.HasPrefix(closing, fence) {
				spans = append(spans, [2]int{i, j})
				i = j
				break
			}
		}
	}
	return spans
}

func blank(s string) string {
	out := []byte(s)
	for i, c := range out {
		if c != '\n' {
			out[i] = ' '
		}
	}
	return string(out)
}

// Mask blanks every fenced block and every code span with spaces and keeps the
// line structure and the offsets, so a code sample cannot read as prose or as
// a link. A span ends on its own line unless crossLines is true.
func Mask(text string, crossLines bool) string {
	masked := MaskFences(text)
	if crossLines {
		return blankSpans(masked)
	}
	parts := strings.Split(masked, "\n")
	for i, line := range parts {
		parts[i] = blankSpans(line)
	}
	return strings.Join(parts, "\n")
}

// MaskFences blanks every fenced block with spaces and keeps the line
// structure and the offsets.
func MaskFences(text string) string {
	lines := strings.Split(text, "\n")
	for _, span := range fenceSpans(lines) {
		for i := span[0]; i <= span[1]; i++ {
			lines[i] = blank(lines[i])
		}
	}
	return strings.Join(lines, "\n")
}

// blankSpans blanks the code spans of a text. A span opens at a run of
// backticks and closes at the next run of the same length.
func blankSpans(text string) string {
	out := []byte(text)
	for i := 0; i < len(text); {
		if text[i] != '`' {
			i++
			continue
		}
		start := i
		for i < len(text) && text[i] == '`' {
			i++
		}
		width := i - start
		closeAt := findRun(text, i, width)
		if closeAt < 0 {
			continue
		}
		for k := start; k < closeAt+width; k++ {
			if out[k] != '\n' {
				out[k] = ' '
			}
		}
		i = closeAt + width
	}
	return string(out)
}

// findRun returns the offset of the next run of exactly width backticks.
func findRun(text string, from, width int) int {
	for i := from; i < len(text); {
		if text[i] != '`' {
			i++
			continue
		}
		start := i
		for i < len(text) && text[i] == '`' {
			i++
		}
		if i-start == width {
			return start
		}
	}
	return -1
}

// StripCode removes fenced blocks and code spans, which Markdown renders
// without formatting.
func StripCode(text string) string {
	lines := strings.Split(text, "\n")
	spans := fenceSpans(lines)
	drop := map[int]bool{}
	for _, span := range spans {
		for i := span[0]; i <= span[1]; i++ {
			drop[i] = true
		}
	}
	kept := lines[:0:0]
	for i, line := range lines {
		if !drop[i] {
			kept = append(kept, line)
		}
	}
	joined := strings.Join(kept, "\n")
	masked := blankSpans(joined)
	out := make([]byte, 0, len(joined))
	for i := 0; i < len(joined); i++ {
		if masked[i] == ' ' && joined[i] != ' ' {
			continue
		}
		out = append(out, joined[i])
	}
	return string(out)
}

var wordFence = regexp.MustCompile("(?ms)^```.+?^```[^\\S\\n]*$")

// ProseWords counts the words that a reader sees: fenced code is excluded so
// that an example does not use the word budget of a file.
func ProseWords(text string) int {
	return len(strings.Fields(wordFence.ReplaceAllString(text, "")))
}

var linkPattern = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)(?:\s[^)]*)?\)`)

// Link is a Markdown link or image outside code.
type Link struct {
	Target string
	Line   int
}

// Links returns the links of a text, outside fenced blocks and code spans.
func Links(text string) []Link {
	masked := Mask(text, false)
	var links []Link
	for _, m := range linkPattern.FindAllStringSubmatchIndex(masked, -1) {
		links = append(links, Link{Target: masked[m[2]:m[3]], Line: LineOf(masked, m[0])})
	}
	return links
}

// Split separates a link target into its path and its fragment.
func Split(target string) (file, fragment string) {
	file, fragment, _ = strings.Cut(target, "#")
	return file, fragment
}

// IsExternal reports whether the target has a scheme or is absolute.
func IsExternal(target string) bool {
	if strings.HasPrefix(target, "/") {
		return true
	}
	colon := strings.Index(target, ":")
	if colon <= 0 {
		return false
	}
	for _, c := range target[:colon] {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

// Placeholder reports whether the target is an illustrative form: it holds a
// template marker or an elision.
func Placeholder(target string) bool {
	return strings.ContainsAny(target, "<{") || strings.Contains(target, "\u2026")
}

// Heading is an ATX heading.
type Heading struct {
	Level int
	Text  string
	Line  int
}

var headingPattern = regexp.MustCompile(`^(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)

// Headings returns the headings of a text, outside fenced blocks.
func Headings(text string) []Heading {
	masked := MaskFences(text)
	var out []Heading
	for i, line := range strings.Split(masked, "\n") {
		if m := headingPattern.FindStringSubmatch(line); m != nil {
			out = append(out, Heading{Level: len(m[1]), Text: m[2], Line: i + 1})
		}
	}
	return out
}

var slugDrop = regexp.MustCompile(`[^\p{L}\p{N}\s-]`)

// Slug returns the anchor of a heading as GitHub writes it.
func Slug(heading string) string {
	s := strings.ToLower(strings.TrimSpace(heading))
	s = strings.NewReplacer("`", "", "*", "", "_", "", "~", "").Replace(s)
	s = slugDrop.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(s), "-")
}

// Anchors returns every anchor that the headings of a text define. A repeated
// heading gains the suffixes -1, -2 and so on.
func Anchors(text string) map[string]bool {
	anchors := map[string]bool{}
	seen := map[string]int{}
	for _, h := range Headings(text) {
		slug := Slug(h.Text)
		name := slug
		if n := seen[slug]; n > 0 {
			name = slug + "-" + itoa(n)
		}
		seen[slug]++
		anchors[name] = true
	}
	return anchors
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}

// DecodeFragment decodes a link fragment and folds its case.
func DecodeFragment(fragment string) string {
	if decoded, err := url.PathUnescape(fragment); err == nil {
		fragment = decoded
	}
	return strings.ToLower(fragment)
}

// DecodePath decodes the path of a link target.
func DecodePath(target string) string {
	if decoded, err := url.PathUnescape(target); err == nil {
		return decoded
	}
	return target
}

// Shingles returns the sets of n consecutive words of the prose of a text,
// without fenced code. A word is a run of lower-case letters, digits and
// underscores.
func Shingles(text string, n int) map[string]struct{} {
	prose := wordFence.ReplaceAllString(strings.ToLower(text), " ")
	words := strings.FieldsFunc(prose, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_'
	})
	set := map[string]struct{}{}
	for i := 0; i+n <= len(words); i++ {
		set[strings.Join(words[i:i+n], " ")] = struct{}{}
	}
	return set
}
