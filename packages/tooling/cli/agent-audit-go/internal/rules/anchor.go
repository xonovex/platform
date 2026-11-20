package rules

import (
	"regexp"
	"strings"
)

var (
	lineRange  = regexp.MustCompile(`^L?\d+(-L?\d+)?$`)
	mdHeading  = regexp.MustCompile(`(?m)^ {0,3}#{1,6}[ \t]+(.+?)[ \t#]*$`)
	htmlID     = regexp.MustCompile(`(?i)\b(?:id|name)\s*=\s*["']([^"']+)["']`)
	htmlHead   = regexp.MustCompile(`(?is)<h[1-6][^>]*>(.*?)</h[1-6]>`)
	htmlTag    = regexp.MustCompile(`<[^>]*>`)
	slugOmit   = regexp.MustCompile(`[^\p{L}\p{N}\s_-]`)
	whitespace = regexp.MustCompile(`\s+`)
)

// anchorPresent reports whether the anchor of a claim can be found in the
// content of its source. The anchor may be a heading slug, an element id, or a
// phrase of the text. A line range such as L10-L20 names a place no digest of
// the headings can check, so it counts as present.
func anchorPresent(body []byte, anchor string) bool {
	anchor = strings.TrimPrefix(strings.TrimSpace(anchor), "#")
	if anchor == "" || lineRange.MatchString(anchor) {
		return true
	}
	text := string(body)
	want := strings.ToLower(anchor)
	for _, candidate := range anchors(text) {
		if candidate == want {
			return true
		}
	}
	return strings.Contains(strings.ToLower(text), want)
}

// anchors returns the lower-case slugs of the Markdown and HTML headings and
// the ids of the HTML elements of a document.
func anchors(text string) []string {
	var out []string
	for _, m := range mdHeading.FindAllStringSubmatch(text, -1) {
		out = append(out, slug(m[1]))
	}
	for _, m := range htmlHead.FindAllStringSubmatch(text, -1) {
		out = append(out, slug(htmlTag.ReplaceAllString(m[1], "")))
	}
	for _, m := range htmlID.FindAllStringSubmatch(text, -1) {
		out = append(out, strings.ToLower(m[1]))
	}
	return out
}

// slug turns a heading into the fragment a renderer gives it: lower case, with
// punctuation dropped and runs of white space turned into one hyphen.
func slug(heading string) string {
	lowered := strings.ToLower(strings.TrimSpace(heading))
	return whitespace.ReplaceAllString(slugOmit.ReplaceAllString(lowered, ""), "-")
}
