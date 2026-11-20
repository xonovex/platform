package launch

import (
	"errors"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
)

// LatestRun is the id that names the newest run record.
const LatestRun = "latest"

// RecordReader is the port the records of past runs are read through.
type RecordReader interface {
	// List returns the records, newest first.
	List() ([]record.Record, error)
	Read(id string) (record.Record, error)
}

// errNoRecords marks a build or a machine that keeps no run records.
var errNoRecords = errors.New("this build keeps no run records")

// ListRuns returns the run records, newest first. records is nil for a build or
// a machine that keeps none.
func ListRuns(records RecordReader) ([]record.Record, error) {
	if records == nil {
		return nil, errNoRecords
	}
	return records.List()
}

// ReadRun returns the record the id names, or the newest one for LatestRun.
func ReadRun(records RecordReader, id string) (record.Record, error) {
	if records == nil {
		return record.Record{}, errNoRecords
	}
	if id != LatestRun {
		return records.Read(id)
	}
	all, err := records.List()
	if err != nil {
		return record.Record{}, err
	}
	if len(all) == 0 {
		return record.Record{}, errors.New("no run is recorded yet")
	}
	return all[0], nil
}

// RunIDs names the ids a person may ask for: LatestRun, then the id of each
// record, newest first. A failure to list the records names only LatestRun, and
// a build without records names none.
func RunIDs(records RecordReader) []string {
	if records == nil {
		return nil
	}
	ids := []string{LatestRun}
	if all, err := records.List(); err == nil {
		for _, r := range all {
			ids = append(ids, r.ID)
		}
	}
	return ids
}
