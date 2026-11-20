// Package baseline records the pass rate of each case and compares a later run
// with that record, so a change that makes a case worse shows up even when the
// run as a whole still passes.
package baseline

import (
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// DefaultTolerance is how far a pass rate may fall before it counts as a
// regression. Model output varies between runs, so an exact match is too strict.
const DefaultTolerance = 0.1

// Record builds a baseline from the results of a finished run.
func Record(results model.Results, tolerance float64, recordedAt string) model.Baseline {
	cases := make([]model.BaselineCase, 0, len(results.Cases))
	for _, c := range results.Cases {
		runs := c.Runs
		if runs == 0 {
			runs = 1
		}
		cases = append(cases, model.BaselineCase{ID: c.ID, Condition: c.Condition, PassRate: rate(c), Runs: runs})
	}
	return model.Baseline{
		Version: model.Version, Protocol: results.Protocol, Profile: results.Profile,
		InputsDigest: results.InputsDigest, RecordedAt: recordedAt, Tolerance: tolerance, Cases: cases,
	}
}

func rate(c model.CaseResult) float64 {
	if c.Score != nil {
		return *c.Score
	}
	if c.Outcome == model.Passed {
		return 1
	}
	return 0
}

// Delta is the change of one case against its baseline.
type Delta struct {
	Key      string
	Baseline float64
	Current  float64
}

// Comparison is the difference between a baseline and a run.
type Comparison struct {
	// Regressions are the cases whose pass rate fell by more than the tolerance.
	Regressions []Delta
	// Improvements are the cases whose pass rate rose by more than the tolerance.
	Improvements []Delta
	// Missing are the baseline cases the run did not finish.
	Missing []string
	// Added are the finished cases the baseline does not list.
	Added []string
}

// Regressed reports whether the run is worse than the baseline: a case fell, or
// a case the baseline lists is missing.
func (c Comparison) Regressed() bool { return len(c.Regressions) > 0 || len(c.Missing) > 0 }

// Compare checks a run against a baseline.
func Compare(base model.Baseline, results model.Results) Comparison {
	current := results.Done()
	var out Comparison
	known := map[string]bool{}
	for _, b := range base.Cases {
		key := model.CaseKey(b.ID, b.Condition)
		known[key] = true
		c, ok := current[key]
		if !ok {
			out.Missing = append(out.Missing, key)
			continue
		}
		delta := Delta{Key: key, Baseline: b.PassRate, Current: rate(c)}
		switch {
		case delta.Current < b.PassRate-base.Tolerance:
			out.Regressions = append(out.Regressions, delta)
		case delta.Current > b.PassRate+base.Tolerance:
			out.Improvements = append(out.Improvements, delta)
		}
	}
	for key := range current {
		if !known[key] {
			out.Added = append(out.Added, key)
		}
	}
	sort.Strings(out.Added)
	sort.Strings(out.Missing)
	return out
}

// Store reads and writes a baseline file.
type Store struct {
	FS     fs.FileSystem
	Reader document.Reader
}

// Load returns the baseline in the file. The bool is false when there is none.
func (s Store) Load(file string) (model.Baseline, bool, error) {
	if ok, err := fs.Exists(s.FS, file); err != nil || !ok {
		return model.Baseline{}, false, err
	}
	b, err := s.Reader.Baseline(file)
	return b, err == nil, err
}

// Save writes the baseline to the file.
func (s Store) Save(file string, b model.Baseline) error {
	data, err := model.Encode(b)
	if err != nil {
		return err
	}
	return s.FS.WriteFile(file, data)
}
