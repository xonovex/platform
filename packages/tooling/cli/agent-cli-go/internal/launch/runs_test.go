package launch

import (
	"errors"
	"slices"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
)

// fakeRecords holds run records newest first, and fails each read with err.
type fakeRecords struct {
	records []record.Record
	err     error
}

func (f fakeRecords) List() ([]record.Record, error) { return f.records, f.err }

func (f fakeRecords) Read(id string) (record.Record, error) {
	for _, r := range f.records {
		if r.ID == id {
			return r, nil
		}
	}
	return record.Record{}, errors.New("no record " + id)
}

func twoRuns() fakeRecords {
	return fakeRecords{records: []record.Record{{ID: "new"}, {ID: "old"}}}
}

func TestListRunsNeedsAStore(t *testing.T) {
	if _, err := ListRuns(nil); err == nil {
		t.Error("ListRuns() without a store error = nil")
	}
	if got, err := ListRuns(twoRuns()); err != nil || len(got) != 2 {
		t.Errorf("ListRuns() = %v, %v", got, err)
	}
}

func TestReadRunNamesARecordByIDOrAsTheLatest(t *testing.T) {
	if got, err := ReadRun(twoRuns(), "old"); err != nil || got.ID != "old" {
		t.Errorf("ReadRun(old) = %v, %v", got.ID, err)
	}
	if got, err := ReadRun(twoRuns(), LatestRun); err != nil || got.ID != "new" {
		t.Errorf("ReadRun(latest) = %v, %v, want the newest", got.ID, err)
	}
	for name, records := range map[string]RecordReader{
		"no store":   nil,
		"no record":  fakeRecords{},
		"list error": fakeRecords{err: errors.New("unreadable")},
	} {
		if _, err := ReadRun(records, LatestRun); err == nil {
			t.Errorf("%s: ReadRun(latest) error = nil", name)
		}
	}
}

func TestRunIDsOfferTheLatestThenEachRecord(t *testing.T) {
	if got := RunIDs(twoRuns()); !slices.Equal(got, []string{LatestRun, "new", "old"}) {
		t.Errorf("RunIDs() = %v", got)
	}
	if got := RunIDs(fakeRecords{err: errors.New("unreadable")}); !slices.Equal(got, []string{LatestRun}) {
		t.Errorf("RunIDs() that cannot list = %v", got)
	}
	if got := RunIDs(nil); got != nil {
		t.Errorf("RunIDs() without a store = %v", got)
	}
}
