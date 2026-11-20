//go:build arch

// Package arch holds the architecture fitness tests of the library. They read
// the source tree, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run.
package arch

import (
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
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

type sourceFile struct {
	path string
	dir  string
	test bool
	// integration marks a file that builds only under the integration tag: it
	// may start real programs and touch the machine.
	integration bool
	imports     []string
	parsed      *ast.File
}

// integrationOnly reports whether the build constraint of file holds with the
// integration tag and fails without it.
func integrationOnly(file *ast.File) bool {
	for _, group := range file.Comments {
		if group.Pos() >= file.Package {
			break
		}
		for _, comment := range group.List {
			if !constraint.IsGoBuild(comment.Text) {
				continue
			}
			expr, err := constraint.Parse(comment.Text)
			if err != nil {
				return false
			}
			return expr.Eval(func(tag string) bool { return tag == "integration" }) && !expr.Eval(func(string) bool { return false })
		}
	}
	return false
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func moduleFiles(t *testing.T) []sourceFile {
	t.Helper()
	root := moduleRoot(t)
	fset := token.NewFileSet()
	var files []sourceFile
	err := filepath.WalkDir(filepath.Join(root, "pkg"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution|parser.ParseComments)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		file := sourceFile{path: path, dir: filepath.ToSlash(rel), test: strings.HasSuffix(path, "_test.go"), integration: integrationOnly(parsed), parsed: parsed}
		for _, spec := range parsed.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			file.imports = append(file.imports, value)
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func nonTest(files []sourceFile) []sourceFile {
	var out []sourceFile
	for _, f := range files {
		if !f.test {
			out = append(out, f)
		}
	}
	return out
}

func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
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
	for _, f := range nonTest(moduleFiles(t)) {
		present[f.dir] = true
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
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			if rel, inside := strings.CutPrefix(imp, module+"/"); inside && !slices.Contains(allowed[f.dir], rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.path, rel, f.dir, allowed[f.dir])
			}
		}
	}
}

// TestTheCoreUsesOnlyTheStandardLibraryAndTheSharedLibraries keeps every outside
// module out of the run kernel.
func TestTheCoreUsesOnlyTheStandardLibraryAndTheSharedLibraries(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if !isStdlib(imp) && !isLibrary(imp) && imp != module && !strings.HasPrefix(imp, module+"/") {
				t.Errorf("%s imports %s; the library uses only the standard library and %v", f.path, imp, libraries)
			}
		}
	}
}

// TestOnlyTheOSAdapterTouchesTheMachine keeps the real disk, environment and
// processes inside oshost. Only an integration test, which builds under its own
// tag, reaches the machine itself.
func TestOnlyTheOSAdapterTouchesTheMachine(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.integration || (f.dir == osAdapter && !f.test) {
			continue
		}
		for _, imp := range f.imports {
			if machineImports[imp] {
				t.Errorf("%s imports %s; reach the machine through a port", f.path, imp)
			}
		}
	}
}

// TestTheCoreNamesNoNetwork keeps connections and listeners in the leaves: a
// core package plans and declares, and a leaf serves or relays the traffic.
func TestTheCoreNamesNoNetwork(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if !isCore(f.dir) {
			continue
		}
		for _, imp := range f.imports {
			if networkImports[imp] {
				t.Errorf("%s imports %s; a core package names no connection", f.path, imp)
			}
		}
	}
}

// TestEveryLeafIsReal keeps the leaf rules from passing on an empty tree: each
// axis has a leaf, and every host adapter exists.
func TestEveryLeafIsReal(t *testing.T) {
	leaves := map[string]bool{}
	for _, f := range nonTest(moduleFiles(t)) {
		if axis, _, ok := leafOf(f.dir); ok {
			leaves[axis] = true
			leaves[f.dir] = true
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
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if _, leaf, isLeaf := leafOf(rel); inside && isLeaf && !strings.HasPrefix(f.dir, leaf) {
				t.Errorf("%s imports the leaf %s", f.path, rel)
			}
		}
	}
}

// TestLeavesDoNotImportSiblings keeps each leaf to its own port, in its tests
// as well: a leaf reaches another variant of its axis only through the axis
// core.
func TestLeavesDoNotImportSiblings(t *testing.T) {
	for _, f := range moduleFiles(t) {
		axis, own, isLeaf := leafOf(f.dir)
		if !isLeaf || axis == "" {
			continue
		}
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if _, leaf, sibling := leafOf(rel); inside && sibling && strings.HasPrefix(rel, axis+"/") && leaf != own {
				t.Errorf("%s reaches the sibling leaf %s", f.path, rel)
			}
		}
	}
}

// TestNoPackageReadsTheClockOrPrints keeps time an input and output a value.
// The OS adapter is the clock that a consumer passes in.
func TestNoPackageReadsTheClockOrPrints(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir == osAdapter {
			continue
		}
		ast.Inspect(f.parsed, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			switch {
			case pkg.Name == "time" && (selector.Sel.Name == "Now" || selector.Sel.Name == "Since" || selector.Sel.Name == "Until"):
				t.Errorf("%s reads the clock with time.%s; pass the instant in", f.path, selector.Sel.Name)
			case pkg.Name == "fmt" && strings.HasPrefix(selector.Sel.Name, "Print"):
				t.Errorf("%s calls fmt.%s", f.path, selector.Sel.Name)
			}
			return true
		})
	}
}
