// Package document reads the YAML frontmatter of a Markdown file. It is a pure
// parser: the text comes in and the fields come out.
package document

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var frontmatterPattern = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)

// Document is a Markdown file split into its frontmatter and its body.
type Document struct {
	// HasFrontmatter is false for a file without a frontmatter block, which
	// reads as an empty mapping so that each missing field is reported.
	HasFrontmatter bool
	// Source is the frontmatter as written.
	Source string
	// Fields is the parsed frontmatter mapping.
	Fields map[string]any
	// Body is the text after the frontmatter.
	Body string
	// BodyLine is the 1-based line of the file on which the body starts.
	BodyLine int
}

// ErrNotMapping reports a frontmatter block that is not a mapping.
var ErrNotMapping = errors.New("frontmatter is not a mapping")

// Parse splits a Markdown file. An error means the frontmatter exists but is
// not a YAML mapping.
func Parse(text string) (Document, error) {
	m := frontmatterPattern.FindStringSubmatch(text)
	if m == nil {
		return Document{Fields: map[string]any{}, Body: text, BodyLine: 1}, nil
	}
	doc := Document{
		HasFrontmatter: true,
		Source:         m[1],
		Body:           m[2],
		BodyLine:       strings.Count(m[1], "\n") + 4,
	}
	var parsed any
	if err := yaml.Unmarshal([]byte(m[1]), &parsed); err != nil {
		return doc, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}
	switch fields := parsed.(type) {
	case nil:
		doc.Fields = map[string]any{}
	case map[string]any:
		doc.Fields = fields
	default:
		return doc, ErrNotMapping
	}
	return doc, nil
}

// KeyLine returns the 1-based line of the file on which the frontmatter key is
// written, or 1 when it is not written.
func KeyLine(source, key string) int {
	for i, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, key+":") {
			return i + 2
		}
	}
	return 1
}

// TypeName names the type of a parsed value in the vocabulary of a YAML
// author.
func TypeName(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case int, int64, uint64:
		return "int"
	case float64:
		return "float"
	case string:
		return "str"
	case []any:
		return "list"
	case map[string]any:
		return "dict"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// Keys returns the keys of the mapping in sorted order.
func (d Document) Keys() []string {
	keys := make([]string, 0, len(d.Fields))
	for key := range d.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// String returns the value of a field that holds text, and whether it does.
func (d Document) String(key string) (string, bool) {
	s, ok := d.Fields[key].(string)
	return s, ok
}
