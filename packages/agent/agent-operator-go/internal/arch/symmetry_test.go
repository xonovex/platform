// Package arch holds cross-module architecture fitness tests for the agent
// consumers (CLI + operator).
package arch

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// sharedAxes are the orthogonal axes both consumers realize. They must exist as a
// top-level internal/ directory in BOTH agent-cli-go and agent-operator-go — that
// name-level intersection is the symmetry the reorg guarantees.
var sharedAxes = []string{"isolation", "network", "provision", "workspace"}

// consumerOnlyAxes documents the deliberate asymmetries (axis present in one
// consumer only), so an unexpected divergence still fails while the agreed ones
// don't:
//   - harness, provider: operator-only. The CLI consumes harness commands and
//     provider env from the shared module (pkg/agents, pkg/providers) rather than
//     as CLI-internal axes.
//   - terminal: CLI-only. The operator has no terminal-output axis (pods, not TTYs).
var consumerOnlyAxes = map[string]string{
	"harness":  "operator-only (CLI uses shared pkg/agents)",
	"provider": "operator-only (CLI uses shared pkg/providers)",
	"terminal": "cli-only (operator runs pods, not terminals)",
}

func hasDir(base, name string) bool {
	info, err := os.Stat(filepath.Join(base, name))
	return err == nil && info.IsDir()
}

func TestAxisSymmetry(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = .../agent-operator-go/internal/arch/symmetry_test.go
	opInternal := filepath.Dir(filepath.Dir(file))
	opRoot := filepath.Dir(opInternal)
	cliInternal := filepath.Join(filepath.Dir(opRoot), "agent-cli-go", "internal")

	if _, err := os.Stat(cliInternal); err != nil {
		t.Skipf("CLI module not found at %s (isolated checkout); skipping cross-module symmetry", cliInternal)
	}

	for _, axis := range sharedAxes {
		if !hasDir(opInternal, axis) {
			t.Errorf("axis %q missing from operator internal/", axis)
		}
		if !hasDir(cliInternal, axis) {
			t.Errorf("axis %q missing from CLI internal/", axis)
		}
	}

	// A consumer-only axis must not silently appear in the other consumer without a
	// documented reason — that would be an undocumented asymmetry.
	if hasDir(cliInternal, "harness") {
		t.Errorf("harness is documented operator-only but exists in CLI internal/")
	}
	if hasDir(cliInternal, "provider") {
		t.Errorf("provider is documented operator-only but exists in CLI internal/")
	}
	if hasDir(opInternal, "terminal") {
		t.Errorf("terminal is documented cli-only but exists in operator internal/")
	}
	_ = consumerOnlyAxes // documentation map (asserted individually above)
}
