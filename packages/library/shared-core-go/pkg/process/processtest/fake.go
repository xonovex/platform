// Package processtest holds the process double that unit tests use in place of
// a real program.
package processtest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
)

// Reply is the scripted outcome of one program run.
type Reply struct {
	// Stdout is the standard output of the program, handed to the Output one
	// line at a time.
	Stdout string
	// Stderr is the standard error of the program.
	Stderr string
	// ExitCode is the exit status.
	ExitCode int
	// TimedOut makes the run end as if the timeout fired.
	TimedOut bool
	// Err makes the run fail as if the program could not be started.
	Err error
}

// Lines renders lines as a standard output in which every line ends with a
// newline.
func Lines(lines ...string) string {
	var out strings.Builder
	for _, line := range lines {
		out.WriteString(line)
		out.WriteByte('\n')
	}
	return out.String()
}

// Key renders a command line the way On expects it.
func Key(name string, args ...string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}

// Fake answers each run from its script of command lines, then from its
// responder, and records every request. A run that neither answers fails, so a
// test never passes by accident.
type Fake struct {
	mu      sync.Mutex
	script  map[string]Reply
	respond func(process.Request) Reply
	calls   []process.Request
}

// Script returns a Fake that answers only the command lines given to On.
func Script() *Fake { return &Fake{script: map[string]Reply{}} }

// New returns a Fake that answers every run with respond.
func New(respond func(process.Request) Reply) *Fake {
	return &Fake{script: map[string]Reply{}, respond: respond}
}

// Sequence returns a Fake that answers the first run with the first reply, the
// second run with the second reply, and so on. A run beyond the script fails.
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

// On scripts the reply to a command line such as "git tag --list". It returns
// the Fake so calls chain.
func (f *Fake) On(commandLine string, reply Reply) *Fake {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.script[commandLine] = reply
	return f
}

// Calls returns the requests in the order they ran.
func (f *Fake) Calls() []process.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]process.Request(nil), f.calls...)
}

// Run records the request, then feeds the scripted standard output to out until
// it asks to stop.
func (f *Fake) Run(_ context.Context, request process.Request, out process.Output) (process.Result, error) {
	reply, err := f.reply(request)
	if err != nil {
		return process.Result{}, err
	}
	if reply.Err != nil {
		return process.Result{}, reply.Err
	}
	lines := process.NewSplitter(out, nil)
	_, _ = lines.Write([]byte(reply.Stdout))
	_ = lines.Close()
	if lines.Stopped() {
		return process.Result{Stderr: reply.Stderr, Stopped: true}, nil
	}
	return process.Result{ExitCode: reply.ExitCode, Stderr: reply.Stderr, TimedOut: reply.TimedOut}, nil
}

func (f *Fake) reply(request process.Request) (Reply, error) {
	f.mu.Lock()
	f.calls = append(f.calls, request)
	scripted, ok := f.script[Key(request.Name, request.Args...)]
	respond := f.respond
	f.mu.Unlock()
	switch {
	case ok:
		return scripted, nil
	case respond != nil:
		return respond(request), nil
	default:
		return Reply{}, fmt.Errorf("processtest: no reply scripted for %q", Key(request.Name, request.Args...))
	}
}
