package compare_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/compare"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/rules"
)

func source(mutate func(*sources.Source)) sources.Source {
	s := sources.Source{
		ID: "doc", Kind: sources.Web, Title: "Doc", URLs: []string{"https://e.test"}, ReviewedAt: "2026-06-01",
		Feeds:  []string{"references/a.md"},
		Claims: []sources.Claim{{ID: "c1", Statement: "s", Anchor: "x"}},
	}
	if mutate != nil {
		mutate(&s)
	}
	return s
}

func catalog(list ...sources.Source) sources.Catalog { return sources.Catalog{Sources: list} }

type want struct {
	code    string
	level   findings.Level
	mention string
}

func expect(t *testing.T, got []findings.Finding, wants ...want) {
	t.Helper()
	if len(got) != len(wants) {
		t.Fatalf("findings = %v; want %d", got, len(wants))
	}
	for i, w := range wants {
		if f := got[i]; f.Code != w.code || f.Level != w.level || f.Check != rules.CheckName || !strings.Contains(f.Message, w.mention) {
			t.Errorf("finding %d = %+v; want %+v", i, f, w)
		}
	}
}

func TestDiffOfUnchangedCatalogs(t *testing.T) {
	expect(t, compare.Diff(catalog(source(nil)), catalog(source(nil)), nil))
	expect(t, compare.Diff(sources.Catalog{}, catalog(source(nil)), []string{"references/a.md"}))
}

func TestDiffRemovedSource(t *testing.T) {
	expect(t, compare.Diff(catalog(source(nil)), sources.Catalog{}, nil), want{rules.CodeSourceRemoved, findings.Warn, `"doc" was in the base catalog`})
}

func TestDiffReviewDates(t *testing.T) {
	regressed := source(func(s *sources.Source) { s.ReviewedAt = "2026-05-01" })
	expect(t, compare.Diff(catalog(source(nil)), catalog(regressed), nil), want{rules.CodeReviewRegressed, findings.Fail, "2026-06-01 at the base ref and is dated 2026-05-01"})

	// A new review explains every other change of the source.
	reviewed := source(func(s *sources.Source) {
		s.ReviewedAt = "2026-11-01"
		s.Version = "2.0.0"
		s.Claims = nil
	})
	expect(t, compare.Diff(catalog(source(nil)), catalog(reviewed), []string{"references/a.md"}))
}

func TestDiffPinChangesNeedAReview(t *testing.T) {
	changes := map[string]func(*sources.Source){
		"urls":          func(s *sources.Source) { s.URLs = []string{"https://other.test"} },
		"path":          func(s *sources.Source) { s.Path = "p" },
		"ref":           func(s *sources.Source) { s.Ref = "v2" },
		"checkout":      func(s *sources.Source) { s.Checkout = "../up" },
		"version":       func(s *sources.Source) { s.Version = "2" },
		"commit":        func(s *sources.Source) { s.Commit = "abc1234" },
		"watch":         func(s *sources.Source) { s.Watch = []sources.Watch{{Path: "src"}} },
		"contentSha256": func(s *sources.Source) { s.ContentSHA256 = "sha256:x" },
		"outputSha256":  func(s *sources.Source) { s.OutputSHA256 = "sha256:y" },
	}
	for field, mutate := range changes {
		t.Run(field, func(t *testing.T) {
			expect(t, compare.Diff(catalog(source(nil)), catalog(source(mutate)), nil),
				want{rules.CodeMetadataChanged, findings.Drift, "changed " + field + " without a new review"})
		})
	}
	both := source(func(s *sources.Source) { s.Version = "2"; s.Commit = "abc1234" })
	expect(t, compare.Diff(catalog(source(nil)), catalog(both), nil), want{rules.CodeMetadataChanged, findings.Drift, "version, commit"})
}

func TestDiffClaimChangesNeedAReview(t *testing.T) {
	dropped := source(func(s *sources.Source) { s.Claims = nil })
	edited := source(func(s *sources.Source) { s.Claims = []sources.Claim{{ID: "c1", Statement: "s", Anchor: "y"}} })
	added := source(func(s *sources.Source) { s.Claims = append(s.Claims, sources.Claim{ID: "c2", Statement: "t"}) })
	expect(t, compare.Diff(catalog(source(nil)), catalog(dropped), nil), want{rules.CodeMetadataChanged, findings.Drift, "c1 removed"})
	expect(t, compare.Diff(catalog(source(nil)), catalog(edited), nil), want{rules.CodeMetadataChanged, findings.Drift, "c1 edited"})
	expect(t, compare.Diff(catalog(source(nil)), catalog(added), nil))
	both := source(func(s *sources.Source) { s.Claims = []sources.Claim{{ID: "c0", Statement: "n"}} })
	old := source(func(s *sources.Source) {
		s.Claims = []sources.Claim{{ID: "c2", Statement: "b"}, {ID: "c1", Statement: "a"}}
	})
	expect(t, compare.Diff(catalog(old), catalog(both), nil), want{rules.CodeMetadataChanged, findings.Drift, "c1 removed, c2 removed"})
}

func TestDiffFedFilesChangedWithoutAReview(t *testing.T) {
	expect(t, compare.Diff(catalog(source(nil)), catalog(source(nil)), []string{"references/a.md", "references/other.md"}),
		want{rules.CodeGitFileChanged, findings.Drift, "feeds changed: references/a.md"})
	expect(t, compare.Diff(catalog(source(nil)), catalog(source(nil)), []string{"references/other.md"}))
}

type base struct {
	catalog sources.Catalog
	changed []string
	catErr  error
	chgErr  error
}

func (b base) Catalog(string) (sources.Catalog, error) { return b.catalog, b.catErr }
func (b base) ChangedFiles(string) ([]string, error)   { return b.changed, b.chgErr }

func TestAgainst(t *testing.T) {
	got, err := compare.Against(base{catalog: catalog(source(nil)), changed: []string{"references/a.md"}}, "d", catalog(source(nil)))
	if err != nil {
		t.Fatal(err)
	}
	expect(t, got, want{rules.CodeGitFileChanged, findings.Drift, "references/a.md"})

	boom := errors.New("boom")
	if _, err := compare.Against(base{catErr: boom}, "d", sources.Catalog{}); !errors.Is(err, boom) {
		t.Errorf("catalog fault: %v", err)
	}
	if _, err := compare.Against(base{chgErr: boom}, "d", sources.Catalog{}); !errors.Is(err, boom) {
		t.Errorf("changed fault: %v", err)
	}
}
