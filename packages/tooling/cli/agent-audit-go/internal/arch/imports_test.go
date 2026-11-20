//go:build arch

// Package arch holds the architecture fitness tests of the tool. They read the
// source tree, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run.
package arch

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"
)

const module = "github.com/xonovex/platform/packages/tooling/cli/agent-audit-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	// The composition root binds every adapter.
	"cmd/agent-audit": {"internal/cmd", "internal/transport/plugins", "internal/web/nethttp"},

	// The driving adapter reaches the core and the ports, never an adapter.
	"internal/cmd": {
		"internal/audit", "internal/catalogs", "internal/compare", "internal/compare/gitbase",
		"internal/legacy", "internal/observe", "internal/render",
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
	"internal/web":         {},
	"internal/web/webtest": {"internal/web"},

	// Adapters: one port each, plus the registry that names them.
	"internal/compare/gitbase":      {"internal/gitfs"},
	"internal/gitfs":                {},
	"internal/web/nethttp":          {"internal/web"},
	"internal/transport/filesystem": {"internal/observe"},
	"internal/transport/git":        {"internal/observe"},
	"internal/transport/http":       {"internal/observe", "internal/web"},
	"internal/transport/plugins": {
		"internal/observe", "internal/transport/filesystem",
		"internal/transport/git", "internal/transport/http", "internal/web",
	},
}

// library is the path prefix of the shared Go libraries.
const library = "github.com/xonovex/platform/packages/library/"

// libraries lists, for each package, the library packages it may import from
// its non-test files. A package missing from the table imports no library, so a
// new edge into a library is a decision in this file.
var libraries = map[string][]string{
	"cmd/agent-audit": {
		"shared-config-go/pkg/osfiles",
		"shared-contract-go/pkg/findings", "shared-contract-go/pkg/fs/osfs", "shared-contract-go/pkg/wiring",
		"shared-core-go/pkg/process/osexec",
	},
	"internal/audit":    {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/fs", "shared-contract-go/pkg/sources"},
	"internal/catalogs": {"shared-contract-go/pkg/discovery", "shared-contract-go/pkg/fs", "shared-contract-go/pkg/sources"},
	"internal/cmd": {
		"shared-config-go/pkg/discover", "shared-config-go/pkg/explain", "shared-config-go/pkg/merge",
		"shared-config-go/pkg/resolve", "shared-config-go/pkg/section/audit", "shared-config-go/pkg/spec",
		"shared-contract-go/pkg/findings", "shared-contract-go/pkg/fs", "shared-contract-go/pkg/jsonschema",
		"shared-contract-go/pkg/report", "shared-contract-go/pkg/schemas", "shared-contract-go/pkg/sources",
		"shared-core-go/pkg/process",
	},
	"internal/compare": {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/sources"},
	"internal/compare/gitbase": {
		"shared-contract-go/pkg/findings", "shared-contract-go/pkg/fs", "shared-contract-go/pkg/sources",
		"shared-core-go/pkg/process",
	},
	"internal/gitfs":                {"shared-contract-go/pkg/fs", "shared-core-go/pkg/process"},
	"internal/legacy":               {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/schemas", "shared-contract-go/pkg/sources"},
	"internal/observe":              {"shared-contract-go/pkg/sources"},
	"internal/render":               {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/report", "shared-contract-go/pkg/sarif"},
	"internal/rules":                {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/sources"},
	"internal/stamp":                {"shared-contract-go/pkg/fs", "shared-contract-go/pkg/sources"},
	"internal/transport/filesystem": {"shared-contract-go/pkg/fs", "shared-contract-go/pkg/sources"},
	"internal/transport/git":        {"shared-contract-go/pkg/fs", "shared-contract-go/pkg/sources", "shared-core-go/pkg/process"},
	"internal/transport/http":       {"shared-contract-go/pkg/sources"},
	"internal/transport/plugins":    {"shared-contract-go/pkg/fs", "shared-core-go/pkg/process"},
}

// hostLibraries are the library leaves that touch the host. Only the
// composition root imports them, and no unit test does.
var hostLibraries = []string{"shared-config-go/pkg/osfiles", "shared-contract-go/pkg/fs/osfs", "shared-core-go/pkg/process/osexec"}

// concrete lists the packages that bind a port to a technology or name a
// concrete transport, with the one package that may import each from non-test
// code. The test doubles are for tests only.
var concrete = map[string]string{
	"internal/web/nethttp":          "cmd/agent-audit",
	"internal/transport/filesystem": "internal/transport/plugins",
	"internal/transport/git":        "internal/transport/plugins",
	"internal/transport/http":       "internal/transport/plugins",
	"internal/transport/plugins":    "cmd/agent-audit",
	"internal/web/webtest":          "",
}

// thirdParty lists the packages that may import a module outside the standard
// library and this repository, with the module prefix each may use.
var thirdParty = map[string]string{
	"internal/cmd": "github.com/spf13/cobra",
}

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
	files, err := archtest.Load(os.DirFS(moduleRoot(t)), "cmd", "internal")
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

// hostAdapters may import hostImports from non-test code.
var hostAdapters = map[string]bool{
	"cmd/agent-audit":      true,
	"internal/web/nethttp": true,
}

// clockRoots may read the clock: the composition root hands the instant to the
// rest of the tool.
var clockRoots = map[string]bool{"cmd/agent-audit": true}

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

// TestInternalImportsFollowTheTable keeps the dependency direction of the
// hexagon: the core knows the ports, an adapter knows one port, and only the
// composition root knows them all.
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

// TestLibraryImportsFollowTheTable keeps each edge into a shared library a
// decision of the libraries table.
func TestLibraryImportsFollowTheTable(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, library)
			if inside && !slices.Contains(libraries[f.Dir], rel) {
				t.Errorf("%s imports %s; %s may import only the libraries %v", f.Path, rel, f.Dir, libraries[f.Dir])
			}
		}
	}
}

// TestNoUnitTestImportsAHostLibrary keeps every unit test off the real host.
func TestNoUnitTestImportsAHostLibrary(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, library)
			if f.Test && inside && slices.Contains(hostLibraries, rel) {
				t.Errorf("%s imports %s; a unit test stays off the host", f.Path, rel)
			}
		}
	}
}

