//go:build arch

// Package arch holds the architecture fitness tests of the library. They read
// the source tree, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run.
package arch

import (
	"go/ast"
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
	"pkg/isolation":  {},
	"pkg/provision":  {},
	"pkg/credential": {},
	"pkg/record":     {},
	"pkg/sandbox":    {"pkg/isolation", "pkg/provision"},
	"pkg/launch":     {"pkg/credential", "pkg/isolation", "pkg/provision", "pkg/sandbox"},
}

// axes are the directories whose subdirectories are leaves: one variant of
// the axis each, behind the port the axis directory declares.
var axes = []string{"pkg/isolation", "pkg/provision", "pkg/credential"}

// hostImports reach the real disk, the environment, the process or the network.
var hostImports = map[string]bool{
	"os": true, "os/exec": true, "os/user": true, "os/signal": true, "io/ioutil": true,
	"net": true, "net/http": true, "net/netip": true, "syscall": true, "runtime": true,
}

type sourceFile struct {
	path    string
	dir     string
	test    bool
	imports []string
	parsed  *ast.File
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
		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		file := sourceFile{path: path, dir: filepath.ToSlash(rel), test: strings.HasSuffix(path, "_test.go"), parsed: parsed}
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
// leaf.
func leafOf(dir string) (string, string, bool) {
	for _, axis := range axes {
		if rest, ok := strings.CutPrefix(dir, axis+"/"); ok {
			leaf, _, _ := strings.Cut(rest, "/")
			return axis, axis + "/" + leaf, true
		}
	}
	return "", "", false
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

// TestNoPackageTouchesTheHost keeps the real disk, environment, process and
// network out of the core and out of every unit test.
func TestNoPackageTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if hostImports[imp] {
				t.Errorf("%s imports %s; reach the machine through a port", f.path, imp)
			}
		}
	}
}

// TestNoPackageImportsALeaf leaves the choice of a leaf to the composition root
// of a consumer.
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

// TestLeavesDoNotImportSiblings keeps each leaf to its own port: a leaf reaches
// another variant of its axis only through the axis core.
func TestLeavesDoNotImportSiblings(t *testing.T) {
	for _, f := range moduleFiles(t) {
		axis, own, isLeaf := leafOf(f.dir)
		if !isLeaf {
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
func TestNoPackageReadsTheClockOrPrints(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
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
