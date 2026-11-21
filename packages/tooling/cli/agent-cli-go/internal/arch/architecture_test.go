//go:build arch

// Package arch holds the architecture fitness tests of agent-cli. They read the
// source tree, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run. Each test locks one structural invariant:
//
//   - TestLeavesDoNotImportCompositionRoot: leaves never import the registry
//     package internal/sandbox/plugins.
//   - TestLeavesDoNotImportSiblings: no leaf imports a sibling leaf in the same
//     axis; cross-leaf knowledge lives in shared/.
//   - TestSharedCoresImportNoLeaf: an axis shared/ package imports no leaf of
//     any axis; bridges flow leaf -> other-axis/shared only.
//   - TestConcreteLeavesHaveOneImporter: only the registry packages and the
//     composition root import a concrete leaf of this module or of
//     shared-launch-go.
//   - TestLaunchLibraryImportsFollowTheTable: each package imports only the
//     shared-launch-go packages its entry lists.
//   - TestNoInternalPackageTouchesTheOperatingSystem: no internal package
//     imports os, os/exec, runtime or syscall; the composition root hands each
//     the OS adapter of shared-launch-go.
//   - TestInternalPackagesDoNotPrint: output goes through a logger or writer.
//   - TestNoHarnessIsNamedOutsideTheAgentRegistry: the planner and the
//     adapters compare the facts a harness declares, never its name.
//   - TestApplicationCoreNamesNoAdapter: the launch core and the cobra adapter
//     name no concrete plugin, no OS adapter, no origin leaf and no overlay
//     adapter.
//   - TestCommandLayerUsesOnlyItsEntryPoints: the cobra adapter reaches the
//     application through the launch and plugin entry points, and names of
//     any other package only what it prints or parses.
//   - TestCommandLayerReachesNoPort: the cobra adapter selects no port of the
//     dependencies it hands to the core; it prints to their Out only.
//
// origin_test.go locks the origin axis (internal/origin), and shared-launch-go
// holds the tests of its isolation, provision, egress and credential leaves,
// the selection layer, the ports and the run spec.
package arch

import (
	"go/ast"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"
)

const modPath = "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go"

// axisNames are the orthogonal axes, each a directory of a shared core and leaves.
var axisNames = []string{"workspace", "terminal"}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(here)))
}

// moduleFiles parses the non-test source of cmd and internal, the packages the
// module builds.
func moduleFiles(t *testing.T) []archtest.File {
	t.Helper()
	files, err := archtest.Load(os.DirFS(moduleRoot(t)), "cmd", "internal")
	if err != nil {
		t.Fatal(err)
	}
	return archtest.Production(files)
}

// filesIn returns the files of the package at dir, without its subpackages.
func filesIn(files []archtest.File, dir string) []archtest.File {
	var out []archtest.File
	for _, f := range files {
		if f.Dir == dir {
			out = append(out, f)
		}
	}
	return out
}

// filesUnder returns the files of the package at dir and of every package
// below it.
func filesUnder(files []archtest.File, dir string) []archtest.File {
	var out []archtest.File
	for _, f := range files {
		if f.Dir == dir || strings.HasPrefix(f.Dir, dir+"/") {
			out = append(out, f)
		}
	}
	return out
}

// dirImports returns the import paths of the package at dir.
func dirImports(files []archtest.File, dir string) []string {
	var imports []string
	for _, f := range filesIn(files, dir) {
		imports = append(imports, f.Imports...)
	}
	return imports
}

// leafDirs returns the directories of the leaves of an axis, keyed by name.
func leafDirs(files []archtest.File, axisDir string) map[string]string {
	out := map[string]string{}
	for _, f := range filesUnder(files, axisDir) {
		if rest, ok := strings.CutPrefix(f.Dir, axisDir+"/"); ok {
			leaf := topSegment(rest)
			out[leaf] = axisDir + "/" + leaf
		}
	}
	return out
}

// topSegment returns the first path segment of s (before any '/').
func topSegment(s string) string {
	if i := strings.IndexByte(s, '/'); i >= 0 {
		return s[:i]
	}
	return s
}

func TestLeavesDoNotImportCompositionRoot(t *testing.T) {
	files := moduleFiles(t)
	sandboxPath := modPath + "/internal/sandbox"
	for _, axis := range axisNames {
		for _, dir := range leafDirs(files, "internal/"+axis) {
			for _, imp := range dirImports(files, dir) {
				if imp == sandboxPath || strings.HasPrefix(imp, sandboxPath+"/") {
					t.Errorf("%s imports the registry package %s", dir, imp)
				}
			}
		}
	}
}

