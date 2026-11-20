package skill

import (
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

var (
	reservedWord    = regexp.MustCompile(`(?i)anthropic|claude`)
	xmlTag          = regexp.MustCompile(`<[^>]+>`)
	usePhrase       = regexp.MustCompile(`(?i)^\s*use (?:this skill )?when`)
	triggersPhrase  = regexp.MustCompile(`(?i)triggers? on`)
	evenWhenPhrase  = regexp.MustCompile(`(?i)even when the user doesn'?t`)
	negativeRouting = regexp.MustCompile(`(?i)(?:^|[.!?]\s+)(?:skip\b|do not use\b|don't use\b|out[- ]of[- ]scope\b)`)
	allowedTools    = regexp.MustCompile(`^[\w():*,.\- ]+$`)
	knownFields     = map[string]bool{
		"name": true, "description": true, "license": true,
		"compatibility": true, "metadata": true, "allowed-tools": true,
	}
)

func checkFrontmatter(report *check.Reporter, options Options, doc document.Document, file, parent string) {
	line := func(key string) int { return document.KeyLine(doc.Source, key) }
	checkName(report, options, doc, file, parent, line("name"))
	checkDescription(report, options, doc, file, line("description"))
	if compat, ok := doc.Fields["compatibility"]; ok && compat != nil {
		if text, isText := compat.(string); !isText {
			report.Fail("skill.compatibility-type", file, line("compatibility"), "frontmatter: compatibility is %s, expected string", typed(compat))
		} else if n := utf8.RuneCountInString(text); n > 500 {
			report.Fail("skill.compatibility-length", file, line("compatibility"), "frontmatter: compatibility is %d chars (>500)", n)
		}
	}
	if license, ok := doc.Fields["license"]; ok && license != nil {
		if _, isText := license.(string); !isText {
			report.Warn("skill.license-type", file, line("license"), "frontmatter: license is %s, expected string", typed(license))
		}
	}
	if meta, ok := doc.Fields["metadata"]; ok && meta != nil && !isStringMap(meta) {
		report.Warn("skill.metadata-type", file, line("metadata"), "frontmatter: metadata should be a string-to-string map")
	}
	if tools, ok := doc.Fields["allowed-tools"]; ok && tools != nil {
		switch v := tools.(type) {
		case string:
			if !allowedTools.MatchString(v) {
				report.Warn("skill.allowed-tools-characters", file, line("allowed-tools"),
					"frontmatter: allowed-tools has unexpected characters (expected space-separated tokens like 'Bash(git:*) Read')")
			}
		default:
			report.Warn("skill.allowed-tools-type", file, line("allowed-tools"),
				"frontmatter: allowed-tools is %s, expected a space-separated string", typed(tools))
		}
	}
	var unknown []string
	for key := range doc.Fields {
		if !knownFields[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		report.Warn("skill.unknown-field", file, 1, "frontmatter: unknown top-level field(s): %s", strings.Join(unknown, ", "))
	}
}

func isStringMap(value any) bool {
	m, ok := value.(map[string]any)
	if !ok {
		return false
	}
	for _, v := range m {
		if _, isText := v.(string); !isText {
			return false
		}
	}
	return true
}

func checkName(report *check.Reporter, options Options, doc document.Document, file, parent string, line int) {
	value, present := doc.Fields["name"]
	if !present || value == nil || value == "" {
		report.Fail("skill.name-missing", file, 1, "frontmatter: missing 'name'")
		return
	}
	name, ok := value.(string)
	if !ok {
		report.Fail("skill.name-type", file, line, "frontmatter: name is %s, expected string", typed(value))
		return
	}
	switch {
	case utf8.RuneCountInString(name) > 64:
		report.Fail("skill.name-length", file, line, "frontmatter: name '%s' is %d chars (>64 spec limit)", name, utf8.RuneCountInString(name))
	case !markdown.IsKebabCase(name):
		report.Fail("skill.name-kebab", file, line, "frontmatter: name '%s' is not kebab-case (regex ^[a-z0-9]+(-[a-z0-9]+)*$)", name)
	case !markdown.IsSkillName(name):
		report.Fail("skill.name-gerund", file, line, "frontmatter: name '%s' is not a gerund and then its object (regex %s)", name, markdown.SkillNameRule)
	case name != parent:
		report.Fail("skill.name-directory", file, line, "frontmatter: name '%s' != parent directory '%s'", name, parent)
	}
	if reservedWord.MatchString(name) && !contains(options.VendorAdapters, name) {
		report.Fail("skill.name-reserved", file, line, "frontmatter: name '%s' uses a reserved word (anthropic/claude)", name)
	}
	if xmlTag.MatchString(name) {
		report.Fail("skill.name-xml", file, line, "frontmatter: name '%s' contains an XML tag", name)
	}
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func checkDescription(report *check.Reporter, options Options, doc document.Document, file string, line int) {
	value, present := doc.Fields["description"]
	if !present || value == nil || value == "" {
		report.Fail("skill.description-missing", file, 1, "frontmatter: missing 'description'")
		return
	}
	desc, ok := value.(string)
	if !ok {
		report.Fail("skill.description-type", file, line, "frontmatter: description is %s, expected string", typed(value))
		return
	}
	if n := utf8.RuneCountInString(desc); n > options.DescriptionLimit {
		report.Fail("skill.description-length", file, line, "description: %d chars (over the %d-character catalog limit)", n, options.DescriptionLimit)
	}
	if !usePhrase.MatchString(desc) {
		report.Warn("skill.description-use-when", file, line, "description: does not start with 'Use when...' or 'Use this skill when...', so agents may not trigger reliably")
	}
	if !triggersPhrase.MatchString(desc) {
		report.Warn("skill.description-triggers", file, line, "description: missing 'Triggers on...', so agents lose the trigger keyword list")
	}
	if !evenWhenPhrase.MatchString(desc) {
		report.Warn("skill.description-even-when", file, line, "description: no 'even when the user doesn't say...' clause, so it may miss implicit triggers")
	}
	if negativeRouting.MatchString(desc) {
		report.Fail("skill.description-negative", file, line, "description: routing must use positive triggers only; move skip and out-of-scope guidance to the body")
	}
	own, _ := doc.String("name")
	var others []string
	for _, mention := range markdown.SkillNameMentions(desc) {
		if mention != strings.ToLower(own) {
			others = append(others, mention)
		}
	}
	if others = markdown.Sorted(others); len(others) > 0 {
		report.Fail("skill.description-names-skill", file, line, "description: names other skill(s) %s; put by-name handoffs in the body", strings.Join(others, ", "))
	}
	if xmlTag.MatchString(desc) {
		report.Warn("skill.description-markup", file, line, "description: contains angle-bracket markup, which is fine for a component or generic reference and otherwise an XML tag to remove")
	}
}
