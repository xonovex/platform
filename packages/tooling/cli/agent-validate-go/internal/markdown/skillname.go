package markdown

import (
	"regexp"
	"sort"
	"strings"
)

// A skill name is a gerund and then its object, as in using-git. The gerund is
// what tells a skill name from another kebab-case token in prose.
const skillName = `[a-z0-9]+ing(?:-[a-z0-9]+)+`

// SkillNameRule is the pattern a skill name has, for a message.
const SkillNameRule = `^[a-z0-9]+ing(-[a-z0-9]+)+$`

var (
	skillNameWhole = regexp.MustCompile(`^` + skillName + `$`)
	nameToken      = regexp.MustCompile(`[A-Za-z0-9-]+`)
	boldName       = regexp.MustCompile(`\*\*(` + skillName + `)\*\*`)
	operationAfter = regexp.MustCompile(`^\s+operation\b`)
	kebab          = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// IsSkillName reports whether the name is a gerund and then its object.
func IsSkillName(name string) bool { return skillNameWhole.MatchString(name) }

// IsKebabCase reports whether the name is lower-case kebab-case.
func IsKebabCase(name string) bool { return kebab.MatchString(name) }

// SkillNameMentions returns each token of the text that has the form of a
// skill name, in lower case and in the order of the text.
func SkillNameMentions(text string) []string {
	var out []string
	for _, token := range nameToken.FindAllString(text, -1) {
		lower := strings.ToLower(token)
		if skillNameWhole.MatchString(lower) {
			out = append(out, lower)
		}
	}
	return out
}

// BoldSkillNames returns the skill names that the Markdown sets in bold, which
// is the notation of a handoff to a skill by name. A bold name before the word
// operation names an operation, and Markdown renders no bold inside code.
func BoldSkillNames(text string) []string {
	stripped := StripCode(text)
	var out []string
	for _, m := range boldName.FindAllStringSubmatchIndex(stripped, -1) {
		if operationAfter.MatchString(stripped[m[1]:]) {
			continue
		}
		out = append(out, stripped[m[2]:m[3]])
	}
	return out
}

// Sorted returns the distinct values in order.
func Sorted(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}
