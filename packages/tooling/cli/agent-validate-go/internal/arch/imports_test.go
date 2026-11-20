//go:build arch

// Package arch holds the architecture fitness tests of the validator. They read
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
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	module   = "github.com/xonovex/platform/packages/tooling/cli/agent-validate-go"
	contract = "github.com/xonovex/platform/packages/library/shared-contract-go/"
)

// checkPrefix starts the directory of every concrete check.
const checkPrefix = "internal/check/"

// support is what a check may import: the port, the readers of the file types a
// check parses, and the shared helpers. Never another check.
var support = []string{
	"internal/budget", "internal/catalog", "internal/check", "internal/component", "internal/document",
	"internal/glob", "internal/markdown", "internal/tree",
}

// core lists, for each package outside the checks, the packages of this module
// it may import from its non-test files. A package that is missing from the
// table, or that imports more than its entry lists, fails the test, so a new
// dependency edge is always a decision in this file.
var core = map[string][]string{
	"internal/arch":      {},
	"cmd/agent-validate": {"internal/cmd", "internal/gitrepo", "internal/registry", "internal/tree"},
	"internal/budget":    {"internal/catalog", "internal/markdown", "internal/tree"},
	"internal/catalog":   {"internal/tree"},
	"internal/check":     {"internal/tree"},
	"internal/checktest": {"internal/check", "internal/tree"},
	"internal/cmd":       {"internal/budget", "internal/check", "internal/project", "internal/report", "internal/runner", "internal/tree"},
	"internal/component": {"internal/check", "internal/document", "internal/tree"},
	"internal/document":  {},
	"internal/gitrepo":   {},
	"internal/glob":      {},
	"internal/markdown":  {},
	"internal/project":   {"internal/check", "internal/tree"},
	"internal/registry":  {"internal/check"},
	"internal/report":    {},
	"internal/runner":    {"internal/check", "internal/project"},
	"internal/tree":      {},
}

// thirdParty lists the packages that may import a module outside the standard
// library and the contract, with the modules each may use.
var thirdParty = map[string][]string{
	"internal/document":        {"gopkg.in/yaml.v3"},
	"internal/check/kit":       {"gopkg.in/yaml.v3"},
	"internal/check/workspace": {"gopkg.in/yaml.v3"},
	"internal/cmd":             {"github.com/spf13/cobra"},
}

// contractAdapters bind a contract port to a technology. The composition root
// binds the real ones; a test double binds the memory one.
var contractAdapters = map[string][]string{
	"pkg/fs/osfs":           {"cmd/agent-validate"},
	"pkg/wiring":            {"cmd/agent-validate", "internal/checktest"},
	"pkg/fs/memfs":          {"internal/checktest"},
	"pkg/jsonschema/tekuri": {},
}

// hostImports reach the real disk, the environment, the process or the network.
var hostImports = []string{
	"os", "os/exec", "os/user", "os/signal", "io/ioutil", "net", "net/http", "syscall",
}

// hostAdapters may import hostImports from their non-test files.
var hostAdapters = []string{"cmd/agent-validate"}

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
	for _, top := range []string{"cmd", "internal"} {
		err := filepath.WalkDir(filepath.Join(root, top), func(path string, entry fs.DirEntry, err error) error {
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
			file := sourceFile{path: path, dir: filepath.ToSlash(rel), test: strings.HasSuffix(path, "_test.go")}
			for _, spec := range parsed.Imports {
				value, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return err
				}
				file.imports = append(file.imports, value)
			}
			file.clock = callsTimeNow(parsed)
			files = append(files, file)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

// callsTimeNow reports whether the file reads the clock.
func callsTimeNow(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if ok && pkg.Name == "time" && slices.Contains([]string{"Now", "Since", "Until"}, selector.Sel.Name) {
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

func isCheck(dir string) bool { return strings.HasPrefix(dir, checkPrefix) }

// TestEveryPackageIsDeclared fails for a package outside the checks that the
// table does not name, and for a table entry with no package.
func TestEveryPackageIsDeclared(t *testing.T) {
	present := map[string]bool{}
	for _, f := range moduleFiles(t) {
		present[f.dir] = true
	}
	for dir := range present {
		if _, ok := core[dir]; !ok && !isCheck(dir) {
			t.Errorf("package %s is not declared in the core table", dir)
		}
	}
	for dir := range core {
		if !present[dir] {
			t.Errorf("the core table names %s, which has no Go files", dir)
		}
	}
}

// TestInternalImportsFollowTheRules keeps the dependency direction of the
// hexagon: a check knows the port and the shared readers, the plan and the
// command tree know the port, and only the composition root knows them all.
func TestInternalImportsFollowTheRules(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		permitted := core[f.dir]
		if isCheck(f.dir) {
			permitted = support
		}
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if !inside {
				continue
			}
			if f.dir == "internal/registry" && isCheck(rel) {
				continue
			}
			if !slices.Contains(permitted, rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.path, rel, f.dir, permitted)
			}
		}
	}
}

// TestOnlyTheRegistryImportsAConcreteCheck keeps a check behind the port: no
// other package, test included except the check's own, names one.
func TestOnlyTheRegistryImportsAConcreteCheck(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if !inside || !isCheck(rel) || rel == "internal/check" {
				continue
			}
			ownTest := f.test && f.dir == rel
			if f.dir != "internal/registry" && !ownTest {
				t.Errorf("%s imports the check %s; only internal/registry may", f.path, rel)
			}
		}
	}
}

