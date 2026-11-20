//go:build arch

// Package arch holds the architecture fitness tests of the library. They read
// the source tree, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run.
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
	module   = "github.com/xonovex/platform/packages/library/shared-config-go"
	contract = "github.com/xonovex/platform/packages/library/shared-contract-go"
)

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	"pkg/spec":              {},
	"pkg/section/run":       {"pkg/spec"},
	"pkg/section/providers": {"pkg/spec"},
	"pkg/section/validate":  {"pkg/spec"},
	"pkg/section/audit":     {"pkg/spec"},
	"pkg/section/eval":      {"pkg/spec"},
	"pkg/layer":             {"pkg/spec"},
	"pkg/discover":          {"pkg/spec"},
	"pkg/merge":             {"pkg/layer", "pkg/spec"},
	"pkg/trust":             {"pkg/layer"},
	"pkg/explain":           {"pkg/merge", "pkg/spec"},
	"pkg/resolve":           {"pkg/discover", "pkg/layer", "pkg/merge", "pkg/spec", "pkg/trust"},
	"pkg/shard":             {"pkg/discover", "pkg/layer", "pkg/spec", "pkg/trust"},
	"pkg/osfiles":           {"pkg/discover"},
	"pkg/configtest":        {"pkg/discover"},
}

// contracts lists, for each package, the packages of shared-contract-go it may
// import from its non-test files. A package missing here imports none.
var contracts = map[string][]string{
	"pkg/section/run":       {"pkg/schemas"},
	"pkg/section/providers": {"pkg/schemas"},
	"pkg/section/validate":  {"pkg/schemas"},
	"pkg/section/audit":     {"pkg/schemas"},
	"pkg/section/eval":      {"pkg/schemas"},
	"pkg/layer":             {"pkg/jsonschema"},
	"pkg/resolve":           {"pkg/jsonschema"},
}

// leaves bind a port to the operating system or to a test double. No package of
// this module imports one from non-test code: a consumer's composition root
// chooses the leaf.
var leaves = []string{"pkg/osfiles", "pkg/configtest"}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(here)))
}

// moduleFiles parses the module's source, without the fitness tests themselves.
func moduleFiles(t *testing.T) []archtest.File {
	t.Helper()
	files, err := archtest.Load(os.DirFS(moduleRoot(t)), "pkg")
	if err != nil {
		t.Fatal(err)
	}
	var out []archtest.File
	for _, f := range files {
		if f.Dir != "internal/arch" {
			out = append(out, f)
		}
	}
	return out
}

// hostLeaves may import hostImports from non-test code, and their integration
// tests may too.
var hostLeaves = []string{"pkg/osfiles"}

// TestEveryPackageIsDeclared fails for a package missing from the table and for
// a table entry with no package.
func TestEveryPackageIsDeclared(t *testing.T) {
	present := map[string]bool{}
	for _, f := range archtest.Production(moduleFiles(t)) {
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

// TestNoPackageImportsALeaf leaves the choice of a leaf to the consumer.
func TestNoPackageImportsALeaf(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && slices.Contains(leaves, rel) {
				t.Errorf("%s imports the leaf %s", f.Path, rel)
			}
		}
	}
}

// TestTheLibraryUsesOnlyTheStandardLibraryAndTheContract keeps outside modules
// out of the configuration core.
func TestTheLibraryUsesOnlyTheStandardLibraryAndTheContract(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if !archtest.IsStandard(imp) && !strings.HasPrefix(imp, module) && !strings.HasPrefix(imp, contract+"/") {
				t.Errorf("%s imports %s; the library uses only the standard library and the contract", f.Path, imp)
			}
		}
	}
}

// TestOnlyTheOSLeafTouchesTheHost keeps the real disk, environment, process and
// network out of the core and out of every unit test.
func TestOnlyTheOSLeafTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if slices.Contains(hostLeaves, f.Dir) && (!f.Test || f.BuildsOnlyWith("integration")) {
			continue
		}
		for _, imp := range f.Imports {
			if archtest.IsHost(imp) {
				t.Errorf("%s imports %s; only %v may", f.Path, imp, hostLeaves)
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

// TestNoPackagePrints keeps the standard streams out of the library: a package
// writes to the io.Writer its caller passes.
func TestNoPackagePrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; write to the io.Writer the caller passes", f.Path, use)
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
