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

const module = "github.com/xonovex/platform/packages/tooling/cli/agent-eval-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	// The composition root binds every adapter and the registries.
	"cmd/agent-eval": {
		"internal/cmd", "internal/execute", "internal/harness", "internal/model", "internal/predicate", "internal/registry",
	},

	// The driving adapter reaches the core and the ports, never an adapter.
	"internal/cmd": {
		"internal/baseline", "internal/budget", "internal/catalog", "internal/cost", "internal/document",
		"internal/harness", "internal/migrate", "internal/model", "internal/predicate", "internal/protocol",
		"internal/report", "internal/run", "internal/settings", "internal/store",
	},

	// The one file that names the concrete harnesses, protocols, predicates and
	// confinements.
	"internal/registry": {
		"internal/confine", "internal/confine/launch", "internal/confine/none",
		"internal/harness", "internal/harness/claude", "internal/harness/codex", "internal/judge", "internal/model",
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
	"internal/execute":  {"internal/confine", "internal/harness"},
	"internal/judge":    {"internal/harness", "internal/model"},

	// Ports. A port holds the contract and the test double holds a fake.
	"internal/confine":   {},
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
	"internal/confine/launch":         {"internal/confine", "internal/model"},
	"internal/confine/none":           {"internal/confine"},
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
		"shared-core-go/pkg/logging", "shared-core-go/pkg/process/osexec", "shared-core-go/pkg/scratch/osscratch",
		"shared-launch-go/pkg/oshost",
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
	"internal/confine": {"shared-core-go/pkg/process"},
	"internal/confine/launch": {
		"shared-agent-go/pkg/agents", "shared-agent-go/pkg/egress", "shared-agent-go/pkg/isolation",
		"shared-agent-go/pkg/network", "shared-agent-go/pkg/policy", "shared-agent-go/pkg/provision",
		"shared-agent-go/pkg/types", "shared-core-go/pkg/envutil", "shared-core-go/pkg/process",
		"shared-launch-go/pkg/credential", "shared-launch-go/pkg/isolation", "shared-launch-go/pkg/launch",
		"shared-launch-go/pkg/sandbox",
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
	"internal/registry": {
		"shared-agent-go/pkg/agents", "shared-agent-go/pkg/isolation", "shared-agent-go/pkg/provision",
		"shared-agent-go/pkg/types", "shared-core-go/pkg/logging",
		"shared-launch-go/pkg/credential", "shared-launch-go/pkg/credential/env", "shared-launch-go/pkg/credential/file",
		"shared-launch-go/pkg/credential/keychain", "shared-launch-go/pkg/credential/pass",
		"shared-launch-go/pkg/credential/secretservice", "shared-launch-go/pkg/egress/managed",
		"shared-launch-go/pkg/isolation", "shared-launch-go/pkg/isolation/bwrap", "shared-launch-go/pkg/isolation/docker",
		"shared-launch-go/pkg/isolation/none", "shared-launch-go/pkg/isolation/seatbelt",
		"shared-launch-go/pkg/provision", "shared-launch-go/pkg/provision/command", "shared-launch-go/pkg/provision/nix",
		"shared-launch-go/pkg/provision/none", "shared-launch-go/pkg/sandbox",
	},
	"internal/report": {"shared-contract-go/pkg/findings", "shared-contract-go/pkg/report"},
	"internal/store":  {"shared-contract-go/pkg/fs"},
}

// hostLibraries are the library leaves that touch the host. Only the
// composition root imports them, and no unit test does.
var hostLibraries = []string{
	"shared-config-go/pkg/osfiles", "shared-contract-go/pkg/fs/osfs", "shared-core-go/pkg/process/osexec", "shared-core-go/pkg/scratch/osscratch",
	"shared-launch-go/pkg/oshost",
}

// concrete lists the packages that bind a port to a technology or name a
// variant, with the one package that may import each from non-test code.
var concrete = map[string]string{
	"internal/registry":               "cmd/agent-eval",
	"internal/confine/launch":         "internal/registry",
	"internal/confine/none":           "internal/registry",
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
	"cmd/agent-eval": true,
}

// clockRoots may read the clock: the composition root hands the instant to the
// rest of the tool.
var clockRoots = map[string]bool{"cmd/agent-eval": true}

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

// TestConcretePackagesHaveOneImporter keeps a variant behind its port: the
// registry is the one file that names the concrete leaves, and the main package
// the one that names the host adapters and the registry.
func TestConcretePackagesHaveOneImporter(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			owner, isConcrete := concrete[rel]
			if inside && isConcrete && f.Dir != owner && !doubles[f.Dir] {
				t.Errorf("%s imports %s; only %q may", f.Path, rel, owner)
			}
		}
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

// TestNoPackageImportsATestDouble keeps the doubles out of shipped code.
func TestNoPackageImportsATestDouble(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && doubles[rel] {
				t.Errorf("%s imports the test double %s", f.Path, rel)
			}
		}
	}
}

// TestTheRegistryIsOneFile pins the one-file rule: adding a variant is one line
// in one file.
func TestTheRegistryIsOneFile(t *testing.T) {
	var registry []string
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == "internal/registry" {
			registry = append(registry, filepath.Base(f.Path))
		}
	}
	if len(registry) != 1 {
		t.Errorf("internal/registry holds %v; it must hold exactly one file", registry)
	}
}

// TestLeavesNeverImportASibling keeps the variants of one axis independent.
func TestLeavesNeverImportASibling(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		axis := filepath.Dir(f.Dir)
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && rel != f.Dir && filepath.Dir(rel) == axis && filepath.Base(axis) != "internal" && !doubles[rel] {
				t.Errorf("%s imports its sibling %s", f.Path, rel)
			}
		}
	}
}

// launcher is the shared library of the isolators of agent-cli.
const launcher = library + "shared-launch-go/"

// launchPorts are the packages of the launcher that name no variant: the
// ports, the registry and the planner.
var launchPorts = map[string]bool{
	"shared-launch-go/pkg/credential": true, "shared-launch-go/pkg/isolation": true, "shared-launch-go/pkg/launch": true,
	"shared-launch-go/pkg/provision": true, "shared-launch-go/pkg/sandbox": true,
}

// TestOnlyTheConfinementReachesTheLauncher keeps the isolators of agent-cli
// behind the confine port: the launch leaf plans through the ports of the
// launcher, the registry names its variants, and the composition root builds
// its host. The core and every other leaf never see it.
func TestOnlyTheConfinementReachesTheLauncher(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, imp := range f.Imports {
			if !strings.HasPrefix(imp, launcher) {
				continue
			}
			rel := strings.TrimPrefix(imp, library)
			switch f.Dir {
			case "internal/registry":
			case "internal/confine/launch":
				if !launchPorts[rel] {
					t.Errorf("%s imports %s; the launch leaf reaches only the ports of the launcher, never a variant", f.Path, rel)
				}
			case "cmd/agent-eval":
				if rel != "shared-launch-go/pkg/oshost" {
					t.Errorf("%s imports %s; the composition root reaches only the host of the launcher", f.Path, rel)
				}
			default:
				t.Errorf("%s imports %s; only internal/confine/launch, internal/registry and cmd/agent-eval reach the launcher", f.Path, rel)
			}
		}
	}
}

// TestOnlyTheCompositionRootPrints keeps the standard streams in the main
// package: every other package writes to the io.Writer it is given.
func TestOnlyTheCompositionRootPrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == "cmd/agent-eval" {
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
