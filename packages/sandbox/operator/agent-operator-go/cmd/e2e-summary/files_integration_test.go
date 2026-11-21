//go:build integration

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests drive run through osFiles against a real directory, proving that
// the adapter reports a missing log as fs.ErrNotExist, appends to the summary
// and surfaces the errors of the real disk.

func TestOSFilesWritesSummaryForAnExistingLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	outputPath := filepath.Join(dir, "summary.md")
	if err := os.WriteFile(logPath, []byte("--- PASS: TestReady\n--- SKIP: TestNeedsKvm\n"), 0o644); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	if err := run([]string{"e2e-kata", logPath, outputPath}, osFiles{}); err != nil {
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

func TestOSFilesWritesMissingSummaryWhenLogIsAbsent(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "summary.md")

	if err := run([]string{"e2e", filepath.Join(dir, "absent.log"), outputPath}, osFiles{}); err != nil {
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

func TestOSFilesAppendsToAnExistingSummary(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	outputPath := filepath.Join(dir, "summary.md")
	if err := os.WriteFile(logPath, []byte("--- PASS: TestReady\n"), 0o644); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	if err := run([]string{"e2e", logPath, outputPath}, osFiles{}); err != nil {
		t.Fatalf("first run returned error %v", err)
	}
	if err := run([]string{"e2e-gvisor", logPath, outputPath}, osFiles{}); err != nil {
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

func TestOSFilesReportsAnUnreadableLog(t *testing.T) {
	dir := t.TempDir()

	// A directory is readable as a path but not as a file, so ReadFile fails
	// with something other than fs.ErrNotExist.
	err := run([]string{"e2e", dir, filepath.Join(dir, "summary.md")}, osFiles{})

	if err == nil {
		t.Fatal("run must return an error when the log cannot be read")
	}
	if !strings.Contains(err.Error(), "read test log") {
		t.Errorf("error %q must identify the read failure", err)
	}
}

func TestOSFilesReportsAnUnwritableSummary(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	if err := os.WriteFile(logPath, []byte("--- PASS: TestReady\n"), 0o644); err != nil {
		t.Fatalf("write test log: %v", err)
	}

	err := run([]string{"e2e", logPath, filepath.Join(dir, "absent-dir", "summary.md")}, osFiles{})

	if err == nil {
		t.Fatal("run must return an error when the summary cannot be opened")
	}
	if !strings.Contains(err.Error(), "open summary") {
		t.Errorf("error %q must identify the open failure", err)
	}
}
