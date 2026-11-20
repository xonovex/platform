// Package process is the port for running a program. The adapter hands the
// standard output to an Output one line at a time, so a caller that knows its
// answer part way through a stream can end the program early, and Capture
// collects the whole output for a caller that needs all of it.
package process

import (
	"bytes"
	"context"
	"time"
)

// Request is one program to run.
type Request struct {
	// Dir is the working directory of the program.
	Dir string
	// Name is the program to start.
	Name string
	// Args are the arguments of the program.
	Args []string
	// Env holds variables that replace or extend the inherited environment.
	Env map[string]string
	// Timeout ends the program when it runs longer. Zero means no limit.
	Timeout time.Duration
}

// Output receives the standard output of a program one line at a time and
// returns true to stop the program there. A line keeps its newline, and only
// the last line can lack one, so the lines joined in order are the exact
// output. An Output may keep the slice it receives.
type Output func(line []byte) (stop bool)

// LineHandler receives one line of standard output without its line ending and
// returns true to stop the program there.
type LineHandler func(line string) (stop bool)

// Result is what a program that ran to its end left behind.
type Result struct {
	// ExitCode is the exit status. It is meaningless when Stopped or TimedOut.
	ExitCode int
	// Stderr is the standard error of the program.
	Stderr string
	// TimedOut reports that the timeout ended the program.
	TimedOut bool
	// Stopped reports that the Output ended the program.
	Stopped bool
}

// Succeeded reports whether the program ran to its end by itself and exited
// with status 0.
func (r Result) Succeeded() bool { return r.ExitCode == 0 && !r.TimedOut && !r.Stopped }

// Runner runs a program and feeds its standard output to an Output, which may
// be nil to discard it. A program that ends with a non-zero status is a Result.
// The error is for a program that could not be started or that the context
// cancelled.
type Runner interface {
	Run(ctx context.Context, request Request, out Output) (Result, error)
}

// Lines adapts a LineHandler to an Output. It removes the newline and a
// carriage return before it, so a handler sees the same text on every platform.
func Lines(handler LineHandler) Output {
	return func(line []byte) bool {
		text := bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))
		return handler(string(text))
	}
}

// Splitter cuts a stream of standard output into the lines of an Output. After
// the Output asks to stop, it calls onStop once and discards the rest, so a
// program that writes on is never blocked on a full pipe.
type Splitter struct {
	out     Output
	onStop  func()
	pending []byte
	stopped bool
}

// NewSplitter returns a Splitter that feeds out, which may be nil to discard
// the stream, and calls onStop, which may be nil, when out asks to stop.
func NewSplitter(out Output, onStop func()) *Splitter {
	return &Splitter{out: out, onStop: onStop}
}

// Write hands every complete line in p to the Output. It never fails.
func (s *Splitter) Write(p []byte) (int, error) {
	if s.stopped || s.out == nil {
		return len(p), nil
	}
	s.pending = append(s.pending, p...)
	for !s.stopped {
		end := bytes.IndexByte(s.pending, '\n')
		if end < 0 {
			break
		}
		line := append([]byte(nil), s.pending[:end+1]...)
		s.pending = s.pending[end+1:]
		s.hand(line)
	}
	return len(p), nil
}

// Close hands the last line, which has no newline, to the Output.
func (s *Splitter) Close() error {
	if !s.stopped && s.out != nil && len(s.pending) > 0 {
		s.hand(s.pending)
	}
	s.pending = nil
	return nil
}

// Stopped reports whether the Output asked to stop.
func (s *Splitter) Stopped() bool { return s.stopped }

func (s *Splitter) hand(line []byte) {
	if !s.out(line) {
		return
	}
	s.stopped = true
	s.pending = nil
	if s.onStop != nil {
		s.onStop()
	}
}

// Captured is a Result with the whole standard output of the program.
type Captured struct {
	Result
	// Stdout is the exact standard output of the program.
	Stdout []byte
}

// Capture runs the request to its end and keeps its whole standard output.
func Capture(ctx context.Context, runner Runner, request Request) (Captured, error) {
	var stdout []byte
	result, err := runner.Run(ctx, request, func(line []byte) bool {
		stdout = append(stdout, line...)
		return false
	})
	if err != nil {
		return Captured{}, err
	}
	return Captured{Result: result, Stdout: stdout}, nil
}