// concrete names, for each leaf of this module and of shared-launch-go, the
// only packages that may import it from non-test code: a registry package or
// the composition root. A leaf of shared-launch-go is keyed by its import path.
var concrete = map[string][]string{
	launchLibrary + "/pkg/isolation/none":            {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/isolation/bwrap":           {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/isolation/docker":          {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/isolation/seatbelt":        {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/isolation/dockersandboxes": {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/isolation/openshell":       {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/provision/none":            {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/provision/command":         {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/provision/nix":             {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/credential/env":            {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/credential/file":           {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/credential/keychain":       {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/credential/pass":           {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/credential/secretservice":  {"internal/sandbox/plugins"},
	launchLibrary + "/pkg/egress/managed":            {"cmd/agent-cli"},
	launchLibrary + "/pkg/oshost":                    {"cmd/agent-cli"},
	"internal/origin/directory":                      {"internal/origin/plugins"},
	"internal/origin/git":                            {"internal/origin/plugins"},
	"internal/terminal/tmux":                         {"cmd/agent-cli"},
	"internal/terminal/herdrtui":                     {"cmd/agent-cli"},
	"internal/terminal/herdrcore":                    {"cmd/agent-cli"},
	"internal/workspace/git":                         {"cmd/agent-cli"},
	"internal/workspace/jj":                          {"cmd/agent-cli"},
	"internal/origin/plugins":                        {"cmd/agent-cli"},
	"internal/overlay":                               {"cmd/agent-cli"},
	"internal/sandbox/plugins":                       {"cmd/agent-cli"},
}

// packageImports returns the non-test imports of every package of the module,
// keyed by the directory of the package relative to the module root.
func packageImports(t *testing.T) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, f := range moduleFiles(t) {
		out[f.Dir] = append(out[f.Dir], f.Imports...)
	}
	return out
}

// concreteKey returns the key of imp in the concrete table: the module-relative
// directory of a package of this module, or the import path of any other.
func concreteKey(imp string) string {
	if rel, inside := strings.CutPrefix(imp, modPath+"/"); inside {
		return rel
	}
	return imp
}

func TestConcreteLeavesHaveOneImporter(t *testing.T) {
	imports := packageImports(t)
	for leaf, importers := range concrete {
		if strings.HasPrefix(leaf, launchLibrary+"/") {
			// A library leaf has no directory here; its registry must reach it.
			for _, importer := range importers {
				if !slices.ContainsFunc(imports[importer], func(imp string) bool { return imp == leaf }) {
					t.Errorf("the concrete table names %s, which %s does not import", leaf, importer)
				}
			}
			continue
		}
		if _, ok := imports[leaf]; !ok {
			t.Errorf("the concrete table names %s, which has no Go files", leaf)
		}
	}
	for dir, paths := range imports {
		for _, imp := range paths {
			if importers, isLeaf := concrete[concreteKey(imp)]; isLeaf && !slices.Contains(importers, dir) {
				t.Errorf("%s imports the concrete leaf %s; only %v may", dir, imp, importers)
			}
		}
	}
}

// launchLibrary is the module path of the shared run kernel.
const launchLibrary = "github.com/xonovex/platform/packages/library/shared-launch-go"

// launchImports lists, for each package of this module, the packages of
// shared-launch-go it may import from non-test code. A package missing here
// imports none.
var launchImports = map[string][]string{
	"cmd/agent-cli":   {"pkg/egress/managed", "pkg/oshost", "pkg/record"},
	"internal/cmd":    {"pkg/record"},
	"internal/launch": {"pkg/credential", "pkg/isolation", "pkg/launch", "pkg/provision", "pkg/record", "pkg/sandbox"},
	"internal/sandbox/plugins": {
		"pkg/credential", "pkg/isolation", "pkg/provision", "pkg/sandbox",
		"pkg/isolation/none", "pkg/isolation/bwrap", "pkg/isolation/docker", "pkg/isolation/seatbelt",
		"pkg/isolation/dockersandboxes", "pkg/isolation/openshell",
		"pkg/provision/none", "pkg/provision/command", "pkg/provision/nix",
		"pkg/credential/env", "pkg/credential/file", "pkg/credential/keychain", "pkg/credential/pass",
		"pkg/credential/secretservice",
	},
}

func TestLaunchLibraryImportsFollowTheTable(t *testing.T) {
	imports := packageImports(t)
	for dir := range launchImports {
		if _, ok := imports[dir]; !ok {
			t.Errorf("the launch table names %s, which has no Go files", dir)
		}
	}
	for dir, paths := range imports {
		for _, imp := range paths {
			if pkg, inside := strings.CutPrefix(imp, launchLibrary+"/"); inside && !slices.Contains(launchImports[dir], pkg) {
				t.Errorf("%s imports %s of shared-launch-go; it may import only %v", dir, pkg, launchImports[dir])
			}
		}
	}
}

func TestLeavesDoNotImportSiblings(t *testing.T) {
	files := moduleFiles(t)
	for _, axis := range axisNames {
		base := modPath + "/internal/" + axis + "/"
		for leaf, dir := range leafDirs(files, "internal/"+axis) {
			for _, imp := range dirImports(files, dir) {
				sib, ok := strings.CutPrefix(imp, base)
				if !ok {
					continue
				}
				sibTop := topSegment(sib)
				if sibTop != "shared" && leaf != "shared" && sibTop != leaf {
					t.Errorf("%s reaches sibling leaf %s", dir, imp)
				}
			}
		}
	}
}

func TestSharedCoresImportNoLeaf(t *testing.T) {
	files := moduleFiles(t)
	for _, axis := range axisNames {
		dir := "internal/" + axis + "/shared"
		for _, imp := range dirImports(files, dir) {
			for _, other := range axisNames {
				base := modPath + "/internal/" + other + "/"
				if leaf, ok := strings.CutPrefix(imp, base); ok && topSegment(leaf) != "shared" {
					t.Errorf("shared core %s imports leaf %s", dir, imp)
				}
			}
		}
	}
}

func TestNoInternalPackageTouchesTheOperatingSystem(t *testing.T) {
	banned := map[string]bool{"os": true, "os/exec": true, "runtime": true, "syscall": true}
	for _, f := range filesUnder(moduleFiles(t), "internal") {
		for _, imp := range f.Imports {
			if banned[imp] {
				t.Errorf("%s imports %s; reach the machine through a port", f.Path, imp)
			}
		}
	}
}

func TestInternalPackagesDoNotPrint(t *testing.T) {
	for _, f := range filesUnder(moduleFiles(t), "internal") {
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s calls %s", f.Path, use)
		}
	}
}

func TestApplicationCoreNamesNoAdapter(t *testing.T) {
	files := moduleFiles(t)
	adapters := []string{
		launchLibrary + "/pkg/oshost",
		launchLibrary + "/pkg/hostlaunch",
		launchLibrary + "/pkg/isolation/none",
		launchLibrary + "/pkg/isolation/bwrap",
		launchLibrary + "/pkg/isolation/docker",
		launchLibrary + "/pkg/isolation/seatbelt",
		launchLibrary + "/pkg/isolation/dockersandboxes",
		launchLibrary + "/pkg/isolation/openshell",
		launchLibrary + "/pkg/egress/managed",
		launchLibrary + "/pkg/credential/env",
		launchLibrary + "/pkg/credential/file",
		launchLibrary + "/pkg/credential/keychain",
		launchLibrary + "/pkg/credential/secretservice",
		launchLibrary + "/pkg/credential/pass",
		launchLibrary + "/pkg/provision/none",
		launchLibrary + "/pkg/provision/command",
		launchLibrary + "/pkg/provision/nix",
		modPath + "/internal/sandbox/plugins",
		modPath + "/internal/workspace/git",
		modPath + "/internal/workspace/jj",
		modPath + "/internal/terminal/tmux",
		modPath + "/internal/origin/directory",
		modPath + "/internal/origin/git",
		modPath + "/internal/origin/plugins",
		modPath + "/internal/overlay",
	}
	for pkg, extra := range map[string][]string{
		"launch": {"github.com/spf13/cobra", modPath + "/internal/cmd", modPath + "/internal/plugin"},
		"cmd":    nil,
	} {
		banned := append(append([]string{}, adapters...), extra...)
		for _, imp := range dirImports(files, "internal/"+pkg) {
			for _, b := range banned {
				if imp == b || strings.HasPrefix(imp, b+"/") {
					t.Errorf("internal/%s imports adapter %s", pkg, imp)
				}
			}
		}
	}
}

// agentTypeSurface lists the identifiers of the shared agent types package that
// name no concrete agent.
var agentTypeSurface = map[string]bool{"AgentType": true, "AgentConfig": true, "AgentExecOptions": true}

func TestNoHarnessIsNamedOutsideTheAgentRegistry(t *testing.T) {
	for _, f := range moduleFiles(t) {
		ast.Inspect(f.Syntax, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if ok && pkg.Name == "types" && strings.HasPrefix(sel.Sel.Name, "Agent") && !agentTypeSurface[sel.Sel.Name] {
				t.Errorf("%s names the concrete agent %s; compare the declared facts instead", f.Path, sel.Sel.Name)
			}
			return true
		})
	}
}

// commandModule is the import path prefix of the packages of this module.
const commandModule = modPath + "/internal/"

// commandImports lists, for each non-standard package internal/cmd may import,
// the identifiers it may name of it; nil allows every identifier. The cobra
// layer is a driving adapter: a rule that decides (a default, a selection, a
// path, an approval) belongs to the application core, which the adapter reaches
// through launch and plugin. Any other package is allowed only for what the
// adapter prints or parses, so naming a deciding identifier of it, such as
// trust.Decide or a runconfig function, fails here.
var commandImports = map[string]map[string]bool{
	"github.com/spf13/cobra": nil,
	commandModule + "launch": nil,
	commandModule + "plugin": nil,
	// The name of the approval flag is shared by every command that approves.
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust": {"ApproveFlag": true},
	// The sources of `config show --sources` are rendered by the library.
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/explain": {"Render": true},
	// The runs commands print a record and its JSON form.
	launchLibrary + "/pkg/record": {"Record": true, "Marshal": true},
	// The schema command prints the contract documents verbatim, by name.
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/schemas": nil,
}

func TestCommandLayerUsesOnlyItsEntryPoints(t *testing.T) {
	for _, f := range filesUnder(moduleFiles(t), "internal/cmd") {
		names := map[string]string{}
		for _, spec := range f.Syntax.Imports {
			imported, _ := strconv.Unquote(spec.Path.Value)
			if archtest.IsStandard(imported) {
				continue
			}
			if _, ok := commandImports[imported]; !ok {
				t.Errorf("%s imports %s; reach the application through launch or plugin", f.Path, imported)
				continue
			}
			name := path.Base(imported)
			if spec.Name != nil {
				name = spec.Name.Name
			}
			names[name] = imported
		}
		ast.Inspect(f.Syntax, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			imported, named := names[pkg.Name]
			if allowed := commandImports[imported]; named && allowed != nil && !allowed[sel.Sel.Name] {
				t.Errorf("%s names %s.%s; that rule belongs to the application core", f.Path, pkg.Name, sel.Sel.Name)
			}
			return true
		})
	}
}

// commandDependencies are the types that carry the ports the cobra adapter
// hands to the core, by package name and type name.
var commandDependencies = map[string]string{"launch": "Dependencies", "plugin": "Manager"}

// dependencyNames returns the names that f declares, as a parameter, a result,
// a struct field or a variable, with a type of commandDependencies.
func dependencyNames(f *ast.File) map[string]bool {
	names := map[string]bool{}
	isDependencies := func(expr ast.Expr) bool {
		sel, ok := expr.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := sel.X.(*ast.Ident)
		return ok && commandDependencies[pkg.Name] == sel.Sel.Name
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Field:
			if isDependencies(node.Type) {
				for _, name := range node.Names {
					names[name.Name] = true
				}
			}
		case *ast.ValueSpec:
			if isDependencies(node.Type) {
				for _, name := range node.Names {
					names[name.Name] = true
				}
			}
		}
		return true
	})
	return names
}

// portFields returns the fields of every type of commandDependencies except
// Out, read from the source of its package.
func portFields(t *testing.T, files []archtest.File) map[string]bool {
	t.Helper()
	fields := map[string]bool{}
	for pkg, typeName := range commandDependencies {
		found := false
		for _, f := range filesUnder(files, "internal/"+pkg) {
			ast.Inspect(f.Syntax, func(n ast.Node) bool {
				spec, ok := n.(*ast.TypeSpec)
				if !ok || spec.Name.Name != typeName {
					return true
				}
				found = true
				for _, field := range spec.Type.(*ast.StructType).Fields.List {
					for _, name := range field.Names {
						fields[name.Name] = name.Name != "Out"
					}
				}
				return false
			})
		}
		if !found {
			t.Errorf("package %s declares no type %s", pkg, typeName)
		}
	}
	return fields
}

// TestCommandLayerReachesNoPort lets a command call the methods of the values
// that carry the ports and print to their Out, and select no other field. A
// command that selects a port, such as the host, a registry or a store, reads
// the machine or a plugin set itself and so decides what the core decides.
func TestCommandLayerReachesNoPort(t *testing.T) {
	files := moduleFiles(t)
	ports := portFields(t, files)
	for _, f := range filesUnder(files, "internal/cmd") {
		carriers := dependencyNames(f.Syntax)
		ast.Inspect(f.Syntax, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if x, ok := sel.X.(*ast.Ident); ok && carriers[x.Name] && ports[sel.Sel.Name] {
				t.Errorf("%s selects %s.%s; pass the dependencies to the core", f.Path, x.Name, sel.Sel.Name)
			}
			return true
		})
	}
}
