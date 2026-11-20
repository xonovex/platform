package artifact

import (
	"fmt"
	"strconv"
	"strings"
)

// Document is a Markdown file with optional front matter. Fields holds the
// top-level scalar keys of the front matter; a key with a nested value or a
// list has an empty value and is only present.
type Document struct {
	Fields         map[string]string
	Body           string
	HasFrontmatter bool
}

// ParseDocument splits text into its front matter and body. A text that does not
// start with a front matter fence is all body. An unterminated fence is an error.
func ParseDocument(text string) (Document, error) {
	lines, body, found, err := splitFrontmatter(text)
	if err != nil {
		return Document{}, err
	}
	document := Document{Fields: map[string]string{}, Body: body, HasFrontmatter: found}
	for _, line := range lines {
		if key, value, ok := topLevel(line); ok {
			document.Fields[key] = unquote(value)
		}
	}
	return document, nil
}

// DropFields returns text without the named top-level front matter keys and
// the lines that continue them. A text with no front matter is returned as is.
func DropFields(text string, keys ...string) string {
	lines, body, found, err := splitFrontmatter(text)
	if err != nil || !found || len(keys) == 0 {
		return text
	}
	kept := make([]string, 0, len(lines))
	dropping := false
	for _, line := range lines {
		if key, _, ok := topLevel(line); ok {
			dropping = contains(keys, key)
		}
		if !dropping {
			kept = append(kept, line)
		}
	}
	return "---\n" + strings.Join(kept, "\n") + fence(kept) + "---\n" + body
}

// fence returns the newline that ends the last front matter line, none for an
// empty front matter.
func fence(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return "\n"
}

func contains(keys []string, key string) bool {
	for _, candidate := range keys {
		if candidate == key {
			return true
		}
	}
	return false
}

// splitFrontmatter returns the lines between the leading fence and the closing
// fence, and the text after the closing fence.
func splitFrontmatter(text string) (lines []string, body string, found bool, err error) {
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		if rest, ok = strings.CutPrefix(text, "---\r\n"); !ok {
			return nil, text, false, nil
		}
	}
	for {
		line, after, more := strings.Cut(rest, "\n")
		if strings.TrimRight(line, "\r") == "---" {
			return lines, after, true, nil
		}
		if !more {
			return nil, "", false, fmt.Errorf("front matter is not closed by a --- line")
		}
		lines = append(lines, strings.TrimRight(line, "\r"))
		rest = after
	}
}

// topLevel reads a "key: value" line that starts in the first column.
func topLevel(line string) (key, value string, ok bool) {
	if line == "" || strings.ContainsRune(" \t-#", rune(line[0])) {
		return "", "", false
	}
	key, value, ok = strings.Cut(line, ":")
	if !ok || key == "" || strings.ContainsAny(key, " \t") {
		return "", "", false
	}
	return key, strings.TrimSpace(value), true
}

func unquote(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if first == '"' && last == '"' {
			if unquoted, err := strconv.Unquote(value); err == nil {
				return unquoted
			}
			return value[1 : len(value)-1]
		}
		if first == '\'' && last == '\'' {
			return strings.ReplaceAll(value[1:len(value)-1], "''", "'")
		}
	}
	return value
}
