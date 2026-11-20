// Package protocoltest holds the doubles that the protocol unit tests share: an
// executor that answers from a script and an environment built over an
// in-memory file system.
package protocoltest

import (
	"context"
	"sync"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/filesystem"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/session"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
)

// Exec is a harness.Executor that answers each request from Respond and records
// the requests.
type Exec struct {
	mu      sync.Mutex
	Respond func(harness.Request) (harness.Observation, error)
	calls   []harness.Request
}

// Execute records the request and answers it.
func (e *Exec) Execute(_ context.Context, req harness.Request) (harness.Observation, error) {
	e.mu.Lock()
	e.calls = append(e.calls, req)
	e.mu.Unlock()
	return e.Respond(req)
}

// Calls returns the recorded requests.
func (e *Exec) Calls() []harness.Request {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]harness.Request(nil), e.calls...)
}

// NewEnv builds an environment over the files. The catalog is loaded from the
// plugin roots, and the filesystem and session predicate leaves are registered,
// plus any extra evaluator the test passes.
func NewEnv(t *testing.T, files map[string]string, roots []string, exec harness.Executor, extra ...predicate.Evaluator) protocol.Env {
	t.Helper()
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	fsys := memfs.New(files)
	cat := catalog.Catalog{}
	if len(roots) > 0 {
		if cat, err = catalog.Load(fsys, roots); err != nil {
			t.Fatal(err)
		}
	}
	reg := predicate.NewRegistry()
	reg.Register(filesystem.New())
	reg.Register(session.New())
	for _, e := range extra {
		reg.Register(e)
	}
	return protocol.Env{FS: fsys, Reader: document.Reader{FS: fsys, Validator: validator}, Catalog: cat, Exec: exec, Predicates: reg}
}

// Result is the outcome of one unit.
type Result struct {
	Unit   protocol.Unit
	Result model.CaseResult
	Err    error
}

// RunAll runs every unit of a plan in order and returns the case results.
func RunAll(t *testing.T, plan protocol.Plan) []Result {
	t.Helper()
	var out []Result
	for _, u := range plan.Units {
		r, err := u.Run(context.Background())
		out = append(out, Result{Unit: u, Err: err})
		if err == nil {
			out[len(out)-1].Result = r
		}
	}
	return out
}
