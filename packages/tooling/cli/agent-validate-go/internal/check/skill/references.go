package skill

import (
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/markdown"
)

var (
	refLink            = regexp.MustCompile(`references/([\w./<>{}-]+\.md)`)
	progressive        = regexp.MustCompile(`(?im)^## *Progressive Disclosure\s*$`)
	nextHeading        = regexp.MustCompile(`(?m)^## `)
	loadWhen           = regexp.MustCompile(`(?i)load when`)
	tocHeading         = regexp.MustCompile(`(?im)^## *(?:Contents|Table of Contents)\b`)
	externalReferences = regexp.MustCompile(`(?im)^## +External References\b`)
)

func checkReferenceLinks(ctx check.Context, report *check.Reporter, doc document.Document, file, dir string) {
	seen := map[string]bool{}
	var refs []string
	for _, m := range refLink.FindAllStringSubmatch(doc.Body, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			refs = append(refs, m[1])
		}
	}
	refs = markdown.Sorted(refs)
	for _, ref := range refs {
		if strings.ContainsAny(ref, "<{") {
			continue
		}
		line := lineOfText(doc, "references/"+ref)
		if !ctx.Tree.IsFile(path.Join(dir, "references", ref)) {
			report.Fail("skill.reference-broken", file, line, "references: broken link to references/%s", ref)
		}
		if strings.Contains(ref, "/") {
			report.Warn("skill.reference-nested", file, line, "references: 'references/%s' is nested deeper than one level", ref)
		}
		name := strings.TrimSuffix(path.Base(ref), ".md")
		if !markdown.IsKebabCase(name) {
			report.Warn("skill.reference-name", file, line, "references: 'references/%s' filename is not kebab-case", ref)
		}
	}
	if len(refs) == 0 {
		return
	}
	if strings.Contains(doc.Body, "@references/") {
		report.Fail("skill.reference-eager", file, lineOfText(doc, "@references/"),
			"references: '@references/' prefix found, which strips progressive disclosure; use plain 'references/'")
	}
	checkLoadTriggers(report, doc, file)
}

// lineOfText returns the file line of the first occurrence of text in the body.
func lineOfText(doc document.Document, text string) int {
	index := strings.Index(doc.Body, text)
	if index < 0 {
		return doc.BodyLine
	}
	return doc.BodyLine + strings.Count(doc.Body[:index], "\n")
}

// Only the Progressive Disclosure section tells an agent when to load a file.
func checkLoadTriggers(report *check.Reporter, doc document.Document, file string) {
	loc := progressive.FindStringIndex(doc.Body)
	if loc == nil {
		return
	}
	section := doc.Body[loc[1]:]
	if end := nextHeading.FindStringIndex(section); end != nil {
		section = section[:end[0]]
	}
	var links, missing int
	for _, line := range markdown.Lines(section) {
		if refLink.MatchString(line) {
			links++
			if !loadWhen.MatchString(line) {
				missing++
			}
		}
	}
	if missing > 0 {
		report.Warn("skill.reference-load-when", file, doc.BodyLine+strings.Count(doc.Body[:loc[0]], "\n"),
			"references: %d of %d link(s) in Progressive Disclosure lack a 'Load when...' trigger", missing, links)
	}
}

// A long reference file opens with a table of contents.
func checkReferenceTocs(report *check.Reporter, references []reference) {
	for _, r := range topLevel(references) {
		lines := markdown.Lines(r.text)
		if len(lines) <= 200 {
			continue
		}
		head := lines
		if len(head) > 15 {
			head = head[:15]
		}
		if !tocHeading.MatchString(strings.Join(head, "\n")) {
			report.Warn("skill.reference-toc", r.path, 1, "references: %d lines and no '## Contents' table of contents in the first lines", len(lines))
		}
	}
}

// A link in a top-level reference file resolves against its own directory.
func checkReferenceFileLinks(ctx check.Context, report *check.Reporter, dir string, references []reference) {
	root := path.Join(dir, "references")
	for _, r := range topLevel(references) {
		for _, link := range markdown.Links(r.text) {
			target, _ := markdown.Split(link.Target)
			if target == "" || hasScheme(link.Target) || markdown.Placeholder(link.Target) {
				continue
			}
			if !ctx.Tree.IsFile(path.Join(root, target)) {
				report.Fail("skill.reference-link-broken", r.path, link.Line, "reference links: broken link in references/%s to %s", r.name, link.Target)
			}
		}
	}
}

func hasScheme(target string) bool {
	lower := strings.ToLower(target)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "mailto:")
}

// External provenance belongs in the source catalog, which the source audit
// reads; a list of sources anywhere else escapes that audit.
func checkSourcePlacement(report *check.Reporter, doc document.Document, file string, references []reference) {
	if externalReferences.MatchString(doc.Body) {
		report.Fail("skill.source-placement", file, 1, "sources: '## External References' belongs in the source catalog, not SKILL.md")
	}
	for _, r := range references {
		if externalReferences.MatchString(r.text) {
			report.Fail("skill.source-placement", r.path, 1, "sources: '## External References' belongs in the source catalog, not references/%s", r.name)
		}
	}
}
