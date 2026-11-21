package main

import (
	"errors"
	"io"
	"io/fs"
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
		err := run(args, newMemoryFiles(nil))
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
	files := newMemoryFiles(map[string]string{"test.log": "--- PASS: TestReady\n--- SKIP: TestNeedsKvm\n"})

	if err := run([]string{"e2e-kata", "test.log", "summary.md"}, files); err != nil {
		t.Fatalf("run returned error %v", err)
	}

	summary := files.contents["summary.md"]
	if !strings.Contains(summary, "| 1 | 0 | 1 |") {
		t.Errorf("summary %q must report the log's counts", summary)
	}
	if !strings.Contains(summary, "### e2e-kata") {
		t.Errorf("summary %q must name the suite", summary)
	}
}

// A suite that dies before running tests leaves no log, which is reported rather
// than treated as a failure to read.
func TestRunWritesMissingSummaryWhenLogIsAbsent(t *testing.T) {
	files := newMemoryFiles(nil)

	if err := run([]string{"e2e", "absent.log", "summary.md"}, files); err != nil {
		t.Fatalf("run returned error %v", err)
	}

	summary := files.contents["summary.md"]
	if !strings.Contains(summary, "No test log produced") {
		t.Errorf("summary %q must explain the absent log", summary)
	}
}

func TestRunAppendsToAnExistingSummary(t *testing.T) {
	files := newMemoryFiles(map[string]string{"test.log": "--- PASS: TestReady\n"})

	if err := run([]string{"e2e", "test.log", "summary.md"}, files); err != nil {
		t.Fatalf("first run returned error %v", err)
	}
	if err := run([]string{"e2e-gvisor", "test.log", "summary.md"}, files); err != nil {
		t.Fatalf("second run returned error %v", err)
	}

	summary := files.contents["summary.md"]
	if !strings.Contains(summary, "### e2e\n") {
		t.Errorf("summary %q must retain the first suite", summary)
	}
	if !strings.Contains(summary, "### e2e-gvisor\n") {
		t.Errorf("summary %q must contain the second suite", summary)
	}
}

func TestRunReportsAnUnreadableLog(t *testing.T) {
	files := newMemoryFiles(nil)
	files.readErr = errors.New("is a directory")

	err := run([]string{"e2e", "test.log", "summary.md"}, files)

	if err == nil {
		t.Fatal("run must return an error when the log cannot be read")
	}
	if !strings.Contains(err.Error(), "read test log") {
		t.Errorf("error %q must identify the read failure", err)
	}
}

func TestRunReportsAnUnwritableSummary(t *testing.T) {
	files := newMemoryFiles(map[string]string{"test.log": "--- PASS: TestReady\n"})
	files.openErr = fs.ErrNotExist

	err := run([]string{"e2e", "test.log", "summary.md"}, files)

	if err == nil {
		t.Fatal("run must return an error when the summary cannot be opened")
	}
	if !strings.Contains(err.Error(), "open summary") {
		t.Errorf("error %q must identify the open failure", err)
	}
}

func TestRunReportsAFailedSummaryWrite(t *testing.T) {
	files := newMemoryFiles(map[string]string{"test.log": "--- PASS: TestReady\n"})
	files.writeErr = errors.New("no space left on device")

	err := run([]string{"e2e", "test.log", "summary.md"}, files)

	if err == nil {
		t.Fatal("run must return an error when the summary cannot be written")
	}
	if !strings.Contains(err.Error(), "write summary") {
		t.Errorf("error %q must identify the write failure", err)
	}
	if !files.closed {
		t.Error("run must close the summary after a failed write")
	}
}

func TestRunReportsAFailedSummaryClose(t *testing.T) {
	files := newMemoryFiles(map[string]string{"test.log": "--- PASS: TestReady\n"})
	files.closeErr = errors.New("input/output error")

	err := run([]string{"e2e", "test.log", "summary.md"}, files)

	if err == nil {
		t.Fatal("run must return an error when the summary cannot be closed")
	}
	if !strings.Contains(err.Error(), "close summary") {
		t.Errorf("error %q must identify the close failure", err)
	}
}

// memoryFiles is the in-memory summaryFiles double: contents holds every file
// by name, and each error field makes the matching operation fail.
type memoryFiles struct {
	contents           map[string]string
	readErr, openErr   error
	writeErr, closeErr error
	closed             bool
}

func newMemoryFiles(contents map[string]string) *memoryFiles {
	files := &memoryFiles{contents: map[string]string{}}
	for name, data := range contents {
		files.contents[name] = data
	}
	return files
}

func (m *memoryFiles) ReadFile(name string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	data, ok := m.contents[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return []byte(data), nil
}

func (m *memoryFiles) OpenAppend(name string) (io.WriteCloser, error) {
	if m.openErr != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: m.openErr}
	}
	return &memoryAppender{files: m, name: name}, nil
}

type memoryAppender struct {
	files *memoryFiles
	name  string
}

func (a *memoryAppender) Write(data []byte) (int, error) {
	if a.files.writeErr != nil {
		return 0, a.files.writeErr
	}
	a.files.contents[a.name] += string(data)
	return len(data), nil
}

func (a *memoryAppender) Close() error {
	a.files.closed = true
	return a.files.closeErr
}
