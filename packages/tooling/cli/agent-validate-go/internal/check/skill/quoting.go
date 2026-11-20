package skill

import (
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

var quotedScalar = regexp.MustCompile(`(?i)^[ \t]*[a-z_][\w-]*:[ \t]*(.+?)[ \t]*$`)

// The skill loader reads frontmatter more naively than YAML does, so a scalar
// that YAML accepts can still end at an inner quote.
func checkLoaderQuoting(report *check.Reporter, doc document.Document, file string) {
	if !doc.HasFrontmatter {
		return
	}
	lines := markdown.Lines(doc.Source)
	for i, line := range lines {
		if !strings.HasPrefix(line, "description") {
			continue
		}
		key, rest, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(key) != "description" {
			continue
		}
		raw := strings.TrimSpace(rest)
		if len(raw) < 2 || !strings.HasPrefix(raw, `"`) || !strings.HasSuffix(raw, `"`) {
			report.Fail("skill.description-quoting", file, i+2,
				"frontmatter: 'description' must use one double-quoted scalar; use single quotes for phrases inside it")
		}
		break
	}
	for i, line := range lines {
		m := quotedScalar.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		rest := m[1]
		if len(rest) < 2 || (rest[0] != '\'' && rest[0] != '"') {
			continue
		}
		quote := rest[:1]
		// A multiline or unterminated scalar cannot be judged from one line.
		if !strings.HasSuffix(rest, quote) {
			continue
		}
		key := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
		switch {
		case quote == "'" && strings.Contains(rest, "''"):
			report.Fail("skill.quoting-escape", file, i+2,
				"frontmatter: '%s' uses single-quote '' escaping, which the skill loader does not support; use a double-quoted scalar with literal apostrophes", key)
		case strings.Count(rest, quote) != 2:
			report.Fail("skill.quoting-inner", file, i+2,
				"frontmatter: '%s' has an inner %s inside a %s-quoted value; the skill loader stops at the first inner quote; use a double-quoted scalar with single quotes for inner phrases", key, quote, quote)
		}
	}
}
