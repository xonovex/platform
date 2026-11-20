package legacy_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/legacy"
)

const web = `# Sources

Intro line.

Second intro line.

## Go Spec

- **URL:** https://go.dev/ref/spec
- **Last reviewed:** 2026-03-01
- **Content SHA256:** ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789
- **Used for:** the language rules in references/syntax.md
- **Aspects extracted:**
  - Declarations
  - Types

Some prose about the spec.

Another paragraph.

## Maintenance

How the catalog is kept.
`

func convert(t *testing.T, text string) legacy.Result {
	t.Helper()
	return legacy.Convert(text, "SOURCES.md")
}

func TestConvertWebSource(t *testing.T) {
	result := convert(t, web)
	if len(result.Findings) != 0 {
		t.Fatalf("findings = %v", result.Findings)
	}
	doc := result.Document
	if doc.Schema != schemas.Sources || doc.Version != 1 || doc.Title != "Sources" || doc.Preamble != "Intro line.\n\nSecond intro line." {
		t.Errorf("document = %+v", doc)
	}
	if len(doc.Sources) != 1 {
		t.Fatalf("sources = %+v", doc.Sources)
	}
	s := doc.Sources[0]
	if s.ID != "go-spec" || s.Kind != sources.Web || !reflect.DeepEqual(s.URLs, []string{"https://go.dev/ref/spec"}) || len(s.URLNotes) != 0 {
		t.Errorf("source = %+v", s)
	}
	if s.ReviewedAt != "2026-03-01" || s.ContentSHA256 != "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789" {
		t.Errorf("review or digest = %+v", s)
	}
	if !reflect.DeepEqual(s.Feeds, []string{"references/syntax.md"}) || !reflect.DeepEqual(s.Usage, []string{"the language rules in references/syntax.md"}) {
		t.Errorf("feeds or usage = %+v", s)
	}
	if len(s.Claims) != 2 || s.Claims[0].ID != "claim-1" || s.Claims[1].Statement != "Types" {
		t.Errorf("claims = %+v", s.Claims)
	}
	if s.Remarks != "Some prose about the spec.\n\nAnother paragraph." {
		t.Errorf("remarks = %q", s.Remarks)
	}
	if len(doc.Notes) != 1 || doc.Notes[0].Title != "Maintenance" || doc.Notes[0].After != "go-spec" {
		t.Errorf("notes = %+v", doc.Notes)
	}
}

func TestConvertedCatalogIsValid(t *testing.T) {
	data, err := sources.Encode(convert(t, web).Document)
	if err != nil {
		t.Fatal(err)
	}
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	violations, err := validator.Validate(schemas.Sources, data)
	if err != nil || len(violations) > 0 {
		t.Fatalf("violations = %v, err = %v", violations, err)
	}
}

const git = `# Sources

## Upstream Tool

- **Citation title:** ignored line
- **Title:** The Tool Book
- **Checkout:** ../../upstream
- **Version:** ` + "`1.2.3`" + `
- **Commit:** abc1234
- **Watch:**
  - src/cli -> references/cli.md, references/flags.md
  - docs -> other.md
- **CLI output SHA256:** 0000000000000000000000000000000000000000000000000000000000000000
- **Provenance:** read from the checkout
- **Last reviewed:** 2026-02-02
- **References:** cli.md, ` + "`flags.md`" + `
- **Also used for:** every reference
`

func TestConvertGitSource(t *testing.T) {
	result := convert(t, git)
	if len(result.Findings) != 0 {
		t.Fatalf("findings = %v", result.Findings)
	}
	s := result.Document.Sources[0]
	if s.Kind != sources.Git || s.Checkout != "../../upstream" || s.Version != "1.2.3" || s.Commit != "abc1234" || s.Citation != "The Tool Book" {
		t.Errorf("source = %+v", s)
	}
	wantWatch := []sources.Watch{
		{Path: "src/cli", Feeds: []string{"references/cli.md", "references/flags.md"}},
		{Path: "docs", Feeds: []string{"references/other.md"}},
	}
	if !reflect.DeepEqual(s.Watch, wantWatch) {
		t.Errorf("watch = %+v", s.Watch)
	}
	if s.OutputSHA256 != "sha256:"+strings.Repeat("0", 64) || !s.CoversAllReferences {
		t.Errorf("digest or coverage = %+v", s)
	}
	if !reflect.DeepEqual(s.Feeds, []string{"references/cli.md", "references/flags.md", "references/other.md"}) {
		t.Errorf("feeds = %v", s.Feeds)
	}
	if !strings.Contains(s.Remarks, "Citation title") {
		t.Errorf("an unknown field must be kept: %q", s.Remarks)
	}
}

func TestReferencesAll(t *testing.T) {
	for _, value := range []string{"all", "All references", "- all"} {
		text := "## S\n\n- **URL:** https://e.test\n- **Last reviewed:** 2026-01-01\n- **References:** " + value + "\n"
		if s := convert(t, text).Document.Sources[0]; !s.CoversAllReferences || len(s.Feeds) != 0 {
			t.Errorf("%q: %+v", value, s)
		}
	}
}

