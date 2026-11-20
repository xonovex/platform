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

const module = "github.com/xonovex/platform/packages/library/shared-launch-go"

// libraries are the other shared modules the library may import.
var libraries = []string{
	"github.com/xonovex/platform/packages/library/shared-core-go",
	"github.com/xonovex/platform/packages/library/shared-contract-go",
	"github.com/xonovex/platform/packages/library/shared-agent-go",
}

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	"pkg/isolation":                 {},
	"pkg/provision":                 {},
	"pkg/credential":                {},
	"pkg/record":                    {},
	"pkg/sandbox":                   {"pkg/isolation", "pkg/provision"},
	"pkg/launch":                    {"pkg/credential", "pkg/isolation", "pkg/provision", "pkg/sandbox"},
	"pkg/hostlaunch":                {"pkg/isolation"},
	"pkg/isolation/none":            {"pkg/hostlaunch", "pkg/isolation"},
	"pkg/isolation/bwrap":           {"pkg/hostlaunch", "pkg/isolation"},
	"pkg/isolation/docker":          {"pkg/hostlaunch", "pkg/isolation"},
	"pkg/isolation/seatbelt":        {"pkg/hostlaunch", "pkg/isolation"},
	"pkg/isolation/dockersandboxes": {"pkg/hostlaunch", "pkg/isolation"},
	"pkg/isolation/openshell":       {"pkg/hostlaunch", "pkg/isolation"},
	"pkg/provision/none":            {"pkg/provision"},
	"pkg/provision/command":         {"pkg/provision"},
	"pkg/provision/nix":             {"pkg/provision"},
	"pkg/credential/env":            {"pkg/credential"},
	"pkg/credential/file":           {"pkg/credential"},
	"pkg/credential/keychain":       {"pkg/credential"},
	"pkg/credential/pass":           {"pkg/credential"},
	"pkg/credential/secretservice":  {"pkg/credential"},
	"pkg/egress/managed":            {"pkg/isolation"},
	"pkg/oshost":                    {},
	"pkg/oshost/oshosttest":         {},
}

// axes are the directories whose subdirectories are leaves: one variant of
// the axis each, behind the port the axis directory declares or, for egress,
// behind the EgressRoute of the isolation core.
var axes = []string{"pkg/isolation", "pkg/provision", "pkg/credential", "pkg/egress"}

// hostAdapters are the leaves that belong to no axis: the operating-system
// adapter and its in-memory double.
var hostAdapters = []string{"pkg/oshost", "pkg/oshost/oshosttest"}

// osAdapter is the only package that reaches the disk, the environment, the
// processes and the clock of the machine.
const osAdapter = "pkg/oshost"

// machineImports reach the real disk, the environment or the processes.
var machineImports = map[string]bool{
	"os": true, "os/exec": true, "os/user": true, "os/signal": true, "io/ioutil": true,
	"syscall": true, "runtime": true,
}

// networkImports name connections and listeners. A leaf that serves or relays
// traffic needs their types; a core package never does.
var networkImports = map[string]bool{"net": true, "net/http": true, "net/netip": true}

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

func isLibrary(path string) bool {
	for _, library := range libraries {
		if path == library || strings.HasPrefix(path, library+"/") {
			return true
		}
	}
	return false
}

// leafOf returns the axis and the leaf directory of dir, or false when dir is no
// leaf. A host adapter is a leaf of its own, with no axis.
func leafOf(dir string) (string, string, bool) {
	for _, axis := range axes {
		if rest, ok := strings.CutPrefix(dir, axis+"/"); ok {
			leaf, _, _ := strings.Cut(rest, "/")
			return axis, axis + "/" + leaf, true
		}
	}
	if slices.Contains(hostAdapters, dir) {
		return "", dir, true
	}
	return "", "", false
}

// isCore reports whether dir is a core package: one that is no leaf.
func isCore(dir string) bool {
	_, _, leaf := leafOf(dir)
	return !leaf
}

// TestEveryPackageIsDeclared fails for a core package missing from the table
// and for a table entry with no package.
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

