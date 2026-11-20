// Package none is the confinement that confines nothing: the harness program
// runs on the host exactly as its leaf planned it.
package none

import "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/confine"

// Confiner passes every request through unchanged.
type Confiner struct{}

var _ confine.Confiner = Confiner{}

// New returns the confinement that confines nothing.
func New() Confiner { return Confiner{} }

// Confine returns the process of the request. Nothing outlives it, so there is
// nothing to stop and nothing to release.
func (Confiner) Confine(request confine.Request) (confine.Confined, error) {
	return confine.Confined{Process: request.Process, Release: func() {}}, nil
}
