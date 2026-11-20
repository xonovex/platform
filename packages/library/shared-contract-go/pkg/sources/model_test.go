package sources_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

func catalog() sources.Catalog {
	return sources.Catalog{Sources: []sources.Source{
		{ID: "b", ReviewAfter: "2026-03-01"},
		{ID: "a", ReviewAfter: "2026-03-01"},
		{ID: "old", ReviewAfter: "2026-01-01"},
		{ID: "future", ReviewAfter: "2027-01-01"},
		{ID: "never"},
		{ID: "replaced", ReviewAfter: "2025-01-01", SupersededBy: "a"},
	}}
}

func ids(list []sources.Source) []string {
	var out []string
	for _, s := range list {
		out = append(out, s.ID)
	}
	return out
}

func TestOverdue(t *testing.T) {
	now := time.Date(2026, 3, 1, 23, 30, 0, 0, time.FixedZone("west", -5*3600))
	// 23:30 at UTC-5 is 04:30 UTC on 2026-03-02, so the review date of
	// 2026-03-01 has come in UTC terms.
	got := ids(catalog().Overdue(now))
	want := []string{"old", "a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Overdue = %v; want %v", got, want)
	}
}

func TestOverdueOnTheReviewDateItself(t *testing.T) {
	got := ids(catalog().Overdue(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	if !reflect.DeepEqual(got, []string{"old"}) {
		t.Fatalf("Overdue = %v", got)
	}
}

func TestOverdueNone(t *testing.T) {
	if got := (sources.Catalog{}).Overdue(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)); len(got) != 0 {
		t.Fatalf("Overdue = %v", got)
	}
}

func TestByID(t *testing.T) {
	if s, ok := catalog().ByID("old"); !ok || s.ID != "old" {
		t.Fatalf("ByID = %+v, %v", s, ok)
	}
	if _, ok := catalog().ByID("zzz"); ok {
		t.Fatal("unknown id must not be found")
	}
}
