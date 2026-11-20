package skill

import (
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
)

var (
	headingAnywhere = regexp.MustCompile(`(?m)^#{1,6} `)
	fenceMarker     = regexp.MustCompile("(?m)^```([\\w+-]+)?[^\\S\\n]*$")
	gotchas         = regexp.MustCompile(`(?m)^## *Gotchas`)
)

func checkBody(report *check.Reporter, options Options, doc document.Document, file string) {
	lines := strings.Count(doc.Body, "\n")
	switch {
	case lines > options.BodyFailLines:
		report.Fail("skill.body-length", file, doc.BodyLine, "body: %d lines (>%d spec target)", lines, options.BodyFailLines)
	case lines > options.BodyWarnLines:
		report.Warn("skill.body-length", file, doc.BodyLine, "body: %d lines (>80%% of the %d-line target)", lines, options.BodyFailLines)
	}
	if !headingAnywhere.MatchString(doc.Body) {
		report.Fail("skill.body-headings", file, doc.BodyLine, "body: no headings found")
	}
	// A bare fence is either an opening without a language or a closing one, and
	// blocks come in pairs, so the unmarked openings are the bare count less the
	// tagged count.
	var bare, tagged int
	for _, m := range fenceMarker.FindAllStringSubmatch(doc.Body, -1) {
		if m[1] == "" {
			bare++
		} else {
			tagged++
		}
	}
	if unmarked := bare - tagged; unmarked > 0 {
		report.Warn("skill.body-fence-language", file, doc.BodyLine, "body: %d code block(s) appear to lack a language marker", unmarked)
	}
	if !gotchas.MatchString(doc.Body) {
		report.Warn("skill.body-gotchas", file, doc.BodyLine, "content: no '## Gotchas' section, so non-obvious environment facts have no home")
	}
}
