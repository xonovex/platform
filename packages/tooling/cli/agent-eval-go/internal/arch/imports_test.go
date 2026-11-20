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

const module = "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	// The composition root binds every adapter and the registries.
	"cmd/agent-eval": {
		"internal/cmd", "internal/execute", "internal/harness", "internal/predicate", "internal/registry",
	},

	// The driving adapter reaches the core and the ports, never an adapter.
	"internal/cmd": {
		"internal/baseline", "internal/budget", "internal/catalog", "internal/cost", "internal/document",
		"internal/harness", "internal/migrate", "internal/model", "internal/predicate", "internal/protocol",
		"internal/report", "internal/run", "internal/settings", "internal/store",
	},

	// The one file that names the concrete harnesses, protocols and predicates.
	"internal/registry": {
		"internal/harness", "internal/harness/claude", "internal/harness/codex", "internal/judge",
		"internal/predicate", "internal/predicate/filesystem", "internal/predicate/llm", "internal/predicate/session",
		"internal/protocol", "internal/protocol/instruction", "internal/protocol/orchestration",
		"internal/protocol/output", "internal/protocol/routing", "internal/protocol/trigger",
	},

	// The core: documents, the model of an eval and the orchestration of a run.
	"internal/model":    {},
	"internal/document": {"internal/model"},
	"internal/ablation": {"internal/model"},
	"internal/budget":   {"internal/model"},
	"internal/catalog":  {"internal/model"},
	"internal/settings": {"internal/budget", "internal/model"},
	"internal/store":    {"internal/document", "internal/model"},
	"internal/baseline": {"internal/document", "internal/model"},
	"internal/cost":     {"internal/document", "internal/model"},
	"internal/migrate":  {"internal/document", "internal/model"},
	"internal/report":   {"internal/model", "internal/run"},
	"internal/run":      {"internal/budget", "internal/model", "internal/protocol", "internal/store"},
	"internal/execute":  {"internal/harness"},
	"internal/judge":    {"internal/harness", "internal/model"},

	// Ports. A port holds the contract and the test double holds a fake.
	"internal/harness":   {"internal/model"},
	"internal/predicate": {"internal/harness", "internal/judge", "internal/model"},
	"internal/protocol": {
		"internal/catalog", "internal/document", "internal/harness", "internal/judge", "internal/model",
		"internal/predicate", "internal/settings",
	},
	"internal/protocol/protocoltest": {
		"internal/catalog", "internal/document", "internal/harness", "internal/model", "internal/predicate",
		"internal/predicate/filesystem", "internal/predicate/session", "internal/protocol",
	},

	// Leaves: one variant of one axis each. A leaf knows its port and never a sibling.
	"internal/harness/claude":         {"internal/harness", "internal/model"},
	"internal/harness/codex":          {"internal/harness", "internal/model"},
	"internal/predicate/filesystem":   {"internal/model", "internal/predicate"},
	"internal/predicate/llm":          {"internal/judge", "internal/model", "internal/predicate"},
	"internal/predicate/session":      {"internal/harness", "internal/model", "internal/predicate"},
	"internal/protocol/trigger":       {"internal/harness", "internal/model", "internal/protocol", "internal/store"},
	"internal/protocol/routing":       {"internal/catalog", "internal/harness", "internal/model", "internal/protocol", "internal/store"},
	"internal/protocol/output":        {"internal/harness", "internal/judge", "internal/model", "internal/predicate", "internal/protocol", "internal/settings"},
	"internal/protocol/instruction":   {"internal/ablation", "internal/harness", "internal/judge", "internal/model", "internal/predicate", "internal/protocol", "internal/settings"},
	"internal/protocol/orchestration": {"internal/document", "internal/harness", "internal/judge", "internal/model", "internal/predicate", "internal/protocol", "internal/settings"},
}

// library is the path prefix of the shared Go libraries.
const library = "github.com/xonovex/platform/packages/library/"

