// Package component holds what the checks of Markdown components share: a
// component is a Markdown file with YAML frontmatter that lives in a directory
// of a fixed name, such as a command, an agent or an output style.
package component

import (
	"errors"
	"path"
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// Files returns the Markdown files below the paths that sit in a directory
// named directory. A skill bundles reference documents of its own, so a file
// below a skills directory is never a component.
func Files(t tree.Tree, paths []string, directory string) ([]string, error) {
	return t.WalkAll(paths, func(file string) bool {
		if path.Ext(file) != ".md" {
			return false
		}
		segments := strings.Split(path.Dir(file), "/")
		return slices.Contains(segments, directory) && !slices.Contains(segments, "skills")
	})
}

// Spec names a kind of component.
type Spec struct {
	// Prefix starts every finding code of the check.
	Prefix string
	// Noun names one component in a message, for example "agent".
	Noun string
	// Known are the frontmatter keys that the harness reads.
	Known []string
	// Required are the keys that must hold a non-empty string.
	Required []string
	// Fields are the rules for the keys that may appear.
	Fields []Field
	// Body names what the body holds, for example "an agent prompt".
	Body string
}

// Field is one rule for the value of a frontmatter key.
type Field struct {
	Key     string
	Valid   func(value any) bool
	Message string
}

// Check reads one component file and reports what is wrong with its
// frontmatter. It returns the document, and false when the frontmatter is not
// readable and so nothing more can be checked.
func (s Spec) Check(report *check.Reporter, file, source string) (document.Document, bool) {
	doc, err := document.Parse(source)
	switch {
	case errors.Is(err, document.ErrNotMapping):
		report.Fail(s.Prefix+".frontmatter", file, 1, "frontmatter of %s must be a YAML mapping", s.Noun)
		return doc, false
	case err != nil:
		report.Fail(s.Prefix+".frontmatter", file, 1, "%v", err)
		return doc, false
	case !doc.HasFrontmatter:
		report.Fail(s.Prefix+".frontmatter", file, 1, "%s has no frontmatter: the harness reads its fields from it", s.Noun)
		return doc, false
	}
	for _, key := range doc.Keys() {
		if !slices.Contains(s.Known, key) {
			report.Fail(s.Prefix+".unknown-field", file, document.KeyLine(doc.Source, key),
				"unknown frontmatter field '%s': the harness ignores it, so whatever it configures is not configured", key)
		}
	}
	for _, key := range s.Required {
		value, ok := doc.String(key)
		if !ok || strings.TrimSpace(value) == "" {
			report.Fail(s.Prefix+"."+key, file, document.KeyLine(doc.Source, key), "%s is required and must be a non-empty string", key)
		}
	}
	for _, f := range s.Fields {
		if value, present := doc.Fields[f.Key]; present && !f.Valid(value) {
			report.Fail(s.Prefix+"."+f.Key, file, document.KeyLine(doc.Source, f.Key), "%s", f.Message)
		}
	}
	if s.Body != "" && strings.TrimSpace(doc.Body) == "" {
		report.Fail(s.Prefix+".body", file, doc.BodyLine, "frontmatter but no %s", s.Body)
	}
	return doc, true
}

// IsString accepts text.
func IsString(value any) bool {
	_, ok := value.(string)
	return ok
}

// IsBool accepts true and false.
func IsBool(value any) bool {
	_, ok := value.(bool)
	return ok
}

// IsPositiveInteger accepts an integer above zero.
func IsPositiveInteger(value any) bool {
	n, ok := value.(int)
	return ok && n > 0
}

// IsMapping accepts a YAML mapping.
func IsMapping(value any) bool {
	_, ok := value.(map[string]any)
	return ok
}

// IsListOrMapping accepts a YAML list or mapping.
func IsListOrMapping(value any) bool {
	if IsMapping(value) {
		return true
	}
	_, ok := value.([]any)
	return ok
}

// IsStrings accepts a non-empty string or a non-empty list of strings.
func IsStrings(value any) bool {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case []any:
		if len(v) == 0 {
			return false
		}
		for _, item := range v {
			if s, ok := item.(string); !ok || strings.TrimSpace(s) == "" {
				return false
			}
		}
		return true
	}
	return false
}

// OneOf accepts a string that is one of the allowed values.
func OneOf(allowed ...string) func(any) bool {
	return func(value any) bool {
		s, ok := value.(string)
		return ok && slices.Contains(allowed, s)
	}
}

// Words splits a value that is a string or a list into its entries. A string
// splits on commas, as a harness reads it.
func Words(value any) []string {
	switch v := value.(type) {
	case string:
		var out []string
		for _, part := range strings.Split(v, ",") {
			out = append(out, strings.TrimSpace(part))
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, _ := item.(string)
			out = append(out, strings.TrimSpace(s))
		}
		return out
	}
	return nil
}
