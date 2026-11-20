package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
)

func preflight(t *testing.T, r *testRoot, args ...string) (string, error) {
	t.Helper()
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	root.SetArgs(append([]string{"preflight"}, args...))
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	err := root.Execute()
	return r.out.String(), err
}

func TestPreflightReportsAReadyRunAndExitsZero(t *testing.T) {
	r := newTestRoot()

	out, err := preflight(t, r, "--isolation", "bwrap")

	if err != nil {
		t.Fatalf("preflight error = %v, want ready", err)
	}
	for _, want := range []string{"Preflight: ready", "isolation  bwrap (flags)", "Guarantees:", "ok    isolation bwrap"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestPreflightExitsOneForAGuaranteeThatDoesNotHold(t *testing.T) {
	r := newTestRoot()
	r.isolator.unpinned = true
	root := NewRootCommand(r.deps, r.ports, RecordPorts{}, r.bridge)
	root.SetArgs([]string{"preflight", "--isolation", "bwrap", "--require-pinned-provision"})
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)

	err := root.Execute()

	var exit launch.ExitError
	if !errors.As(err, &exit) || exit.Code != ExitNotReady {
		t.Fatalf("preflight error = %v, want exit code %d", err, ExitNotReady)
	}
	if !strings.Contains(r.out.String(), "Preflight: not ready") || !strings.Contains(r.out.String(), "required, NOT provided") {
		t.Errorf("report:\n%s", r.out.String())
	}
	if output.Len() != 0 {
		t.Errorf("cobra printed %q; the report is the whole output", output.String())
	}
}

func TestPreflightReportsAPlanProblem(t *testing.T) {
	r := newTestRoot()

	out, err := preflight(t, r, "--agent", "nosuch")

	var exit launch.ExitError
	if !errors.As(err, &exit) || exit.Code != ExitNotReady || !strings.Contains(out, "Problem:") {
		t.Errorf("error = %v, output = %q; want not ready with the problem", err, out)
	}
}

func TestPreflightPrintsJSON(t *testing.T) {
	r := newTestRoot()

	out, err := preflight(t, r, "--isolation", "bwrap", "--json")

	if err != nil {
		t.Fatalf("preflight error = %v", err)
	}
	var report launch.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil || !report.Ready || len(report.Axes) == 0 {
		t.Errorf("report = %+v, %v from %q", report, err, out)
	}
}

func TestPreflightTreatsAnUnreadableCommandLineAsUsage(t *testing.T) {
	for name, args := range map[string][]string{"unknown flag": {"--no-such-flag"}, "argument": {"extra"}} {
		r := newTestRoot()

		_, err := preflight(t, r, args...)

		var usage UsageError
		if !errors.As(err, &usage) {
			t.Errorf("%s: error = %v, want a usage error", name, err)
		}
	}
	if (UsageError{Err: io.EOF}).Error() != "EOF" || !errors.Is(UsageError{Err: io.EOF}, io.EOF) {
		t.Error("UsageError must carry its cause")
	}
}

func TestPreflightDoesNotRunTheAgent(t *testing.T) {
	r := newTestRoot()
	r.isolator.exitCode = 9

	if _, err := preflight(t, r, "--isolation", "bwrap"); err != nil {
		t.Fatalf("preflight error = %v", err)
	}
	if r.isolator.run.Agent != nil {
		t.Error("preflight built the sandbox command of a run")
	}
}

func TestPreflightSharesTheRunFlags(t *testing.T) {
	root := NewRootCommand(newTestRoot().deps, plugin.Manager{}, RecordPorts{}, &recordingBridge{})
	command, _, _ := root.Find([]string{"preflight"})
	for _, flag := range []string{"agent", "isolation", "network", "credential", "plugin", "terminal", "require-kernel-isolation", "json"} {
		if command.Flags().Lookup(flag) == nil {
			t.Errorf("preflight has no --%s flag", flag)
		}
	}
	if command.Flags().Lookup("dry-run") != nil {
		t.Error("preflight never runs, so it has no --dry-run flag")
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestPreflightReportsAFailureToWrite(t *testing.T) {
	report := launch.Report{Ready: true, Problem: "p", Axes: []launch.AxisChoice{{Axis: "a", Choice: "b", Source: "c"}},
		Guarantees:    []launch.GuaranteeStatus{{Guarantee: "g", Required: false, Provided: true}},
		Checks:        []launch.Check{{Name: "n", OK: true}},
		Configuration: []string{"k: v (user)"}}
	for _, asJSON := range []bool{true, false} {
		if err := writeReport(brokenWriter{}, report, asJSON); err == nil {
			t.Errorf("writeReport(json=%v) error = nil, want the write failure", asJSON)
		}
	}
}

func TestDescribeGuarantee(t *testing.T) {
	for want, status := range map[string]launch.GuaranteeStatus{
		"required, provided":              {Required: true, Provided: true},
		"provided":                        {Provided: true},
		"required, NOT provided: because": {Required: true, Reason: "because"},
		"not provided, not required":      {},
	} {
		if got := describeGuarantee(status); got != want {
			t.Errorf("describeGuarantee(%+v) = %q, want %q", status, got, want)
		}
	}
}
