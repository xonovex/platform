// Package process is the port for running a harness program. A harness streams
// its answer as lines on standard output, and an evaluator often knows the
// verdict part way through, so the port hands each line to a handler that can
// end the run early instead of paying for the rest of the stream.
package process

import (
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

// LineHandler receives each complete line of standard output and returns true
// to stop the program there.
type LineHandler func(line string) (stop bool)

// Result is what a program that ran to its end left behind.
type Result struct {
	// ExitCode is the exit status. It is meaningless when Stopped or TimedOut.
	ExitCode int
	// Stderr is the standard error of the program.
	Stderr string
	// TimedOut reports that the timeout ended the program.
	TimedOut bool
	// Stopped reports that the line handler ended the program.
	Stopped bool
}

// Runner runs a program and feeds its standard output to a handler. A program
// that ends with a non-zero status is a Result. The error is for a program that
// could not be started or that the context cancelled.
type Runner interface {
	Run(ctx context.Context, request Request, onLine LineHandler) (Result, error)
}
