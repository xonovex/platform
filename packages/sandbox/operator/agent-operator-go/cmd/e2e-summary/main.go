package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// summaryFiles is the file access run needs: it reads the test log and appends
// to the summary, so a unit test drives run with an in-memory double and only
// osFiles reaches the disk.
type summaryFiles interface {
	ReadFile(name string) ([]byte, error)
	OpenAppend(name string) (io.WriteCloser, error)
}

func countLines(lines []string, marker string) int {
	count := 0
	for _, line := range lines {
		if strings.Contains(line, marker) {
			count++
		}
	}
	return count
}

func renderSummary(suite string, log string) string {
	lines := strings.Split(strings.ReplaceAll(log, "\r\n", "\n"), "\n")
	passed := countLines(lines, "--- PASS:")
	failed := countLines(lines, "--- FAIL:")
	skipped := make([]string, 0)
	for _, line := range lines {
		if strings.Contains(line, "--- SKIP:") {
			skipped = append(skipped, line)
		}
	}

	rows := []string{
		fmt.Sprintf("### %s", suite),
		"",
		"| Executed (pass) | Executed (fail) | Skipped |",
		"| --- | --- | --- |",
		fmt.Sprintf("| %d | %d | %d |", passed, failed, len(skipped)),
		"",
	}
	if len(skipped) > 0 {
		rows = append(
			rows,
			"Skipped tests (kata/coco isolation tests self-skip in unprivileged Kind):",
			"",
			"```",
		)
		rows = append(rows, skipped...)
		rows = append(rows, "```", "")
	}
	return strings.Join(rows, "\n") + "\n"
}

func renderMissingSummary(suite string) string {
	return fmt.Sprintf("### %s\n\nNo test log produced (suite failed before running tests).\n", suite)
}

func run(args []string, files summaryFiles) error {
	if len(args) != 3 {
		return errors.New("usage: e2e-summary SUITE LOG OUTPUT")
	}

	log, err := files.ReadFile(args[1])
	summary := ""
	if errors.Is(err, fs.ErrNotExist) {
		summary = renderMissingSummary(args[0])
	} else if err != nil {
		return fmt.Errorf("read test log: %w", err)
	} else {
		summary = renderSummary(args[0], string(log))
	}

	output, err := files.OpenAppend(args[2])
	if err != nil {
		return fmt.Errorf("open summary: %w", err)
	}
	_, writeErr := io.WriteString(output, summary)
	closeErr := output.Close()
	if writeErr != nil {
		return fmt.Errorf("write summary: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close summary: %w", closeErr)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], osFiles{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