func TestAnnotatedAndMultipleURLs(t *testing.T) {
	text := `## Docs

- **URLs:**
  - https://a.test/x.
  - https://b.test/y (the guide)
  - https://a.test/x
  - https://c.test/(z)
- **Last reviewed:** 2026-01-01
`
	s := convert(t, text).Document.Sources[0]
	if !reflect.DeepEqual(s.URLs, []string{"https://a.test/x", "https://b.test/y", "https://c.test/(z)"}) {
		t.Errorf("urls = %v", s.URLs)
	}
	if !reflect.DeepEqual(s.URLNotes, []string{"https://a.test/x.", "https://b.test/y (the guide)"}) {
		t.Errorf("urlNotes = %v", s.URLNotes)
	}
}

func TestSourceWithoutURLIsOriginal(t *testing.T) {
	text := "## Book\n\n- **Provenance:** written by hand\n- **Last reviewed:** 2026-01-01\n"
	s := convert(t, text).Document.Sources[0]
	if s.Kind != sources.Original || s.Provenance != "written by hand" {
		t.Errorf("source = %+v", s)
	}
	noteOnly := "## Book\n\n- **URL:** see the shelf\n- **Last reviewed:** 2026-01-01\n"
	s = convert(t, noteOnly).Document.Sources[0]
	if s.Kind != sources.Original || len(s.URLNotes) != 1 {
		t.Errorf("source = %+v", s)
	}
}

func TestRepeatedAndMalformedFieldsAreKept(t *testing.T) {
	text := `## S

- **URL:** https://e.test
- **Last reviewed:** 2026-01-01
- **Last reviewed:** 2026-02-02
- **Content SHA256:** not-a-digest
- **Content SHA256:**
- **Watch:** nonsense without arrow
- **Title:**
- **Mystery:** value
`
	s := convert(t, text).Document.Sources[0]
	if s.ReviewedAt != "2026-01-01" || s.ContentSHA256 != "" {
		t.Errorf("source = %+v", s)
	}
	for _, want := range []string{"2026-02-02", "not-a-digest", "nonsense without arrow", "Mystery"} {
		if !strings.Contains(s.Remarks, want) {
			t.Errorf("remarks lost %q: %q", want, s.Remarks)
		}
	}
}

func TestSecondDigestOfTheSameFieldIsKept(t *testing.T) {
	d := strings.Repeat("a", 64)
	text := "## S\n\n- **URL:** https://e.test\n- **Last reviewed:** 2026-01-01\n- **Content SHA256:** " + d + "\n- **Content SHA256:** " + strings.Repeat("b", 64) + "\n"
	s := convert(t, text).Document.Sources[0]
	if s.ContentSHA256 != "sha256:"+d || !strings.Contains(s.Remarks, strings.Repeat("b", 64)) {
		t.Errorf("source = %+v", s)
	}
}

func TestIdentifiersAreUniqueAndBounded(t *testing.T) {
	long := strings.Repeat("word ", 30)
	text := "## " + long + "\n\n- **URL:** https://e.test/1\n- **Last reviewed:** 2026-01-01\n\n## " + long + "\n\n- **URL:** https://e.test/2\n- **Last reviewed:** 2026-01-01\n\n## !!!\n\n- **URL:** https://e.test/3\n- **Last reviewed:** 2026-01-01\n"
	got := convert(t, text).Document.Sources
	if len(got) != 3 {
		t.Fatalf("sources = %d", len(got))
	}
	for _, s := range got {
		if len(s.ID) > 64 || s.ID == "" {
			t.Errorf("id = %q", s.ID)
		}
	}
	if got[0].ID == got[1].ID || got[2].ID != "source" {
		t.Errorf("ids = %q %q %q", got[0].ID, got[1].ID, got[2].ID)
	}
	// A suffix must fit when the base is at the limit.
	if !strings.HasSuffix(got[1].ID, "-2") {
		t.Errorf("second id = %q", got[1].ID)
	}
}

func TestFindings(t *testing.T) {
	text := "## No date\n\n- **URL:** https://e.test\n\n## Empty\n\n"
	result := convert(t, text)
	codes := []string{}
	for _, f := range result.Findings {
		if f.Level != findings.Fail || f.Check != legacy.CheckName || f.Path != "SOURCES.md" {
			t.Errorf("finding = %+v", f)
		}
		codes = append(codes, f.Code)
	}
	if !reflect.DeepEqual(codes, []string{legacy.CodeReviewedMissing, legacy.CodeEmptySection}) {
		t.Errorf("codes = %v", codes)
	}
}

func TestNoteBeforeTheFirstSource(t *testing.T) {
	text := "## Policy\n\nBe kind.\n\n## S\n\n- **URL:** https://e.test\n- **Last reviewed:** 2026-01-01\n"
	doc := convert(t, text).Document
	if len(doc.Notes) != 1 || doc.Notes[0].After != "" || doc.Notes[0].Text != "Be kind." {
		t.Errorf("notes = %+v", doc.Notes)
	}
}

func TestWindowsLineEndings(t *testing.T) {
	text := strings.ReplaceAll("# T\n\n## S\n\n- **URL:** https://e.test\n- **Last reviewed:** 2026-01-01\n", "\n", "\r\n")
	if doc := convert(t, text).Document; doc.Title != "T" || len(doc.Sources) != 1 || doc.Sources[0].URLs[0] != "https://e.test" {
		t.Errorf("document = %+v", doc)
	}
}
