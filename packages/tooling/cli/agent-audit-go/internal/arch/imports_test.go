//go:build arch

// Package arch holds the architecture fitness tests of the tool. They read the
// source tree, so they carry the arch build tag and run in the go-test-arch
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

const module = "github.com/xonovex/platform/packages/tooling/cli/agent-audit-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	// The composition root binds every adapter.
	"cmd/agent-audit": {"internal/cmd", "internal/process/osexec", "internal/transport/plugins", "internal/web/nethttp"},

	// The driving adapter reaches the core and the ports, never an adapter.
	"internal/cmd": {
		"internal/audit", "internal/catalogs", "internal/compare", "internal/compare/gitbase",
		"internal/legacy", "internal/observe", "internal/process", "internal/render",
		"internal/rules", "internal/stamp",
	},

	// The core: rules are pure, the orchestrator composes them with the ports.
	"internal/audit":    {"internal/observe", "internal/rules"},
	"internal/catalogs": {},
	"internal/compare":  {"internal/rules"},
	"internal/legacy":   {},
	"internal/observe":  {},
	"internal/render":   {"internal/audit"},
	"internal/rules":    {"internal/observe"},
	"internal/stamp":    {},

	// Ports and the test doubles of the ports.
	"internal/process":             {},
	"internal/process/processtest": {"internal/process"},
	"internal/web":                 {},
	"internal/web/webtest":         {"internal/web"},

	// Adapters: one port each, plus the registry that names them.
	"internal/compare/gitbase":      {"internal/gitfs", "internal/process"},
	"internal/gitfs":                {"internal/process"},
	"internal/process/osexec":       {"internal/process"},
	"internal/web/nethttp":          {"internal/web"},
	"internal/transport/filesystem": {"internal/observe"},
	"internal/transport/git":        {"internal/observe", "internal/process"},
	"internal/transport/http":       {"internal/observe", "internal/web"},
	"internal/transport/plugins": {
		"internal/observe", "internal/process", "internal/transport/filesystem",
		"internal/transport/git", "internal/transport/http", "internal/web",
	},
}

// concrete lists the packages that bind a port to a technology or name a
// concrete transport, with the one package that may import each from non-test
// code. The test doubles are for tests only.
var concrete = map[string]string{
	"internal/process/osexec":       "cmd/agent-audit",
	"internal/web/nethttp":          "cmd/agent-audit",
	"internal/transport/filesystem": "internal/transport/plugins",
	"internal/transport/git":        "internal/transport/plugins",
	"internal/transport/http":       "internal/transport/plugins",
	"internal/transport/plugins":    "cmd/agent-audit",
	"internal/process/processtest":  "",
	"internal/web/webtest":          "",
}

// thirdParty lists the packages that may import a module outside the standard
// library and this repository, with the module prefix each may use.
var thirdParty = map[string]string{
	"internal/cmd": "github.com/spf13/cobra",
}

// hostImports reach the real disk, the environment, the process or the network.
var hostImports = map[string]bool{
	"os": true, "os/exec": true, "os/user": true, "os/signal": true,
	"io/ioutil": true, "net": true, "net/http": true, "syscall": true,
}

// hostAdapters may import hostImports from non-test code.
var hostAdapters = map[string]bool{
	"cmd/agent-audit":         true,
	"internal/process/osexec": true,
	"internal/web/nethttp":    true,
}

// clockRoots may read the clock: the composition root hands the instant to the
// rest of the tool.
var clockRoots = map[string]bool{"cmd/agent-audit": true}

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
		if !f.test && f.dir != "internal/arch" {
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

// TestInternalImportsFollowTheTable keeps the dependency direction of the
// hexagon: the core knows the ports, an adapter knows one port, and only the
// composition root knows them all.
func TestInternalImportsFollowTheTable(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && !contains(allowed[f.dir], rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.path, rel, f.dir, allowed[f.dir])
			}
		}
	}
}

// TestConcretePackagesHaveOneImporter keeps a technology behind its port: the
// registry is the one file that names the concrete transports, and the main
// package the one that names the adapters and the registry.
func TestConcretePackagesHaveOneImporter(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			owner, isConcrete := concrete[rel]
			if inside && isConcrete && f.dir != owner {
				t.Errorf("%s imports %s; only %q may", f.path, rel, owner)
			}
		}
	}
}

// TestTheRegistryIsTheOnlyFileThatNamesTransports pins the one-file rule: the
// registry package holds a single source file.
func TestTheRegistryIsTheOnlyFileThatNamesTransports(t *testing.T) {
	var registry []string
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir == "internal/transport/plugins" {
			registry = append(registry, filepath.Base(f.path))
		}
	}
	if len(registry) != 1 {
		t.Errorf("internal/transport/plugins holds %v; it must hold exactly one file", registry)
	}
}

// TestOnlyDeclaredPackagesImportThirdParty keeps the core on the standard
// library and the contract module.
func TestOnlyDeclaredPackagesImportThirdParty(t *testing.T) {
	const contract = "github.com/xonovex/platform/packages/library/shared-contract-go"
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if isStdlib(imp) || strings.HasPrefix(imp, module) || strings.HasPrefix(imp, contract) {
				continue
			}
			want, ok := thirdParty[f.dir]
			if !ok || !strings.HasPrefix(imp, want) {
				t.Errorf("%s imports %s; %s may import no outside module except %q", f.path, imp, f.dir, want)
			}
		}
	}
}

// TestOnlyHostAdaptersTouchTheHost keeps the disk, environment, process and
// network out of the core and out of every unit test.
func TestOnlyHostAdaptersTouchTheHost(t *testing.T) {
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

// TestOnlyTheCompositionRootReadsTheClock keeps time an input: a function takes
// the instant it needs as an argument.
func TestOnlyTheCompositionRootReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if f.clock && (!clockRoots[f.dir] || f.test) {
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