// TestContractAdaptersStayAtTheEdge keeps the real disk, the schema engine and
// the wiring out of the core.
func TestContractAdaptersStayAtTheEdge(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, contract)
			if !inside {
				continue
			}
			allowed, restricted := contractAdapters[rel]
			if !restricted {
				continue
			}
			if f.test {
				// A test may bind the memory adapter and the wiring, to build a
				// context; it never binds the disk or the schema engine.
				if rel == "pkg/fs/osfs" || rel == "pkg/jsonschema/tekuri" {
					t.Errorf("%s imports %s; a unit test stays off the disk", f.path, rel)
				}
				continue
			}
			if !slices.Contains(allowed, f.dir) {
				t.Errorf("%s imports %s; only %v may", f.path, rel, allowed)
			}
		}
	}
}

// TestOnlyDeclaredPackagesImportThirdParty keeps the core on the standard
// library and the contract.
func TestOnlyDeclaredPackagesImportThirdParty(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if isStdlib(imp) || strings.HasPrefix(imp, module) || strings.HasPrefix(imp, contract) {
				continue
			}
			ok := false
			for _, want := range thirdParty[f.dir] {
				ok = ok || strings.HasPrefix(imp, want)
			}
			if !ok {
				t.Errorf("%s imports %s; %s may import only %v", f.path, imp, f.dir, thirdParty[f.dir])
			}
		}
	}
}

// TestOnlyTheCompositionRootTouchesTheHost keeps the real disk, environment,
// process and network out of the core and out of every unit test.
func TestOnlyTheCompositionRootTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if slices.Contains(hostAdapters, f.dir) && !f.test {
			continue
		}
		for _, imp := range f.imports {
			if slices.Contains(hostImports, imp) {
				t.Errorf("%s imports %s; only the non-test files of %v may", f.path, imp, hostAdapters)
			}
		}
	}
}

// TestNoPackageReadsTheClock keeps time an input.
func TestNoPackageReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.clock {
			t.Errorf("%s reads the clock; pass the instant in", f.path)
		}
	}
}

// TestChecksDoNotShareState keeps each check independent: no check keeps a
// mutable package variable, which a second run in one process would inherit.
func TestChecksDoNotShareState(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()
	var offenders []string
	for _, f := range nonTest(moduleFiles(t)) {
		if !isCheck(f.dir) {
			continue
		}
		parsed, err := parser.ParseFile(fset, f.path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range parsed.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				for _, name := range spec.(*ast.ValueSpec).Names {
					if !isImmutableName(spec.(*ast.ValueSpec)) {
						offenders = append(offenders, filepath.ToSlash(mustRel(t, root, f.path))+": "+name.Name)
					}
				}
			}
		}
	}
	sort.Strings(offenders)
	for _, offender := range offenders {
		t.Errorf("%s is a package variable that holds anything but a compiled expression, a table or an error", offender)
	}
}

func mustRel(t *testing.T, root, path string) string {
	t.Helper()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

// isImmutableName accepts a package variable initialised once and never
// reassigned: a regexp.MustCompile, a composite literal or a call of errors.New.
// A variable without an initialiser can be assigned later, so it fails.
func isImmutableName(spec *ast.ValueSpec) bool {
	if len(spec.Values) == 0 {
		return false
	}
	for _, value := range spec.Values {
		switch v := value.(type) {
		case *ast.CompositeLit, *ast.BasicLit, *ast.UnaryExpr:
		case *ast.CallExpr:
			selector, ok := v.Fun.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			pkg, ok := selector.X.(*ast.Ident)
			if !ok || !(pkg.Name == "regexp" || pkg.Name == "errors") {
				return false
			}
		default:
			return false
		}
	}
	return true
}