// libraries lists, for each package, the library packages it may import from
// its non-test files. A package missing from the table imports no library, so a
// new edge into a library is a decision in this file.
var libraries = map[string][]string{
	"cmd/agent-eval": {
		"shared-agent-go/pkg/agents",
		"shared-config-go/pkg/osfiles",
		"shared-contract-go/pkg/findings", "shared-contract-go/pkg/fs/osfs", "shared-contract-go/pkg/wiring",
		"shared-core-go/pkg/process/osexec", "shared-core-go/pkg/scratch/osscratch",
	},
	"internal/baseline": {"shared-contract-go/pkg/fs"},
	"internal/budget":   {"shared-config-go/pkg/section/eval"},
	"internal/catalog":  {"shared-agent-go/pkg/types", "shared-contract-go/pkg/fs"},
	"internal/cmd": {
		"shared-agent-go/pkg/agents", "shared-agent-go/pkg/types",
		"shared-config-go/pkg/discover", "shared-config-go/pkg/explain", "shared-config-go/pkg/merge",
		"shared-config-go/pkg/resolve", "shared-config-go/pkg/section/eval", "shared-config-go/pkg/shard",
		"shared-config-go/pkg/spec", "shared-config-go/pkg/trust",
		"shared-contract-go/pkg/findings", "shared-contract-go/pkg/fs", "shared-contract-go/pkg/jsonschema",
		"shared-contract-go/pkg/report",
	},
	"internal/cost":                  {"shared-contract-go/pkg/fs"},
	"internal/document":              {"shared-contract-go/pkg/fs", "shared-contract-go/pkg/jsonschema", "shared-contract-go/pkg/schemas"},
	"internal/execute":               {"shared-contract-go/pkg/fs", "shared-core-go/pkg/process", "shared-core-go/pkg/scratch"},
	"internal/harness":               {"shared-core-go/pkg/process"},
	"internal/harness/claude":        {"shared-agent-go/pkg/agents", "shared-agent-go/pkg/types", "shared-core-go/pkg/process"},
	"internal/harness/codex":         {"shared-agent-go/pkg/agents", "shared-agent-go/pkg/types", "shared-core-go/pkg/process"},
	"internal/migrate":               {"shared-contract-go/pkg/fs", "shared-contract-go/pkg/schemas"},
	"internal/model":                 {"shared-contract-go/pkg/condition"},
	"internal/protocol":              {"shared-contract-go/pkg/fs"},
	"internal/protocol/protocoltest": {"shared-agent-go/pkg/agents", "shared-agent-go/pkg/types", "shared-contract-go/pkg/fs/memfs", "shared-contract-go/pkg/wiring"},
	"internal/registry":              {"shared-agent-go/pkg/types"},
	"internal/report":                {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/report"},
	"internal/store":                 {"shared-contract-go/pkg/fs"},
}

// hostLibraries are the library leaves that touch the host. Only the
// composition root imports them, and no unit test does.
var hostLibraries = []string{
	"shared-config-go/pkg/osfiles", "shared-contract-go/pkg/fs/osfs", "shared-core-go/pkg/process/osexec", "shared-core-go/pkg/scratch/osscratch",
}

// concrete lists the packages that bind a port to a technology or name a
// variant, with the one package that may import each from non-test code.
var concrete = map[string]string{
	"internal/registry":               "cmd/agent-eval",
	"internal/harness/claude":         "internal/registry",
	"internal/harness/codex":          "internal/registry",
	"internal/predicate/filesystem":   "internal/registry",
	"internal/predicate/llm":          "internal/registry",
	"internal/predicate/session":      "internal/registry",
	"internal/protocol/trigger":       "internal/registry",
	"internal/protocol/routing":       "internal/registry",
	"internal/protocol/output":        "internal/registry",
	"internal/protocol/instruction":   "internal/registry",
	"internal/protocol/orchestration": "internal/registry",
}

// doubles are the test-double packages. They are for tests only: no non-test
// package outside the table imports them, and they may name leaves.
var doubles = map[string]bool{
	"internal/protocol/protocoltest": true,
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
	"cmd/agent-eval": true,
}

// clockRoots may read the clock: the composition root hands the instant to the
// rest of the tool.
var clockRoots = map[string]bool{"cmd/agent-eval": true}

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

// TestLibraryImportsFollowTheTable keeps each edge into a shared library a
// decision of the libraries table.
func TestLibraryImportsFollowTheTable(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, library)
			if inside && !contains(libraries[f.dir], rel) {
				t.Errorf("%s imports %s; %s may import only the libraries %v", f.path, rel, f.dir, libraries[f.dir])
			}
		}
	}
}

// TestNoUnitTestImportsAHostLibrary keeps every unit test off the real host.
func TestNoUnitTestImportsAHostLibrary(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, library)
			if f.test && inside && contains(hostLibraries, rel) {
				t.Errorf("%s imports %s; a unit test stays off the host", f.path, rel)
			}
		}
	}
}

// TestConcretePackagesHaveOneImporter keeps a variant behind its port: the
// registry is the one file that names the concrete leaves, and the main package
// the one that names the host adapters and the registry.
func TestConcretePackagesHaveOneImporter(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			owner, isConcrete := concrete[rel]
			if inside && isConcrete && f.dir != owner && !doubles[f.dir] {
				t.Errorf("%s imports %s; only %q may", f.path, rel, owner)
			}
		}
	}
}

// TestOnlyDeclaredPackagesImportThirdParty keeps the core on the standard
// library and the shared libraries.
func TestOnlyDeclaredPackagesImportThirdParty(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.imports {
			if isStdlib(imp) || strings.HasPrefix(imp, module) || strings.HasPrefix(imp, library) {
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

// TestNoPackageImportsATestDouble keeps the doubles out of shipped code.
func TestNoPackageImportsATestDouble(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && doubles[rel] {
				t.Errorf("%s imports the test double %s", f.path, rel)
			}
		}
	}
}

// TestTheRegistryIsOneFile pins the one-file rule: adding a variant is one line
// in one file.
func TestTheRegistryIsOneFile(t *testing.T) {
	var registry []string
	for _, f := range nonTest(moduleFiles(t)) {
		if f.dir == "internal/registry" {
			registry = append(registry, filepath.Base(f.path))
		}
	}
	if len(registry) != 1 {
		t.Errorf("internal/registry holds %v; it must hold exactly one file", registry)
	}
}

// TestLeavesNeverImportASibling keeps the variants of one axis independent.
func TestLeavesNeverImportASibling(t *testing.T) {
	for _, f := range nonTest(moduleFiles(t)) {
		axis := filepath.Dir(f.dir)
		for _, imp := range f.imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && rel != f.dir && filepath.Dir(rel) == axis && filepath.Base(axis) != "internal" && !doubles[rel] {
				t.Errorf("%s imports its sibling %s", f.path, rel)
			}
		}
	}
}
