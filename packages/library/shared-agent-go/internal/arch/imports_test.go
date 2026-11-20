//go:build arch

package arch

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"
)

const (
	module   = "github.com/xonovex/platform/packages/library/shared-agent-go"
	contract = "github.com/xonovex/platform/packages/library/shared-contract-go"
)

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	"pkg/agentcmd":         {"pkg/providers", "pkg/types"},
	"pkg/agents":           {"pkg/types"},
	"pkg/artifact":         {},
	"pkg/egress":           {},
	"pkg/isolation":        {},
	"pkg/network":          {},
	"pkg/overlay":          {"pkg/overlay/claude", "pkg/overlay/codex", "pkg/overlay/copilot", "pkg/overlay/kiro", "pkg/overlay/opencode", "pkg/overlay/pi", "pkg/overlay/shared", "pkg/types"},
	"pkg/overlay/claude":   {"pkg/artifact", "pkg/overlay/shared"},
	"pkg/overlay/codex":    {"pkg/artifact", "pkg/overlay/shared"},
	"pkg/overlay/copilot":  {"pkg/artifact", "pkg/overlay/shared"},
	"pkg/overlay/kiro":     {"pkg/artifact", "pkg/overlay/shared"},
	"pkg/overlay/opencode": {"pkg/artifact", "pkg/overlay/shared"},
	"pkg/overlay/pi":       {"pkg/artifact", "pkg/overlay/shared"},
	"pkg/overlay/shared":   {"pkg/artifact"},
	"pkg/plugin":           {"pkg/agents", "pkg/artifact", "pkg/types"},
	"pkg/policy":           {},
	"pkg/providers":        {"pkg/types"},
	"pkg/provision":        {},
	"pkg/provision/nix":    {},
	"pkg/types":            {},
	"pkg/validation":       {},
	"pkg/workspace":        {},
}

// contracts lists, for each package, the packages of shared-contract-go it may
// import from its non-test files. A package missing here imports none.
var contracts = map[string][]string{
	"pkg/agents": {"pkg/jsonschema", "pkg/schemas"},
}

// leaves are the overlay of each harness. The overlay registry is the one
// package that imports them, so no other package names a harness layout.
var leaves = map[string]string{
	"pkg/overlay/claude":   "pkg/overlay",
	"pkg/overlay/codex":    "pkg/overlay",
	"pkg/overlay/copilot":  "pkg/overlay",
	"pkg/overlay/kiro":     "pkg/overlay",
	"pkg/overlay/opencode": "pkg/overlay",
	"pkg/overlay/pi":       "pkg/overlay",
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(here)))
}

func moduleFiles(t *testing.T) []archtest.File {
	t.Helper()
	files, err := archtest.Load(os.DirFS(moduleRoot(t)), "pkg")
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// TestEveryPackageIsDeclared fails for a package missing from the table and for
// a table entry with no package.
func TestEveryPackageIsDeclared(t *testing.T) {
	present := map[string]bool{}
	for _, f := range moduleFiles(t) {
		present[f.Dir] = true
	}
	for dir := range present {
		if _, ok := allowed[dir]; !ok {
			t.Errorf("package %s is not declared in the allowed table", dir)
		}
	}
	for dir := range allowed {
		if !present[dir] {
			t.Errorf("the allowed table names %s, which has no Go files", dir)
		}
	}
}

// TestImportsFollowTheTables keeps the dependency direction inside the module
// and toward the contract.
func TestImportsFollowTheTables(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			if rel, inside := strings.CutPrefix(imp, module+"/"); inside && !slices.Contains(allowed[f.Dir], rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.Path, rel, f.Dir, allowed[f.Dir])
			}
			if rel, inside := strings.CutPrefix(imp, contract+"/"); inside && !slices.Contains(contracts[f.Dir], rel) {
				t.Errorf("%s imports the contract package %s; %s may import only %v", f.Path, rel, f.Dir, contracts[f.Dir])
			}
		}
	}
}

// TestALeafHasOneImporter keeps a harness overlay behind the registry: no
// sibling and no other package imports it.
func TestALeafHasOneImporter(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if owner, leaf := leaves[rel]; inside && leaf && f.Dir != owner {
				t.Errorf("%s imports the leaf %s; only %s may", f.Path, rel, owner)
			}
		}
	}
}

// TestTheLibraryUsesOnlyTheStandardLibraryAndTheContract keeps outside modules
// out of the agent types.
func TestTheLibraryUsesOnlyTheStandardLibraryAndTheContract(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			if !archtest.IsStandard(imp) && !strings.HasPrefix(imp, module+"/") && !strings.HasPrefix(imp, contract+"/") {
				t.Errorf("%s imports %s; the library uses only the standard library and the contract", f.Path, imp)
			}
		}
	}
}

// TestNoPackageTouchesTheHost keeps the disk, the environment, the process and
// the network out of the library and out of every unit test: the library
// describes an agent and a policy, and a consumer's adapter acts on them.
func TestNoPackageTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if archtest.IsHost(imp) {
				t.Errorf("%s imports %s; the library reaches no host", f.Path, imp)
			}
		}
	}
}

// TestNoPackageReadsTheClock keeps time an input.
func TestNoPackageReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if archtest.ReadsClock(f.Syntax) {
			t.Errorf("%s reads the clock with time.Now, time.Since or time.Until; pass the instant in", f.Path)
		}
	}
}

// TestNoPackagePrints keeps the standard streams out of the library.
func TestNoPackagePrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; return the text to the caller", f.Path, use)
		}
	}
}

// TestNoPackageKeepsState keeps every package free of state that one caller
// could change under another.
func TestNoPackageKeepsState(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, state := range archtest.PackageState(f.Syntax) {
			t.Errorf("%s declares %s; pass state as a value", f.Path, state)
		}
	}
}
