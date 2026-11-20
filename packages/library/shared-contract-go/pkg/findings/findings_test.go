package findings_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
)

func TestParseLevel(t *testing.T) {
	for _, level := range findings.Levels() {
		got, err := findings.ParseLevel(string(level))
		if err != nil || got != level {
			t.Fatalf("ParseLevel(%q) = %q, %v", level, got, err)
		}
	}
	if _, err := findings.ParseLevel("error"); !errors.Is(err, findings.ErrUnknownLevel) {
		t.Fatalf("err = %v", err)
	}
}

func TestRank(t *testing.T) {
	want := map[findings.Level]int{findings.Pass: 0, findings.Warn: 1, findings.Drift: 2, findings.Fail: 3, "other": -1}
	for level, rank := range want {
		if got := level.Rank(); got != rank {
			t.Errorf("Rank(%q) = %d; want %d", level, got, rank)
		}
	}
}

func TestFindingString(t *testing.T) {
	cases := []struct {
		name string
		in   findings.Finding
		want string
	}{
		{"path and line", findings.Finding{Level: findings.Fail, Code: "a.b", Path: "x.md", Line: 3, Message: "bad"}, "x.md:3: fail a.b: bad"},
		{"path only", findings.Finding{Level: findings.Warn, Code: "a.b", Path: "x.md", Message: "bad"}, "x.md: warn a.b: bad"},
		{"no path", findings.Finding{Level: findings.Drift, Code: "a.b", Message: "bad"}, "drift a.b: bad"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.String(); got != tc.want {
				t.Fatalf("String() = %q; want %q", got, tc.want)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	all := []findings.Finding{
		{Check: "b", Level: findings.Pass},
		{Check: "a", Level: findings.Fail},
		{Check: "a", Level: findings.Warn},
		{Check: "a", Level: findings.Drift},
		{Check: "a", Level: findings.Fail},
		{Check: "b", Level: findings.Pass},
	}
	want := []findings.Summary{
		{Check: "a", Warn: 1, Drift: 1, Fail: 2},
		{Check: "b", Pass: 2},
	}
	got := findings.Summarize(all)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Summarize = %+v; want %+v", got, want)
	}
	if got[0].Total() != 4 {
		t.Fatalf("Total = %d", got[0].Total())
	}
	if findings.Summarize(nil) == nil {
		t.Fatal("Summarize(nil) must return an empty slice, not nil")
	}
}

func TestWorst(t *testing.T) {
	if got := findings.Worst(nil); got != findings.Pass {
		t.Fatalf("Worst(nil) = %q", got)
	}
	all := []findings.Finding{{Level: findings.Warn}, {Level: findings.Fail}, {Level: findings.Drift}}
	if got := findings.Worst(all); got != findings.Fail {
		t.Fatalf("Worst = %q", got)
	}
}

func TestExitCode(t *testing.T) {
	errInfra := errors.New("cannot read input")
	cases := []struct {
		name   string
		all    []findings.Finding
		err    error
		policy findings.Policy
		want   int
	}{
		{"no findings", nil, nil, findings.DefaultPolicy(), findings.ExitClean},
		{"pass only", []findings.Finding{{Level: findings.Pass}}, nil, findings.DefaultPolicy(), findings.ExitClean},
		{"warn passes by default", []findings.Finding{{Level: findings.Warn}}, nil, findings.DefaultPolicy(), findings.ExitClean},
		{"warn fails when strict", []findings.Finding{{Level: findings.Warn}}, nil, findings.StrictPolicy(), findings.ExitFindings},
		{"drift fails by default", []findings.Finding{{Level: findings.Drift}}, nil, findings.DefaultPolicy(), findings.ExitFindings},
		{"fail fails", []findings.Finding{{Level: findings.Fail}}, nil, findings.DefaultPolicy(), findings.ExitFindings},
		{"zero policy behaves as default", []findings.Finding{{Level: findings.Warn}}, nil, findings.Policy{}, findings.ExitClean},
		{"pass threshold is not a pass-everything switch", []findings.Finding{{Level: findings.Pass}}, nil, findings.Policy{FailOn: findings.Pass}, findings.ExitClean},
		{"error beats clean findings", nil, errInfra, findings.DefaultPolicy(), findings.ExitUsage},
		{"error beats failing findings", []findings.Finding{{Level: findings.Fail}}, errInfra, findings.DefaultPolicy(), findings.ExitUsage},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := findings.ExitCode(tc.all, tc.err, tc.policy); got != tc.want {
				t.Fatalf("ExitCode = %d; want %d", got, tc.want)
			}
		})
	}
}
