// Package legacy converts a SOURCES.md file into a source catalog. The
// conversion keeps every word of the file: fields become typed members, the
// prose that fits no field is kept as remarks or as a note, and a line that
// the converter does not know is kept verbatim.
package legacy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

// CheckName is the name findings of this package carry.
const CheckName = "convert"

// Finding codes of the conversion.
const (
	CodeReviewedMissing = "convert.reviewed-missing"
	CodeEmptySection    = "convert.empty-section"
)

const maxIDLength = 64

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Result is a converted catalog and the problems found on the way. A catalog
// with a finding must not be written.
type Result struct {
	Document sources.Document
	Findings []findings.Finding
}

// Convert turns the text of a SOURCES.md file into a catalog document. The file
// name only labels findings.
func Convert(text, file string) Result {
	title, preamble, sections := split(text)
	result := Result{Document: sources.Document{
		Schema:   schemas.Sources,
		Version:  1,
		Title:    title,
		Preamble: strings.TrimSpace(strings.Join(preamble, "\n")),
		Sources:  []sources.Source{},
	}}

	used := map[string]bool{}
	last := ""
	for _, s := range sections {
		b := parseBlock(s)
		if !b.isSource() {
			note, problem := noteOf(b, last, file)
			if problem != nil {
				result.Findings = append(result.Findings, *problem)
				continue
			}
			result.Document.Notes = append(result.Document.Notes, note)
			continue
		}
		source := sourceOf(b, used)
		if b.reviewed == "" {
			result.Findings = append(result.Findings, finding(CodeReviewedMissing, file,
				fmt.Sprintf("section %q has no **Last reviewed:** date", b.title)))
		}
		result.Document.Sources = append(result.Document.Sources, source)
		last = source.ID
	}
	return result
}

func noteOf(b *block, after, file string) (sources.Note, *findings.Finding) {
	text := strings.TrimSpace(strings.Join(b.body, "\n"))
	if text == "" {
		f := finding(CodeEmptySection, file, fmt.Sprintf("section %q has no content and no source field", b.title))
		return sources.Note{}, &f
	}
	return sources.Note{Title: b.title, Text: text, After: after}, nil
}

func sourceOf(b *block, used map[string]bool) sources.Source {
	s := sources.Source{
		ID:                  uniqueID(slug(b.title), used),
		Kind:                kindOf(b),
		Title:               b.title,
		Citation:            b.citation,
		URLs:                b.urls,
		URLNotes:            b.urlNotes,
		Checkout:            b.checkout,
		Version:             b.version,
		Commit:              b.commit,
		ContentSHA256:       b.contentSHA,
		OutputSHA256:        b.outputSHA,
		Provenance:          b.provenance,
		ReviewedAt:          b.reviewed,
		Feeds:               b.feeds,
		CoversAllReferences: b.covers,
		Usage:               b.usage,
		Remarks:             strings.TrimSpace(strings.Join(b.remarks, "\n")),
	}
	for _, w := range b.watch {
		s.Watch = append(s.Watch, sources.Watch{Path: w.path, Feeds: w.feeds})
	}
	for i, aspect := range b.aspects {
		s.Claims = append(s.Claims, sources.Claim{ID: fmt.Sprintf("claim-%d", i+1), Statement: aspect})
	}
	return s
}

// kindOf chooses how an audit reads the source: through its checkout when it
// has one, through its URLs when it has some, and not at all otherwise.
func kindOf(b *block) sources.Kind {
	switch {
	case b.checkout != "":
		return sources.Git
	case len(b.urls) > 0:
		return sources.Web
	default:
		return sources.Original
	}
}

// slug turns a heading into an identifier: lower-case words joined by hyphens,
// cut at a word boundary to fit the identifier length.
func slug(title string) string {
	id := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if len(id) > maxIDLength {
		id = id[:maxIDLength]
		if i := strings.LastIndex(id, "-"); i > 0 {
			id = id[:i]
		}
	}
	return strings.Trim(id, "-")
}

// uniqueID returns id, or id with a numeric suffix when it was used already.
func uniqueID(id string, used map[string]bool) string {
	if id == "" {
		id = "source"
	}
	candidate := id
	for n := 2; used[candidate]; n++ {
		suffix := fmt.Sprintf("-%d", n)
		base := id
		if len(base)+len(suffix) > maxIDLength {
			base = strings.TrimRight(base[:maxIDLength-len(suffix)], "-")
		}
		candidate = base + suffix
	}
	used[candidate] = true
	return candidate
}

func finding(code, file, message string) findings.Finding {
	return findings.Finding{Check: CheckName, Level: findings.Fail, Code: code, Path: file, Message: message}
}
