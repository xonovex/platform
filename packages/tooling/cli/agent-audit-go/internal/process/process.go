// Package process is the port for running a command. A core package that needs
// git or another program declares what it asks through Runner, so a unit test
// scripts the answers with processtest.Fake and only the osexec adapter starts
// a real process.
package process

import "context"

// Command is one program to run.
type Command struct {
	// Dir is the working directory of the program.
	Dir string
	// Name is the program to start.
	Name string
	// Args are the arguments of the program.
	Args []string
}

// Result is what a program that ran to its end left behind.
type Result struct {
	// Stdout is the standard output of the program.
	Stdout []byte
	// ExitCode is the exit status of the program.
	ExitCode int
}

// Succeeded reports whether the program exited with status 0.
func (r Result) Succeeded() bool { return r.ExitCode == 0 }

// Runner runs a command to its end. A program that ends with a non-zero status
// is a Result. The error is for a program that could not be started or that the
// context cancelled.
type Runner interface {
	Run(ctx context.Context, command Command) (Result, error)
}