// TestImportsFollowTheTable keeps the dependency direction inside the module.
func TestImportsFollowTheTable(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			if rel, inside := strings.CutPrefix(imp, module+"/"); inside && !slices.Contains(allowed[f.Dir], rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.Path, rel, f.Dir, allowed[f.Dir])
			}
		}
	}
}

// TestTheCoreUsesOnlyTheStandardLibraryAndTheSharedLibraries keeps every outside
// module out of the run kernel.
func TestTheCoreUsesOnlyTheStandardLibraryAndTheSharedLibraries(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if !archtest.IsStandard(imp) && !isLibrary(imp) && imp != module && !strings.HasPrefix(imp, module+"/") {
				t.Errorf("%s imports %s; the library uses only the standard library and %v", f.Path, imp, libraries)
			}
		}
	}
}

// TestOnlyTheOSAdapterTouchesTheMachine keeps the real disk, environment and
// processes inside oshost. Only an integration test, which builds under its own
// tag, reaches the machine itself.
func TestOnlyTheOSAdapterTouchesTheMachine(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.BuildsOnlyWith("integration") || (f.Dir == osAdapter && !f.Test) {
			continue
		}
		for _, imp := range f.Imports {
			if machineImports[imp] {
				t.Errorf("%s imports %s; reach the machine through a port", f.Path, imp)
			}
		}
	}
}

// TestTheCoreNamesNoNetwork keeps connections and listeners in the leaves: a
// core package plans and declares, and a leaf serves or relays the traffic.
func TestTheCoreNamesNoNetwork(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if !isCore(f.Dir) {
			continue
		}
		for _, imp := range f.Imports {
			if networkImports[imp] {
				t.Errorf("%s imports %s; a core package names no connection", f.Path, imp)
			}
		}
	}
}

// TestEveryLeafIsReal keeps the leaf rules from passing on an empty tree: each
// axis has a leaf, and every host adapter exists.
func TestEveryLeafIsReal(t *testing.T) {
	leaves := map[string]bool{}
	for _, f := range archtest.Production(moduleFiles(t)) {
		if axis, _, ok := leafOf(f.Dir); ok {
			leaves[axis] = true
			leaves[f.Dir] = true
		}
	}
	for _, axis := range axes {
		if !leaves[axis] {
			t.Errorf("the axis %s has no leaf", axis)
		}
	}
	for _, adapter := range hostAdapters {
		if !leaves[adapter] {
			t.Errorf("the host adapter %s has no Go files", adapter)
		}
	}
}

// TestNoPackageImportsALeaf leaves the choice of a leaf to the registry or the
// composition root of a consumer: no package of the library, and no leaf of
// another axis, imports a leaf from its production code.
func TestNoPackageImportsALeaf(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if _, leaf, isLeaf := leafOf(rel); inside && isLeaf && !strings.HasPrefix(f.Dir, leaf) {
				t.Errorf("%s imports the leaf %s", f.Path, rel)
			}
		}
	}
}

// TestLeavesDoNotImportSiblings keeps each leaf to its own port, in its tests
// as well: a leaf reaches another variant of its axis only through the axis
// core.
func TestLeavesDoNotImportSiblings(t *testing.T) {
	for _, f := range moduleFiles(t) {
		axis, own, isLeaf := leafOf(f.Dir)
		if !isLeaf || axis == "" {
			continue
		}
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if _, leaf, sibling := leafOf(rel); inside && sibling && strings.HasPrefix(rel, axis+"/") && leaf != own {
				t.Errorf("%s reaches the sibling leaf %s", f.Path, rel)
			}
		}
	}
}

// TestNoPackageReadsTheClockOrPrints keeps time an input and output a value.
// The OS adapter is the clock that a consumer passes in.
func TestNoPackageReadsTheClockOrPrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == osAdapter {
			continue
		}
		if archtest.ReadsClock(f.Syntax) {
			t.Errorf("%s reads the clock with time.Now, time.Since or time.Until; pass the instant in", f.Path)
		}
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; return the output to the caller", f.Path, use)
		}
	}
}

// TestNoPackageKeepsState keeps every package free of state that one run
// could leave behind for the next.
func TestNoPackageKeepsState(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, state := range archtest.PackageState(f.Syntax) {
			t.Errorf("%s declares %s; pass state as a value", f.Path, state)
		}
	}
}
