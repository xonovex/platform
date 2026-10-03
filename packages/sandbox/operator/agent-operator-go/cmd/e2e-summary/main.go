package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

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

func run(args []string) error {
	if len(args) != 3 {
		return errors.New("usage: e2e-summary SUITE LOG OUTPUT")
	}

	log, err := os.ReadFile(args[1])
	summary := ""
	if errors.Is(err, os.ErrNotExist) {
		summary = renderMissingSummary(args[0])
	} else if err != nil {
		return fmt.Errorf("read test log: %w", err)
	} else {
		summary = renderSummary(args[0], string(log))
	}

	output, err := os.OpenFile(args[2], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open summary: %w", err)
	}
	_, writeErr := output.WriteString(summary)
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
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
