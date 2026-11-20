// Package processtest holds the process double that the unit tests use in
// place of a real harness program.
package processtest

import (
	"context"
	"errors"
	"sync"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

// Reply is the scripted outcome of one program run.
type Reply struct {
	// Lines are the lines the program writes to standard output.
	Lines []string
	// Stderr is the standard error of the program.
	Stderr string
	// ExitCode is the exit status.
	ExitCode int
	// TimedOut makes the run end as if the timeout fired.
	TimedOut bool
	// Err makes the run fail as if the program could not be started.
	Err error
}

// Fake answers each run from a responder and records every request.
type Fake struct {
	mu      sync.Mutex
	respond func(process.Request) Reply
	calls   []process.Request
}

// New returns a Fake that answers every run with respond.
func New(respond func(process.Request) Reply) *Fake { return &Fake{respond: respond} }

// Sequence returns a Fake that answers the first run with the first reply, the
// second run with the second reply, and so on. A run beyond the script fails, so
// a test never passes by accident.
func Sequence(replies ...Reply) *Fake {
	next := 0
	var mu sync.Mutex
	return New(func(process.Request) Reply {
		mu.Lock()
		defer mu.Unlock()
		if next >= len(replies) {
			return Reply{Err: errors.New("processtest: no reply scripted")}
		}
		reply := replies[next]
		next++
		return reply
	})
}

// Calls returns the requests in the order they ran.
func (f *Fake) Calls() []process.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]process.Request(nil), f.calls...)
}

// Run records the request, then feeds the scripted lines to the handler until
// it asks to stop.
func (f *Fake) Run(_ context.Context, request process.Request, onLine process.LineHandler) (process.Result, error) {
	f.mu.Lock()
	f.calls = append(f.calls, request)
	f.mu.Unlock()
	reply := f.respond(request)
	if reply.Err != nil {
		return process.Result{}, reply.Err
	}
	for _, line := range reply.Lines {
		if onLine != nil && onLine(line) {
			return process.Result{Stopped: true}, nil
		}
	}
	return process.Result{ExitCode: reply.ExitCode, Stderr: reply.Stderr, TimedOut: reply.TimedOut}, nil
}
