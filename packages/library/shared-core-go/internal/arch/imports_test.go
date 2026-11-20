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

const module = "github.com/xonovex/platform/packages/library/shared-core-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	"pkg/archtest":            {},
	"pkg/colors":              {},
	"pkg/envutil":             {},
	"pkg/logging":             {"pkg/colors"},
	"pkg/shell":               {},
	"pkg/process":             {},
	"pkg/process/osexec":      {"pkg/envutil", "pkg/process"},
	"pkg/process/processtest": {"pkg/process"},
	"pkg/scratch":             {},
	"pkg/scratch/osscratch":   {},
	"pkg/scratch/scratchtest": {},
}

// leaves bind a port to the operating system or to a test double. No package of
// this module imports one from non-test code: a consumer's composition root
// chooses the leaf.
var leaves = []string{
	"pkg/process/osexec", "pkg/process/processtest",
	"pkg/scratch/osscratch", "pkg/scratch/scratchtest",
}

// hostLeaves may import a host package from non-test code, and their integration
// tests may too.
var hostLeaves = []string{"pkg/process/osexec", "pkg/scratch/osscratch"}

func moduleFiles(t *testing.T) []archtest.File {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	files, err := archtest.Load(os.DirFS(filepath.Dir(filepath.Dir(filepath.Dir(here)))), "pkg")
	if err != nil {
		t.Fatal(err)
	}
	return files
}

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

// TestInternalImportsFollowTheTable keeps the dependency direction: a leaf knows
// its port, and a port knows no leaf.
func TestInternalImportsFollowTheTable(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && !slices.Contains(allowed[f.Dir], rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.Path, rel, f.Dir, allowed[f.Dir])
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

// TestTheLibraryUsesOnlyTheStandardLibrary keeps the core free of outside
// modules.
func TestTheLibraryUsesOnlyTheStandardLibrary(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if !archtest.IsStandard(imp) && !strings.HasPrefix(imp, module) {
				t.Errorf("%s imports %s; the library uses only the standard library", f.Path, imp)
			}
		}
	}
}

// TestOnlyTheOSLeavesTouchTheHost keeps the real disk, environment, process and
// network out of the core and out of every unit test. The integration tests of
// an OS leaf exercise the real host.
func TestOnlyTheOSLeavesTouchTheHost(t *testing.T) {
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

// TestNoPackageReadsTheClock keeps time an input: a function takes the instant
// it needs as an argument.
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
