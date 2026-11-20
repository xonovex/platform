package sources_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		name    string
		sources []sources.Source
		want    []string
		level   findings.Level
		file    string
		mention string
	}{
		{
			name:    "valid catalog reports one pass",
			sources: []sources.Source{{ID: "a", ReviewedAt: "2026-01-01", Claims: []sources.Claim{{ID: "c", Anchor: "intro"}}}},
			want:    []string{sources.CodeOK},
			level:   findings.Pass,
		},
		{
			name:    "empty catalog reports one pass",
			sources: nil,
			want:    []string{sources.CodeOK},
			level:   findings.Pass,
		},
		{
			name:    "duplicate id names the first file",
			sources: []sources.Source{{ID: "a", File: "x/sources.json"}, {ID: "a", File: "x/sources.d/y.json"}},
			want:    []string{sources.CodeDuplicateID},
			level:   findings.Fail,
			file:    "x/sources.d/y.json",
			mention: "x/sources.json",
		},
		{
			name:    "self supersession",
			sources: []sources.Source{{ID: "a", SupersededBy: "a"}},
			want:    []string{sources.CodeSupersededSelf},
			level:   findings.Fail,
		},
		{
			name:    "unknown supersession",
			sources: []sources.Source{{ID: "a", SupersededBy: "ghost"}},
			want:    []string{sources.CodeSupersededUnset},
			level:   findings.Fail,
			mention: "ghost",
		},
		{
			name: "supersession cycle is reported once",
			sources: []sources.Source{
				{ID: "a", SupersededBy: "b"}, {ID: "b", SupersededBy: "c"}, {ID: "c", SupersededBy: "a"},
			},
			want:  []string{sources.CodeSupersededCycle},
			level: findings.Fail,
		},
		{
			name: "a chain into a cycle reports the cycle only",
			sources: []sources.Source{
				{ID: "tail", SupersededBy: "a"}, {ID: "a", SupersededBy: "b"}, {ID: "b", SupersededBy: "a"},
			},
			want:  []string{sources.CodeSupersededCycle},
			level: findings.Fail,
		},
		{
			name:    "a plain chain is fine",
			sources: []sources.Source{{ID: "a", SupersededBy: "b"}, {ID: "b"}},
			want:    []string{sources.CodeOK},
			level:   findings.Pass,
		},
		{
			name:    "review before the last review",
			sources: []sources.Source{{ID: "a", ReviewedAt: "2026-02-01", ReviewAfter: "2026-01-01"}},
			want:    []string{sources.CodeReviewOrder},
			level:   findings.Fail,
		},
		{
			name:    "review on the same day passes",
			sources: []sources.Source{{ID: "a", ReviewedAt: "2026-02-01", ReviewAfter: "2026-02-01"}},
			want:    []string{sources.CodeOK},
			level:   findings.Pass,
		},
		{
			name:    "duplicate claim",
			sources: []sources.Source{{ID: "a", Claims: []sources.Claim{{ID: "c", Anchor: "x"}, {ID: "c", Anchor: "y"}}}},
			want:    []string{sources.CodeDuplicateClaim},
			level:   findings.Fail,
		},
		{
			name:    "claim without anchor warns",
			sources: []sources.Source{{ID: "a", Claims: []sources.Claim{{ID: "c"}}}},
			want:    []string{sources.CodeClaimUnanchored},
			level:   findings.Warn,
		},
		{
			name:    "unanchored claims count once for each source",
			sources: []sources.Source{{ID: "a", Claims: []sources.Claim{{ID: "c"}, {ID: "d"}, {ID: "e", Anchor: "x"}}}},
			want:    []string{sources.CodeClaimUnanchored},
			level:   findings.Warn,
			mention: "2 of 3 claims",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sources.Check(sources.Catalog{Sources: tc.sources})
			if len(got) != len(tc.want) {
				t.Fatalf("findings = %v; want codes %v", got, tc.want)
			}
			for i, f := range got {
				if f.Code != tc.want[i] || f.Level != tc.level || f.Check != sources.CheckName {
					t.Errorf("finding %d = %+v; want code %s level %s", i, f, tc.want[i], tc.level)
				}
				if tc.file != "" && f.Path != tc.file {
					t.Errorf("path = %q; want %q", f.Path, tc.file)
				}
				if tc.mention != "" && !strings.Contains(f.Message, tc.mention) {
					t.Errorf("message %q lacks %q", f.Message, tc.mention)
				}
			}
		})
	}
}
