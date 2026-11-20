// Package report turns the outcome of a run into text or JSON and into the
// process exit code. The exit code tells a script what happened without parsing
// the text: 0 pass, 1 fail, 2 invalid, 3 stopped by a budget with work left.
package report

import (
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/run"
)

// Exit codes.
const (
	ExitPass       = 0
	ExitFail       = 1
	ExitInvalid    = 2
	ExitIncomplete = 3
)

// Format selects the output form.
type Format string

// The formats.
const (
	Text Format = "text"
	JSON Format = "json"
)

// ExitCode maps a report to the process exit code. A dry run plans work and
// passes.
func ExitCode(r run.Report) int {
	if r.Dry {
		return ExitPass
	}
	switch r.Results.Status {
	case model.Invalid:
		return ExitInvalid
	case model.Incomplete:
		return ExitIncomplete
	}
	if r.Verdict.Pass {
		return ExitPass
	}
	return ExitFail
}

// Document is the JSON form of a report.
type Document struct {
	Results   model.Results `json:"results"`
	Pass      *bool         `json:"pass,omitempty"`
	Verdict   []string      `json:"verdict,omitempty"`
	Planned   int           `json:"planned"`
	Pending   int           `json:"pending"`
	ModelRuns int           `json:"modelRuns"`
	Resumed   int           `json:"resumed"`
	Dry       bool          `json:"dry,omitempty"`
}

// Render writes the report in the format.
func Render(r run.Report, format Format) ([]byte, error) {
	if format == JSON {
		doc := Document{
			Results: r.Results, Verdict: r.Verdict.Lines, Planned: r.Planned, Pending: r.Pending,
			ModelRuns: r.ModelRuns, Resumed: r.Resumed, Dry: r.Dry,
		}
		if r.Results.Status == model.Complete && !r.Dry {
			pass := r.Verdict.Pass
			doc.Pass = &pass
		}
		return model.Encode(doc)
	}
	return []byte(renderText(r)), nil
}

func renderText(r run.Report) string {
	var b strings.Builder
	if r.Dry {
		fmt.Fprintf(&b, "dry run: %d units planned, %d pending, up to %d model runs\n", r.Planned, r.Pending, r.ModelRuns)
		fmt.Fprintf(&b, "model: %s, judge model: %s\n", r.Results.Model, r.Results.JudgeModel)
		return b.String()
	}
	res := r.Results
	fmt.Fprintf(&b, "%s %s (%s): %s\n", res.Protocol, res.RunID, res.Profile, status(r))
	fmt.Fprintf(&b, "model: %s, judge model: %s\n", res.Model, res.JudgeModel)
	for _, c := range res.Cases {
		fmt.Fprintf(&b, "  %-8s %s%s\n", c.Outcome, c.Key(), score(c))
	}
	for _, line := range r.Verdict.Lines {
		fmt.Fprintf(&b, "%s\n", line)
	}
	if res.Usage != nil {
		fmt.Fprintf(&b, "usage: %d input tokens, %d output tokens, $%.4f\n", res.Usage.InputTokens, res.Usage.OutputTokens, res.Usage.CostUSD)
	}
	switch res.Status {
	case model.Incomplete:
		fmt.Fprintf(&b, "stopped: %s limit reached with %d units left; run again with --resume\n", res.Stop, r.Pending)
	case model.Invalid:
		fmt.Fprintf(&b, "invalid: an infrastructure fault voids the run; it is no pass and no fail\n")
	}
	return b.String()
}

func status(r run.Report) string {
	switch r.Results.Status {
	case model.Complete:
		if r.Verdict.Pass {
			return "PASS"
		}
		return "FAIL"
	case model.Incomplete:
		return "INCOMPLETE"
	}
	return "INVALID"
}

func score(c model.CaseResult) string {
	if c.Score == nil {
		return ""
	}
	return fmt.Sprintf(" score %.2f", *c.Score)
}