// TestConcretePackagesHaveOneImporter keeps a technology behind its port: the
// registry is the one file that names the concrete transports, and the main
// package the one that names the adapters and the registry.
func TestConcretePackagesHaveOneImporter(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			owner, isConcrete := concrete[rel]
			if inside && isConcrete && f.Dir != owner {
				t.Errorf("%s imports %s; only %q may", f.Path, rel, owner)
			}
		}
	}
}

// TestTheRegistryIsTheOnlyFileThatNamesTransports pins the one-file rule: the
// registry package holds a single source file.
func TestTheRegistryIsTheOnlyFileThatNamesTransports(t *testing.T) {
	var registry []string
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == "internal/transport/plugins" {
			registry = append(registry, filepath.Base(f.Path))
		}
	}
	if len(registry) != 1 {
		t.Errorf("internal/transport/plugins holds %v; it must hold exactly one file", registry)
	}
}

// TestOnlyDeclaredPackagesImportThirdParty keeps the core on the standard
// library and the shared libraries.
func TestOnlyDeclaredPackagesImportThirdParty(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if archtest.IsStandard(imp) || strings.HasPrefix(imp, module) || strings.HasPrefix(imp, library) {
				continue
			}
			want, ok := thirdParty[f.Dir]
			if !ok || !strings.HasPrefix(imp, want) {
				t.Errorf("%s imports %s; %s may import no outside module except %q", f.Path, imp, f.Dir, want)
			}
		}
	}
}

// TestOnlyHostAdaptersTouchTheHost keeps the disk, environment, process and
// network out of the core and out of every unit test.
func TestOnlyHostAdaptersTouchTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if hostAdapters[f.Dir] && !f.Test {
			continue
		}
		for _, imp := range f.Imports {
			if archtest.IsHost(imp) {
				t.Errorf("%s imports %s; only the non-test files of %v may", f.Path, imp, keys(hostAdapters))
			}
		}
	}
}

// TestOnlyTheCompositionRootReadsTheClock keeps time an input: a function takes
// the instant it needs as an argument.
func TestOnlyTheCompositionRootReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if archtest.ReadsClock(f.Syntax) && (!clockRoots[f.Dir] || f.Test) {
			t.Errorf("%s reads the clock with time.Now, time.Since or time.Until; pass the instant in", f.Path)
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

// TestOnlyTheCompositionRootPrints keeps the standard streams in the main
// package: every other package writes to the io.Writer it is given.
func TestOnlyTheCompositionRootPrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == "cmd/agent-audit" {
			continue
		}
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; write to the io.Writer the command is given", f.Path, use)
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
