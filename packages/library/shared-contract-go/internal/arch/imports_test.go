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
	"sort"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/xonovex/platform/packages/library/shared-contract-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	"pkg/condition":         {},
	"pkg/discovery":         {"pkg/fs"},
	"pkg/findings":          {},
	"pkg/fs":                {},
	"pkg/fs/memfs":          {"pkg/fs"},
	"pkg/fs/osfs":           {"pkg/fs"},
	"pkg/jsonschema":        {},
	"pkg/jsonschema/tekuri": {"pkg/jsonschema"},
	"pkg/report":            {},
	"pkg/sarif":             {"pkg/findings"},
	"pkg/schemas":           {},
	"pkg/sources":           {"pkg/findings", "pkg/fs", "pkg/jsonschema", "pkg/schemas"},
	"pkg/wiring":            {"pkg/fs", "pkg/jsonschema", "pkg/jsonschema/tekuri", "pkg/schemas", "pkg/sources"},
}

// adapters are the packages that bind a port to a concrete technology. Only the
// composition root imports them from non-test code.
var adapters = []string{"pkg/fs/memfs", "pkg/fs/osfs", "pkg/jsonschema/tekuri"}

// thirdParty lists the packages that may import a module outside the standard
// library, with the modules each may use.
var thirdParty = map[string]string{
	"pkg/jsonschema/tekuri": "github.com/santhosh-tekuri/jsonschema/v6",
}

// hostImports reach the real disk, the environment, the process or the network.
var hostImports = map[string]bool{
	"os": true, "os/exec": true, "os/user": true, "os/signal": true,
	"io/ioutil": true, "net": true, "net/http": true, "syscall": true,
}

// hostAdapters may import hostImports.
var hostAdapters = map[string]bool{"pkg/fs/osfs": true}

type sourceFile struct {
	path    string
	dir     string
	test    bool
	imports []string
	clock   bool
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
		parsed, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly|parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		full, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		file := sourceFile{
			path: path,
			dir:  filepath.ToSlash(mustRel(t, root, filepath.Dir(path))),
			test: strings.HasSuffix(path, "_test.go"),
		}
		for _, spec := range parsed.Imports {
			value, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			file.imports = append(file.imports, value)
		}
		file.clock = callsTimeNow(full)
		files = append(files, file)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func mustRel(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

// callsTimeNow reports whether the file reads the clock through time.Now or
// time.Since.
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

func contains(list []string, item string) bool {
	for _, candidate := range list {
		if candidate == item {
			return true
		}
	}
	return false
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
	for _, f := range moduleFiles(t) {
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

// TestInternalImportsFollowTheTable keeps the dependency direction of the
// hexagon: the core packages know the ports, adapters know one port each, and
// only the composition root knows them all.
func TestInternalImportsFollowTheTable(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		permitted := allowed[f.dir]
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if !inside {
				continue
			}
			if !contains(permitted, rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.path, rel, f.dir, permitted)
			}
		}
	}
}

// TestOnlyTheCompositionRootImportsAdapters keeps the concrete technology
// behind the ports.
func TestOnlyTheCompositionRootImportsAdapters(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir == "pkg/wiring" {
			continue
		}
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && contains(adapters, rel) {
				t.Errorf("%s imports the adapter %s; only pkg/wiring may", f.path, rel)
			}
		}
	}
}

// TestOnlyDeclaredAdaptersImportThirdParty keeps the core on the standard
// library.
func TestOnlyDeclaredAdaptersImportThirdParty(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if isStdlib(imp) || strings.HasPrefix(imp, module) {
				continue
			}
			want, ok := thirdParty[f.dir]
			if !ok || !strings.HasPrefix(imp, want) {
				t.Errorf("%s imports %s; %s may import no module outside the standard library except %q", f.path, imp, f.dir, want)
			}
		}
	}
}

// TestOnlyTheOSAdapterTouchesTheHost keeps the real disk, environment, process
// and network out of the core and out of every unit test.
func TestOnlyTheOSAdapterTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if hostAdapters[f.dir] && !f.test {
			continue
		}
		for _, imp := range f.imports {
			if hostImports[imp] {
				t.Errorf("%s imports %s; only the non-test files of %v may", f.path, imp, keys(hostAdapters))
			}
		}
	}
}

// TestNoPackageReadsTheClock keeps time an input: a function takes the instant
// it needs as an argument.
func TestNoPackageReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.clock {
			t.Errorf("%s reads the clock with time.Now, time.Since or time.Until; pass the instant in", f.path)
		}
	}
}

func keys(set map[string]bool) []string {
	var out []string
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
