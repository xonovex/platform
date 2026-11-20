package process_test

import (
	"context"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
)

// runner is a Runner that feeds fixed lines and returns a fixed outcome.
type runner struct {
	lines  []string
	result process.Result
	err    error
}

func (r runner) Run(_ context.Context, _ process.Request, out process.Output) (process.Result, error) {
	for _, line := range r.lines {
		if out([]byte(line)) {
			return process.Result{Stopped: true}, nil
		}
	}
	return r.result, r.err
}

func TestLinesRemovesTheLineEnding(t *testing.T) {
	var seen []string
	out := process.Lines(func(line string) bool { seen = append(seen, line); return line == "stop" })
	for _, line := range []string{"a\n", "b\r\n", "c", "\n"} {
		if out([]byte(line)) {
			t.Fatalf("%q must not stop", line)
		}
	}
	if !out([]byte("stop\n")) {
		t.Fatal("the handler's stop must pass through")
	}
	want := []string{"a", "b", "c", "", "stop"}
	if len(seen) != len(want) {
		t.Fatalf("seen = %q", seen)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("seen = %q, want %q", seen, want)
		}
	}
}

func TestCaptureKeepsTheExactOutput(t *testing.T) {
	got, err := process.Capture(context.Background(), runner{
		lines:  []string{"a\r\n", "\n", "tail"},
		result: process.Result{ExitCode: 0, Stderr: "warn"},
	}, process.Request{Name: "x"})
	if err != nil || string(got.Stdout) != "a\r\n\ntail" || got.Stderr != "warn" || !got.Succeeded() {
		t.Fatalf("Capture = %+v, %v", got, err)
	}
}

func TestCaptureReturnsTheFault(t *testing.T) {
	boom := errors.New("cannot start")
	if _, err := process.Capture(context.Background(), runner{err: boom}, process.Request{}); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}

func TestSucceeded(t *testing.T) {
	cases := map[string]struct {
		result process.Result
		want   bool
	}{
		"clean exit":   {process.Result{}, true},
		"failure exit": {process.Result{ExitCode: 1}, false},
		"timed out":    {process.Result{TimedOut: true}, false},
		"stopped":      {process.Result{Stopped: true}, false},
	}
	for name, c := range cases {
		if got := c.result.Succeeded(); got != c.want {
			t.Errorf("%s: Succeeded = %v", name, got)
		}
	}
}

func TestSplitterCutsChunksIntoExactLines(t *testing.T) {
	var lines []string
	splitter := process.NewSplitter(func(line []byte) bool { lines = append(lines, string(line)); return false }, nil)
	for _, chunk := range []string{"ab", "c\nd", "\r\n\n", "tail"} {
		if n, err := splitter.Write([]byte(chunk)); n != len(chunk) || err != nil {
			t.Fatalf("Write = %d, %v", n, err)
		}
	}
	if err := splitter.Close(); err != nil || splitter.Stopped() {
		t.Fatalf("Close = %v, stopped %v", err, splitter.Stopped())
	}
	want := []string{"abc\n", "d\r\n", "\n", "tail"}
	if len(lines) != len(want) {
		t.Fatalf("lines = %q", lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("lines = %q, want %q", lines, want)
		}
	}
}

func TestSplitterStopsOnceAndDiscardsTheRest(t *testing.T) {
	stops, seen := 0, 0
	splitter := process.NewSplitter(func([]byte) bool { seen++; return true }, func() { stops++ })
	_, _ = splitter.Write([]byte("a\nb\n"))
	_, _ = splitter.Write([]byte("c\nrest"))
	_ = splitter.Close()
	if !splitter.Stopped() || stops != 1 || seen != 1 {
		t.Fatalf("stopped %v, stops %d, seen %d", splitter.Stopped(), stops, seen)
	}
}

func TestSplitterStopsOnTheLastLine(t *testing.T) {
	splitter := process.NewSplitter(func(line []byte) bool { return string(line) == "end" }, nil)
	_, _ = splitter.Write([]byte("a\nend"))
	_ = splitter.Close()
	if !splitter.Stopped() {
		t.Fatal("the last line must reach the Output")
	}
}

func TestSplitterWithoutAnOutputDiscards(t *testing.T) {
	splitter := process.NewSplitter(nil, nil)
	if n, err := splitter.Write([]byte("a\nb")); n != 3 || err != nil {
		t.Fatalf("Write = %d, %v", n, err)
	}
	if err := splitter.Close(); err != nil || splitter.Stopped() {
		t.Fatalf("Close = %v", err)
	}
}
