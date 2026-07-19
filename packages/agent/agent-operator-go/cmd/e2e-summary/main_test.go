package main

import (
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

func TestRenderMissingSummaryExplainsAbsentLog(t *testing.T) {
	summary := renderMissingSummary("e2e")

	if !strings.Contains(summary, "No test log produced") {
		t.Fatalf("expected missing-log explanation, got %q", summary)
	}
}
