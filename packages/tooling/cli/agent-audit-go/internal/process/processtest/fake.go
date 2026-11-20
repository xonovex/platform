// Package processtest holds the process test double that the unit tests of this
// module use in place of a real program.
package processtest

import (
	"context"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process"
)

// Reply is the scripted outcome of one command.
type Reply struct {
	Stdout   string
	ExitCode int
	// Err makes Run fail as if the program could not be started.
	Err error
}

// Fake answers commands from a script and records every command it was given.
type Fake struct {
	script map[string]Reply
	// Calls lists the commands in the order they ran.
	Calls []process.Command
}

// New returns a Fake with an empty script.
func New() *Fake { return &Fake{script: map[string]Reply{}} }

// Key renders a command line the way On expects it.
func Key(name string, args ...string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}

// On scripts the reply to a command line such as "git tag --list". It returns
// the Fake so calls chain.
func (f *Fake) On(commandLine string, reply Reply) *Fake {
	f.script[commandLine] = reply
	return f
}

// Run records the command and returns its scripted reply. An unscripted command
// is an error, so a test never passes by accident.
func (f *Fake) Run(_ context.Context, command process.Command) (process.Result, error) {
	f.Calls = append(f.Calls, command)
	reply, ok := f.script[Key(command.Name, command.Args...)]
	if !ok {
		return process.Result{}, fmt.Errorf("processtest: no reply scripted for %q", Key(command.Name, command.Args...))
	}
	if reply.Err != nil {
		return process.Result{}, reply.Err
	}
	return process.Result{Stdout: []byte(reply.Stdout), ExitCode: reply.ExitCode}, nil
}
