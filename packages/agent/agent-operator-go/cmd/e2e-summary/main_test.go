package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderSummaryReportsExecutedAndSkippedTests(t *testing.T) {
	log := strings.Join([]string{
		"--- PASS: TestReady",
		"--- FAIL: TestBroken",
		"--- SKIP: TestNeedsKvm (0.00s)",
	}, "\n")

	summary := renderSummary("e2e-kata", log)

	if !strings.Contains(summary, "| 1 | 1 | 1 |") {
		t.Fatalf("expected result counts in summary, got %q", summary)
	}
	if !strings.Contains(summary, "--- SKIP: TestNeedsKvm") {
		t.Fatalf("expected skipped test in summary, got %q", summary)
	}
}

func TestRenderSummaryOmitsSkipSectionWhenNothingSkipped(t *testing.T) {
	summary := renderSummary("e2e", "--- PASS: TestReady\n--- PASS: TestAlsoReady")

	if !strings.Contains(summary, "| 2 | 0 | 0 |") {
		t.Fatalf("expected two passes and no skips, got %q", summary)
	}
	if strings.Contains(summary, "Skipped tests") {
		t.Errorf("summary must omit the skip section when nothing skipped, got %q", summary)
	}
}

func TestRenderSummaryCountsCarriageReturnDelimitedLogs(t *testing.T) {
	summary := renderSummary("e2e", "--- PASS: TestReady\r\n--- FAIL: TestBroken\r\n")

	if !strings.Contains(summary, "| 1 | 1 | 0 |") {
		t.Fatalf("expected CRLF lines to be counted, got %q", summary)
	}
}

func TestRenderMissingSummaryExplainsAbsentLog(t *testing.T) {
	summary := renderMissingSummary("e2e")

	if !strings.Contains(summary, "No test log produced") {
		t.Fatalf("expected missing-log explanation, got %q", summary)
	}
}

func TestRunRejectsWrongArgumentCount(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"e2e"},
		{"e2e", "log"},
		{"e2e", "log", "out", "extra"},
	} {
		err := run(args)
		if err == nil {
			t.Errorf("run(%q) must return an error", args)
			continue
		}
		if !strings.Contains(err.Error(), "usage:") {
			t.Errorf("run(%q) error %q must state the usage", args, err)
		}
	}
}

func TestRunWritesSummaryForAnExistingLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	outputPath := filepath.Join(dir, "summary.md")
	if err := os.WriteFile(logPath, []byte("--- PASS: TestReady\n--- SKIP: TestNeedsKvm\n"), 0o644); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	if err := run([]string{"e2e-kata", logPath, outputPath}); err != nil {
		t.Fatalf("run returned error %v", err)
	}

	summary, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if !strings.Contains(string(summary), "| 1 | 0 | 1 |") {
		t.Errorf("summary %q must report the log's counts", summary)
	}
	if !strings.Contains(string(summary), "### e2e-kata") {
		t.Errorf("summary %q must name the suite", summary)
	}
}

// A suite that dies before running tests leaves no log, which is reported rather
// than treated as a failure to read.
func TestRunWritesMissingSummaryWhenLogIsAbsent(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "summary.md")

	if err := run([]string{"e2e", filepath.Join(dir, "absent.log"), outputPath}); err != nil {
		t.Fatalf("run returned error %v", err)
	}

	summary, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if !strings.Contains(string(summary), "No test log produced") {
		t.Errorf("summary %q must explain the absent log", summary)
	}
}

func TestRunAppendsToAnExistingSummary(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	outputPath := filepath.Join(dir, "summary.md")
	if err := os.WriteFile(logPath, []byte("--- PASS: TestReady\n"), 0o644); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	if err := run([]string{"e2e", logPath, outputPath}); err != nil {
		t.Fatalf("first run returned error %v", err)
	}
	if err := run([]string{"e2e-gvisor", logPath, outputPath}); err != nil {
		t.Fatalf("second run returned error %v", err)
	}

	summary, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if !strings.Contains(string(summary), "### e2e\n") {
		t.Errorf("summary %q must retain the first suite", summary)
	}
	if !strings.Contains(string(summary), "### e2e-gvisor\n") {
		t.Errorf("summary %q must contain the second suite", summary)
	}
}

func TestRunReportsAnUnreadableLog(t *testing.T) {
	dir := t.TempDir()

	// A directory is readable as a path but not as a file, so ReadFile fails
	// with something other than os.ErrNotExist.
	err := run([]string{"e2e", dir, filepath.Join(dir, "summary.md")})

	if err == nil {
		t.Fatal("run must return an error when the log cannot be read")
	}
	if !strings.Contains(err.Error(), "read test log") {
		t.Errorf("error %q must identify the read failure", err)
	}
}

func TestRunReportsAnUnwritableSummary(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	if err := os.WriteFile(logPath, []byte("--- PASS: TestReady\n"), 0o644); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	err := run([]string{"e2e", logPath, filepath.Join(dir, "absent-dir", "summary.md")})

	if err == nil {
		t.Fatal("run must return an error when the summary cannot be opened")
	}
	if !strings.Contains(err.Error(), "open summary") {
		t.Errorf("error %q must identify the open failure", err)
	}
}
