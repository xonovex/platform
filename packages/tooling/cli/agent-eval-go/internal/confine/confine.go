// Package confine is the port that places a harness program inside a sandbox.
// A confinement turns the process a harness leaf planned into the process that
// runs it inside the sandbox, and says how to end what that process leaves
// running. The executor runs both through the process port, so the core never
// names a sandbox and never touches the machine.
package confine

import "github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"

// Request is one harness process to confine.
type Request struct {
	// Harness names the harness whose program the process starts.
	Harness string
	// Process is the program the harness leaf planned. Its Dir is the private
	// workspace of the run.
	Process process.Request
}

// Confined is the process that runs a request inside its confinement.
type Confined struct {
	// Process runs the program of the request inside the confinement, with the
	// timeout of the request.
	Process process.Request
	// Stop, when set, ends what Process left running after it ended before the
	// program did: on a timeout, on a stop of its output or on a cancelled run.
	Stop *process.Request
	// Release frees what the confinement holds for the run. It is never nil.
	Release func()
}

// Confiner places harness processes inside one confinement.
type Confiner interface {
	Confine(request Request) (Confined, error)
}
