//go:build arch

// Package arch holds the architecture fitness tests of the validator. They read
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
	// The plugin checks read the manifest and the marketplace paths of each
	// harness from its facts.
	"internal/check/marketplace": {agentLibrary + "pkg/agents", agentLibrary + "pkg/types"},
	"internal/check/plugin":      {agentLibrary + "pkg/agents", agentLibrary + "pkg/types"},
	"internal/cmd": {
		"github.com/spf13/cobra",
		configLibrary + "pkg/discover", configLibrary + "pkg/explain", configLibrary + "pkg/merge",
		configLibrary + "pkg/resolve", configLibrary + "pkg/section/validate", configLibrary + "pkg/spec",
		configLibrary + "pkg/configtest",
	},
	// The composition root binds the configuration ports to the disk.
	"cmd/agent-validate": {configLibrary + "pkg/osfiles"},
}

// agentLibrary holds the facts of the harnesses.
const agentLibrary = "github.com/xonovex/platform/packages/library/shared-agent-go/"

// configLibrary reads the layered agentic.json. Only the command tree reads it,
// and only the composition root binds its host adapter.
const configLibrary = "github.com/xonovex/platform/packages/library/shared-config-go/"

// contractAdapters bind a contract port to a technology. The composition root
// binds the real ones; a test double binds the memory one.
var contractAdapters = map[string][]string{
	"pkg/fs/osfs":           {"cmd/agent-validate"},
	"pkg/wiring":            {"cmd/agent-validate", "internal/checktest"},
	"pkg/fs/memfs":          {"internal/checktest"},
	"pkg/jsonschema/tekuri": {},
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

// hostAdapters may import hostImports from their non-test files.
var hostAdapters = []string{"cmd/agent-validate"}

func isCheck(dir string) bool { return strings.HasPrefix(dir, checkPrefix) }

// TestEveryPackageIsDeclared fails for a package outside the checks that the
// table does not name, and for a table entry with no package.
func TestEveryPackageIsDeclared(t *testing.T) {
	present := map[string]bool{}
	for _, f := range moduleFiles(t) {
		present[f.Dir] = true
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
	for _, f := range archtest.Production(moduleFiles(t)) {
		permitted := core[f.Dir]
		if isCheck(f.Dir) {
			permitted = support
		}
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if !inside {
				continue
			}
			if f.Dir == "internal/registry" && isCheck(rel) {
				continue
			}
			if !slices.Contains(permitted, rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.Path, rel, f.Dir, permitted)
			}
		}
	}
}

// TestOnlyTheRegistryImportsAConcreteCheck keeps a check behind the port: no
// other package, test included except the check's own, names one.
func TestOnlyTheRegistryImportsAConcreteCheck(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if !inside || !isCheck(rel) || rel == "internal/check" {
				continue
			}
			ownTest := f.Test && f.Dir == rel
			if f.Dir != "internal/registry" && !ownTest {
				t.Errorf("%s imports the check %s; only internal/registry may", f.Path, rel)
			}
		}
	}
}

// TestContractAdaptersStayAtTheEdge keeps the real disk, the schema engine and
// the wiring out of the core.
func TestContractAdaptersStayAtTheEdge(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, contract)
			if !inside {
				continue
			}
			allowed, restricted := contractAdapters[rel]
			if !restricted {
				continue
			}
			if f.Test {
				// A test may bind the memory adapter and the wiring, to build a
				// context; it never binds the disk or the schema engine.
				if rel == "pkg/fs/osfs" || rel == "pkg/jsonschema/tekuri" {
					t.Errorf("%s imports %s; a unit test stays off the disk", f.Path, rel)
				}
				continue
			}
			if !slices.Contains(allowed, f.Dir) {
				t.Errorf("%s imports %s; only %v may", f.Path, rel, allowed)
			}
		}
	}
}

// TestOnlyDeclaredPackagesImportThirdParty keeps the core on the standard
// library and the contract.
func TestOnlyDeclaredPackagesImportThirdParty(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if archtest.IsStandard(imp) || strings.HasPrefix(imp, module) || strings.HasPrefix(imp, contract) {
				continue
			}
			ok := false
			for _, want := range thirdParty[f.Dir] {
				ok = ok || strings.HasPrefix(imp, want)
			}
			if !ok {
				t.Errorf("%s imports %s; %s may import only %v", f.Path, imp, f.Dir, thirdParty[f.Dir])
			}
		}
	}
}

// TestOnlyTheCompositionRootTouchesTheHost keeps the real disk, environment,
// process and network out of the core and out of every unit test.
func TestOnlyTheCompositionRootTouchesTheHost(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if slices.Contains(hostAdapters, f.Dir) && !f.Test {
			continue
		}
		for _, imp := range f.Imports {
			if archtest.IsHost(imp) {
				t.Errorf("%s imports %s; only the non-test files of %v may", f.Path, imp, hostAdapters)
			}
		}
	}
}

// TestNoPackageReadsTheClock keeps time an input.
func TestNoPackageReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if archtest.ReadsClock(f.Syntax) {
			t.Errorf("%s reads the clock; pass the instant in", f.Path)
		}
	}
}

// TestOnlyTheCompositionRootPrints keeps the standard streams in the main
// package: every other package writes to the io.Writer it is given.
func TestOnlyTheCompositionRootPrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == "cmd/agent-validate" {
			continue
		}
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; write to the io.Writer the command is given", f.Path, use)
		}
	}
}

// TestNoPackageKeepsState keeps each check and every other package
// independent: no package keeps state that a second run in one process would
// inherit.
func TestNoPackageKeepsState(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, state := range archtest.PackageState(f.Syntax) {
			t.Errorf("%s declares %s; pass state as a value", f.Path, state)
		}
	}
}
