// Package checktest builds the memory-backed fixtures that the unit tests of
// the checks share.
package checktest

import (
	"encoding/json"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// Repo is a repository that answers the git questions from lists.
type Repo struct {
	Tracked []string
	Ignored []string
	Err     error
}

// TrackedFiles returns the tracked list.
func (r Repo) TrackedFiles() ([]string, error) { return r.Tracked, r.Err }

// IgnoredPaths returns the ignored list.
func (r Repo) IgnoredPaths() ([]string, error) { return r.Ignored, r.Err }

// Context builds the context of a check over an in-memory repository.
func Context(t testing.TB, files map[string]string, paths ...string) check.Context {
	t.Helper()
	validator, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	return check.Context{
		Tree:       tree.Tree{FS: memfs.New(files), Root: "."},
		Paths:      paths,
		Validator:  validator,
		Repository: Repo{},
	}
}

// WithOptions sets the options of a context from a value.
func WithOptions(t testing.TB, ctx check.Context, options any) check.Context {
	t.Helper()
	data, err := json.Marshal(options)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Options = data
	return ctx
}

// Run runs a check and fails the test on a fault.
func Run(t testing.TB, c check.Check, ctx check.Context) []findings.Finding {
	t.Helper()
	got, err := c.Run(ctx)
	if err != nil {
		t.Fatalf("%s: %v", c.Name(), err)
	}
	return got
}

// Codes returns the code of every finding that is not a pass, in order.
func Codes(all []findings.Finding) []string {
	var out []string
	for _, f := range all {
		if f.Level != findings.Pass {
			out = append(out, f.Code)
		}
	}
	return out
}

// Has reports whether a finding carries the code.
func Has(all []findings.Finding, code string) bool {
	for _, f := range all {
		if f.Code == code {
			return true
		}
	}
	return false
}

// Find returns the first finding with the code.
func Find(all []findings.Finding, code string) (findings.Finding, bool) {
	for _, f := range all {
		if f.Code == code {
			return f, true
		}
	}
	return findings.Finding{}, false
}
