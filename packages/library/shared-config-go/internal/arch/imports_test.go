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

// hostImports reach the real disk, the environment, the process or the network.
var hostImports = map[string]bool{
	"os": true, "os/exec": true, "os/user": true, "os/signal": true,
	"io/ioutil": true, "net": true, "net/http": true, "syscall": true,
}

// hostLeaves may import hostImports from non-test code, and their integration
// tests may too.
var hostLeaves = []string{"pkg/osfiles"}

type sourceFile struct {
	path        string
	dir         string
	test        bool
	integration bool
	imports     []string
	clock       bool
	parsed      *ast.File
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
		parsed, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		file := sourceFile{
			path:        path,
			dir:         filepath.ToSlash(rel),
			test:        strings.HasSuffix(path, "_test.go"),
			integration: hasBuildTag(parsed, "integration"),
			clock:       callsTimeNow(parsed),
			parsed:      parsed,
		}
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

// hasBuildTag reports whether the file carries a //go:build line naming tag.
func hasBuildTag(file *ast.File, tag string) bool {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if expr, ok := strings.CutPrefix(comment.Text, "//go:build "); ok && slices.Contains(strings.Fields(expr), tag) {
				return true
			}
		}
	}
	return false
}

// callsTimeNow reports whether the file reads the clock through time.Now,
// time.Since or time.Until.
func callsTimeNow(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if ok && pkg.Name == "time" && (selector.Sel.Name == "Now" || selector.Sel.Name == "Since" || selector.Sel.Name == "Until") {
			found = true
		}
		return true
	})
	return found
}

func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".")
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

// TestEveryPackageIsDeclared fails for a package missing from the table and for
// a table entry with no package.
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

// TestImportsFollowTheTables keeps the dependency direction inside the module
// and toward the contract.
func TestImportsFollowTheTables(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			if rel, inside := strings.CutPrefix(imp, module+"/"); inside && !slices.Contains(allowed[f.dir], rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.path, rel, f.dir, allowed[f.dir])
			}
			if rel, inside := strings.CutPrefix(imp, contract+"/"); inside && !slices.Contains(contracts[f.dir], rel) {
				t.Errorf("%s imports the contract package %s; %s may import only %v", f.path, rel, f.dir, contracts[f.dir])
			}
		}
	}
}

// TestNoPackageImportsALeaf leaves the choice of a leaf to the consumer.
func TestNoPackageImportsALeaf(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && slices.Contains(leaves, rel) {
				t.Errorf("%s imports the leaf %s", f.path, rel)
			}
		}
	}
}

// TestTheLibraryUsesOnlyTheStandardLibraryAndTheContract keeps outside modules
// out of the configuration core.
func TestTheLibraryUsesOnlyTheStandardLibraryAndTheContract(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if !isStdlib(imp) && !strings.HasPrefix(imp, module) && !strings.HasPrefix(imp, contract+"/") {
				t.Errorf("%s imports %s; the library uses only the standard library and the contract", f.path, imp)
			}
		}
	}
}

// TestOnlyTheOSLeafTouchesTheHost keeps the real disk, environment, process and
// network out of the core and out of every unit test.
func TestOnlyTheOSLeafTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if slices.Contains(hostLeaves, f.dir) && (!f.test || f.integration) {
			continue
		}
		for _, imp := range f.imports {
			if hostImports[imp] {
				t.Errorf("%s imports %s; only %v may", f.path, imp, hostLeaves)
			}
		}
	}
}

// TestNoPackageReadsTheClock keeps time an input.
func TestNoPackageReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.clock {
			t.Errorf("%s reads the clock with time.Now, time.Since or time.Until; pass the instant in", f.path)
		}
	}
}
