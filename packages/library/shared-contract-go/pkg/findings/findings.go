// Package findings is the result model shared by every validation, audit and
// eval check: a finding with a level, a summary per check, and the mapping from
// findings to a process exit code.
package findings

import (
	"errors"
	"fmt"
	"sort"
)

// Level ranks a finding. The order is pass, warn, drift, fail.
type Level string

const (
	// Pass records a check that held.
	Pass Level = "pass"
	// Warn records a problem that does not block a change.
	Warn Level = "warn"
	// Drift records a source or baseline that no longer matches what it feeds.
	Drift Level = "drift"
	// Fail records a violated rule.
	Fail Level = "fail"
)

// Levels lists every level from least to most severe.
func Levels() []Level { return []Level{Pass, Warn, Drift, Fail} }

// ErrUnknownLevel reports a level name that is not one of Levels.
var ErrUnknownLevel = errors.New("unknown level")

// ParseLevel returns the level a name denotes.
func ParseLevel(name string) (Level, error) {
	for _, level := range Levels() {
		if string(level) == name {
			return level, nil
		}
	}
	return "", fmt.Errorf("%w %q", ErrUnknownLevel, name)
}

// Rank returns the severity of a level, 0 for pass up to 3 for fail. An unknown
// level ranks -1.
func (l Level) Rank() int {
	for rank, level := range Levels() {
		if level == l {
			return rank
		}
	}
	return -1
}

// Finding is one result of one check.
type Finding struct {
	Check   string `json:"check"`
	Level   Level  `json:"level"`
	Code    string `json:"code"`
	Path    string `json:"path,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

// String renders a finding on one line in the form path:line: level code: message.
func (f Finding) String() string {
	location := f.Path
	if location != "" && f.Line > 0 {
		location = fmt.Sprintf("%s:%d", f.Path, f.Line)
	}
	if location != "" {
		location += ": "
	}
	return fmt.Sprintf("%s%s %s: %s", location, f.Level, f.Code, f.Message)
}

// Summary counts the findings of one check by level.
type Summary struct {
	Check string `json:"check"`
	Pass  int    `json:"pass"`
	Warn  int    `json:"warn"`
	Drift int    `json:"drift"`
	Fail  int    `json:"fail"`
}

// Total returns the number of findings the summary counts.
func (s Summary) Total() int { return s.Pass + s.Warn + s.Drift + s.Fail }

// Summarize counts findings per check. The result is sorted by check name.
func Summarize(all []Finding) []Summary {
	byCheck := map[string]*Summary{}
	for _, f := range all {
		s, ok := byCheck[f.Check]
		if !ok {
			s = &Summary{Check: f.Check}
			byCheck[f.Check] = s
		}
		switch f.Level {
		case Pass:
			s.Pass++
		case Warn:
			s.Warn++
		case Drift:
			s.Drift++
		case Fail:
			s.Fail++
		}
	}
	out := make([]Summary, 0, len(byCheck))
	for _, s := range byCheck {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Check < out[j].Check })
	return out
}

// Worst returns the most severe level among findings, or Pass when there are none.
func Worst(all []Finding) Level {
	worst := Pass
	for _, f := range all {
		if f.Level.Rank() > worst.Rank() {
			worst = f.Level
		}
	}
	return worst
}

// Exit codes of every command.
const (
	// ExitClean means no finding reached the policy threshold.
	ExitClean = 0
	// ExitFindings means at least one finding reached the policy threshold.
	ExitFindings = 1
	// ExitUsage means the command could not run: a bad flag, an unreadable
	// input or another infrastructure fault. A run that ends so has no verdict.
	ExitUsage = 2
)

// Policy sets the lowest level that makes a run fail.
type Policy struct {
	FailOn Level
}

// DefaultPolicy fails a run on drift and above, and lets a warning pass.
func DefaultPolicy() Policy { return Policy{FailOn: Drift} }

// StrictPolicy fails a run on a warning too.
func StrictPolicy() Policy { return Policy{FailOn: Warn} }

// ExitCode maps the outcome of a run to a process exit code. A non-nil err is an
// infrastructure fault and yields ExitUsage whatever the findings say, because
// partial results must never read as a pass.
func ExitCode(all []Finding, err error, policy Policy) int {
	if err != nil {
		return ExitUsage
	}
	threshold := policy.FailOn.Rank()
	if threshold < 1 {
		threshold = Drift.Rank()
	}
	for _, f := range all {
		if f.Level.Rank() >= threshold {
			return ExitFindings
		}
	}
	return ExitClean
}
