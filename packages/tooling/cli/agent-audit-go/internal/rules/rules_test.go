package rules_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/rules"
)

var now = time.Date(2026, 11, 20, 15, 4, 5, 0, time.UTC)

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
		f := got[i]
		if f.Check != rules.CheckName || f.Code != w.code || f.Level != w.level || !strings.Contains(f.Message, w.mention) {
			t.Errorf("finding %d = %+v; want %+v", i, f, w)
		}
	}
}

func TestReview(t *testing.T) {
	windows := rules.Windows{Default: 180, Versioned: 90}
	cases := []struct {
		name   string
		source sources.Source
		want   []want
	}{
		{"fresh", sources.Source{ID: "a", ReviewedAt: "2026-11-01"}, nil},
		{"on the edge of the window", sources.Source{ID: "a", ReviewedAt: "2026-05-24"}, nil},
		{"one day past the window", sources.Source{ID: "a", ReviewedAt: "2026-05-23"}, []want{{rules.CodeReviewOverdue, findings.Drift, "181 days ago, the window is 180"}}},
		{"a pinned version has the shorter window", sources.Source{ID: "a", Version: "1.0.0", ReviewedAt: "2026-08-01"}, []want{{rules.CodeReviewOverdue, findings.Drift, "the window is 90"}}},
		{"review date has come", sources.Source{ID: "a", ReviewedAt: "2026-11-01", ReviewAfter: "2026-11-20"}, []want{{rules.CodeReviewOverdue, findings.Drift, "a review was due on 2026-11-20"}}},
		{"review date is ahead", sources.Source{ID: "a", ReviewedAt: "2026-11-01", ReviewAfter: "2026-11-21"}, nil},
		{"both reasons make one finding", sources.Source{ID: "a", ReviewedAt: "2025-01-01", ReviewAfter: "2025-06-01"}, []want{{rules.CodeReviewOverdue, findings.Drift, "the window is 180 days; a review was due"}}},
		{"superseded", sources.Source{ID: "a", ReviewedAt: "2020-01-01", SupersededBy: "b"}, nil},
		{"an unreadable date is skipped", sources.Source{ID: "a", ReviewedAt: "soon"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { expect(t, rules.Review(tc.source, windows, now), tc.want...) })
	}
}

func TestReviewIgnoresTheTimeOfDay(t *testing.T) {
	late := time.Date(2026, 11, 20, 23, 59, 59, 0, time.UTC)
	source := sources.Source{ID: "a", ReviewedAt: "2026-05-24"}
	if got := rules.Review(source, rules.Windows{Default: 180}, late); len(got) != 0 {
		t.Fatalf("findings = %v", got)
	}
}

func TestMapping(t *testing.T) {
	exists := func(p string) bool { return p == "references/a.md" }
	expect(t, rules.Mapping(sources.Source{ID: "s", Feeds: []string{"references/a.md"}}, exists))
	expect(t, rules.Mapping(sources.Source{ID: "s", CoversAllReferences: true}, exists))
	expect(t, rules.Mapping(sources.Source{ID: "s"}, exists), want{rules.CodeReferenceMissing, findings.Fail, "names no file"})
	expect(t, rules.Mapping(sources.Source{ID: "s", Feeds: []string{"references/a.md", "references/b.md"}}, exists),
		want{rules.CodeReferenceDangling, findings.Fail, "references/b.md"})
}

func TestPin(t *testing.T) {
	web := []string{"https://e.test"}
	expect(t, rules.Pin(sources.Source{ID: "s", URLs: web}))
	expect(t, rules.Pin(sources.Source{ID: "s", Version: "1.0.0", Provenance: "here"}))
	expect(t, rules.Pin(sources.Source{ID: "s", Version: "1.0.0", URLs: web, ContentSHA256: "sha256:x"}))
	expect(t, rules.Pin(sources.Source{ID: "s", Version: "1.0.0", URLs: web, Checkout: "c", Commit: "abc1234", Watch: []sources.Watch{{Path: "src"}}}))
	for name, s := range map[string]sources.Source{
		"nothing":          {ID: "s", Version: "1.0.0", URLs: web},
		"no commit":        {ID: "s", Version: "1.0.0", URLs: web, Checkout: "c", Watch: []sources.Watch{{Path: "src"}}},
		"no watched paths": {ID: "s", Version: "1.0.0", URLs: web, Checkout: "c", Commit: "abc1234"},
	} {
		t.Run(name, func(t *testing.T) {
			expect(t, rules.Pin(s), want{rules.CodePinMissing, findings.Fail, "pins version 1.0.0"})
		})
	}
}

func TestCoverage(t *testing.T) {
	files := []string{"references/a.md", "references/b.md", "references/c.md"}
	catalog := sources.Catalog{Sources: []sources.Source{
		{ID: "one", Feeds: []string{"references/a.md"}},
		{ID: "two", Watch: []sources.Watch{{Path: "src", Feeds: []string{"references/b.md"}}}},
	}}
	expect(t, rules.Coverage(catalog, files), want{rules.CodeReferenceUncovered, findings.Fail, "references/c.md"})
	expect(t, rules.Coverage(catalog, files[:2]))
	catalog.Sources = append(catalog.Sources, sources.Source{ID: "all", CoversAllReferences: true})
	expect(t, rules.Coverage(catalog, files))
	expect(t, rules.Coverage(sources.Catalog{}, files), want{rules.CodeNoSources, findings.Fail, "no source"})
}

func doc(locator string, status observe.Status, body string) observe.Document {
	d := observe.Document{Locator: locator, Status: status, Detail: "HTTP"}
	if body != "" {
		d.Body = []byte(body)
		d.SHA256 = observe.Hash(d.Body)
	}
	return d
}

func web(docs ...observe.Document) []observe.Observation {
	return []observe.Observation{{Transport: "http", Documents: docs}}
}

func TestObservedProviderFindings(t *testing.T) {
	s := sources.Source{ID: "s"}
	got := rules.Observed(s, web(doc("https://a", observe.Missing, ""), doc("https://b", observe.Failed, ""), doc("https://c", observe.Restricted, ""), doc("https://d", observe.OK, "x")))
	expect(t, got,
		want{rules.CodeProviderError, findings.Fail, "https://a is missing"},
		want{rules.CodeProviderError, findings.Fail, "https://b is error"},
		want{rules.CodeProviderRestricted, findings.Warn, "https://c refused"},
	)
	expect(t, rules.Observed(s, nil))
	fault := rules.ProviderFault(s, "git", errors.New("not installed"))
	if fault.Code != rules.CodeProviderError || fault.Level != findings.Fail || !strings.Contains(fault.Message, "the git transport failed: not installed") {
		t.Errorf("fault = %+v", fault)
	}
}

func TestObservedContentDigest(t *testing.T) {
	a := doc("https://a", observe.OK, "alpha")
	digest, _ := observe.Aggregate([]observe.Document{a})
	expect(t, rules.Observed(sources.Source{ID: "s", ContentSHA256: digest}, web(a)))
	expect(t, rules.Observed(sources.Source{ID: "s"}, web(a)))
	expect(t, rules.Observed(sources.Source{ID: "s", ContentSHA256: digest}, nil))

	other, _ := observe.Aggregate([]observe.Document{doc("https://a", observe.OK, "beta")})
	expect(t, rules.Observed(sources.Source{ID: "s", ContentSHA256: other}, web(a)),
		want{rules.CodeMetadataChanged, findings.Drift, "the content changed"})
	expect(t, rules.Observed(sources.Source{ID: "s", ContentSHA256: digest}, web(doc("https://a", observe.Restricted, ""))),
		want{rules.CodeProviderRestricted, findings.Warn, "refused"},
		want{rules.CodeContentUnverified, findings.Warn, "not every document"})
}

func TestObservedAnchors(t *testing.T) {
	body := "# Getting Started\n\n## The `Hook` API: part 2\n\n<h3 id=\"custom-id\">Plain <b>heading</b></h3>\n<a name=\"legacy\"></a>\nA phrase about retries.\n"
	claims := func(anchors ...string) sources.Source {
		s := sources.Source{ID: "s"}
		for i, a := range anchors {
			s.Claims = append(s.Claims, sources.Claim{ID: "c" + string(rune('1'+i)), Statement: "x", Anchor: a})
		}
		return s
	}
	present := claims("getting-started", "#the-hook-api-part-2", "custom-id", "plain-heading", "legacy", "A phrase about retries", "L10-L20", "42", "", "GETTING-STARTED")
	expect(t, rules.Observed(present, web(doc("https://a", observe.OK, body))))

	expect(t, rules.Observed(claims("getting-started", "gone"), web(doc("https://a", observe.OK, body))),
		want{rules.CodeAnchorMissing, findings.Drift, `the anchor "gone" of claim "c2"`})
	expect(t, rules.Observed(claims("gone"), web(doc("https://a", observe.Restricted, ""))), want{rules.CodeProviderRestricted, findings.Warn, "refused"})
	expect(t, rules.Observed(claims("gone"), web(doc("https://a", observe.OK, "other"), doc("https://b", observe.OK, body))),
		want{rules.CodeAnchorMissing, findings.Drift, "gone"})
	expect(t, rules.Observed(claims("custom-id"), web(doc("https://a", observe.OK, "other"), doc("https://b", observe.OK, body))))
}

func TestObservedGit(t *testing.T) {
	s := sources.Source{ID: "s"}
	git := func(g observe.Git) []observe.Observation {
		g.Checkout = "../up"
		return []observe.Observation{{Transport: "git", Git: &g}}
	}
	expect(t, rules.Observed(s, git(observe.Git{Resolved: true, Ref: "abc"})))
	expect(t, rules.Observed(s, git(observe.Git{})), want{rules.CodeProviderError, findings.Fail, "../up is missing"})
	expect(t, rules.Observed(s, git(observe.Git{Resolved: true, PullFailed: true})), want{rules.CodeProviderError, findings.Fail, "refreshing the tags"})
	expect(t, rules.Observed(s, git(observe.Git{Resolved: true, NoRef: true})), want{rules.CodePinMissing, findings.Fail, "neither a commit nor a version"})
	expect(t, rules.Observed(s, git(observe.Git{Resolved: true, RefMissing: true, Ref: "abc"})), want{rules.CodeProviderError, findings.Fail, "pinned ref abc"})
	expect(t, rules.Observed(s, git(observe.Git{Resolved: true, Ref: "abc", Behind: true, PinnedVersion: "1.0.0", LatestVersion: "v2.0.0"})),
		want{rules.CodeBehindLatestTag, findings.Drift, "pinned 1.0.0, the latest release is v2.0.0"})
	expect(t, rules.Observed(s, git(observe.Git{Resolved: true, Ref: "abc", CommitCount: 2})),
		want{rules.CodeGitFileChanged, findings.Drift, "2 commits changed the watched paths since abc"})
	got := rules.Observed(s, git(observe.Git{Resolved: true, Ref: "abc", CommitCount: 1, ReviewFeeds: []string{"references/a.md", "references/b.md"}}))
	expect(t, got, want{rules.CodeGitFileChanged, findings.Drift, "review references/a.md, references/b.md"})
}
