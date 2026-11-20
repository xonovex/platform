package report_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/report"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/run"
)

func rep(status model.Status, pass bool) run.Report {
	s := 0.5
	return run.Report{
		Results: model.Results{
			Protocol: model.Trigger, RunID: "trigger-x-1", Profile: "p", Status: status, Stop: model.StopBudget,
			Cases: []model.CaseResult{{ID: "q", Outcome: model.Passed, Score: &s}, {ID: "c", Condition: model.Absent, Outcome: model.Failed}},
			Usage: &model.Usage{InputTokens: 1, OutputTokens: 2, CostUSD: 0.5},
		},
		Verdict: protocol.Verdict{Pass: pass, Lines: []string{"line one"}}, Planned: 4, Pending: 2, ModelRuns: 6,
	}
}

func TestExitCodes(t *testing.T) {
	for name, tc := range map[string]struct {
		r    run.Report
		want int
	}{
		"pass":       {rep(model.Complete, true), report.ExitPass},
		"fail":       {rep(model.Complete, false), report.ExitFail},
		"invalid":    {rep(model.Invalid, false), report.ExitInvalid},
		"incomplete": {rep(model.Incomplete, false), report.ExitIncomplete},
		"dry":        {run.Report{Dry: true}, report.ExitPass},
	} {
		if got := report.ExitCode(tc.r); got != tc.want {
			t.Errorf("%s: %d", name, got)
		}
	}
}

func text(t *testing.T, r run.Report) string {
	out, err := report.Render(r, report.Text)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestTextForEachStatus(t *testing.T) {
	pass := text(t, rep(model.Complete, true))
	for _, want := range []string{"trigger trigger-x-1 (p): PASS", "passed   q score 0.50", "failed   c@absent", "line one", "usage: 1 input tokens, 2 output tokens, $0.5000"} {
		if !strings.Contains(pass, want) {
			t.Errorf("missing %q in\n%s", want, pass)
		}
	}
	if !strings.Contains(text(t, rep(model.Complete, false)), "FAIL") {
		t.Fatal("fail")
	}
	if !strings.Contains(text(t, rep(model.Incomplete, false)), "budget limit reached with 2 units left; run again with --resume") {
		t.Fatal("incomplete")
	}
	if !strings.Contains(text(t, rep(model.Invalid, false)), "no pass and no fail") {
		t.Fatal("invalid")
	}
	if dry := text(t, run.Report{Dry: true, Planned: 3, Pending: 3, ModelRuns: 9}); !strings.Contains(dry, "dry run: 3 units planned, 3 pending, up to 9 model runs") {
		t.Fatal(dry)
	}
}

func TestJSON(t *testing.T) {
	out, err := report.Render(rep(model.Complete, true), report.JSON)
	if err != nil || !strings.Contains(string(out), `"pass": true`) || !strings.Contains(string(out), `"modelRuns": 6`) {
		t.Fatalf("%v %s", err, out)
	}
	partial, _ := report.Render(rep(model.Incomplete, false), report.JSON)
	if strings.Contains(string(partial), `"pass"`) {
		t.Fatal("an incomplete run has no pass member")
	}
	dry, _ := report.Render(run.Report{Dry: true}, report.JSON)
	if strings.Contains(string(dry), `"pass"`) || !strings.Contains(string(dry), `"dry": true`) {
		t.Fatalf("%s", dry)
	}
}
