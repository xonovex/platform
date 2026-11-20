package sandbox

// Architecture / import fitness tests. Each test locks one structural invariant of
// the orthogonal-axis layout so it cannot silently regress:
//
//   - TestLeavesDoNotImportCompositionRoot — hexagonal: leaves never import the
//     composition root (internal/sandbox); only internal/sandbox/plugins may know
//     the concrete leaves.
//   - TestLeavesDoNotImportSiblings — orthogonal bare leaves: no leaf imports a
//     sibling leaf in the same axis; cross-leaf knowledge lives in shared/.
//   - TestSharedCoresImportNoLeaf — one-way cross-axis glue: an axis shared/ core
//     imports no leaf of any axis; bridges flow leaf -> other-axis/shared only.
//   - TestSelectionLayerNamesNoVariant — microkernel capability gate: the Select /
//     policy layer names no concrete method constant; it gates on Capabilities.
//   - TestConcreteConstructorsOnlyInPlugins — sole composition root: the leaf
//     constructors are referenced only from plugins.go.
//   - TestRegistryFactoriesAreLazy — microkernel lazy binding: the registry stores
//     func()-typed factories, not eager instances.
//   - TestOnlySystemAdapterTouchesTheOperatingSystem — hexagonal edge: only
//     internal/system imports os, os/exec, runtime or syscall; every other
//     package reaches the machine through a port it declares.
//   - TestInternalPackagesDoNotPrint — explicit output: no package writes to the
//     process streams through fmt.Print*; output goes through a logger or writer.
//   - TestNoHarnessIsNamedOutsideTheAgentRegistry — facts, not names: no
//     production file of this module names a concrete agent type; the planner and
//     the isolators compare the facts a harness declares.
//   - TestApplicationCoreNamesNoAdapter — the launch core and the cobra adapter
//     name no concrete plugin, no OS adapter, and (for launch) no cobra.
//
// The checks use only the standard library (go/parser, go/ast) — no new dependency.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modPath = "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go"

// axisNames are the orthogonal axes, each a directory of a shared core and leaves.
var axisNames = []string{"isolation", "provision", "network", "workspace", "terminal"}

// internalRoot returns the absolute path of the internal/ directory.
func internalRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = .../internal/sandbox/architecture_test.go
	return filepath.Dir(filepath.Dir(file))
}

// leafDirs returns the immediate subdirectories of an axis dir, keyed by name.
func leafDirs(t *testing.T, axisDir string) map[string]string {
	entries, err := os.ReadDir(axisDir)
	if err != nil {
		t.Fatalf("read %s: %v", axisDir, err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() {
			out[e.Name()] = filepath.Join(axisDir, e.Name())
		}
	}
	return out
}

// dirImports returns the import paths of all non-test Go files in dir.
func dirImports(t *testing.T, dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	fset := token.NewFileSet()
	var imps []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, im := range f.Imports {
			if path, err := strconv.Unquote(im.Path.Value); err == nil {
				imps = append(imps, path)
			}
		}
	}
	return imps
}

// topSegment returns the first path segment of s (before any '/').
func topSegment(s string) string {
	if i := strings.IndexByte(s, '/'); i >= 0 {
		return s[:i]
	}
	return s
}

func TestLeavesDoNotImportCompositionRoot(t *testing.T) {
	root := internalRoot(t)
	sandboxPath := modPath + "/internal/sandbox"
	for _, axis := range axisNames {
		for _, dir := range leafDirs(t, filepath.Join(root, axis)) {
			for _, imp := range dirImports(t, dir) {
				if imp == sandboxPath || (strings.HasPrefix(imp, sandboxPath+"/") && imp != sandboxPath+"/plugins") {
					t.Errorf("%s imports composition root %s", dir, imp)
				}
			}
		}
	}
}

func TestLeavesDoNotImportSiblings(t *testing.T) {
	root := internalRoot(t)
	for _, axis := range axisNames {
		base := modPath + "/internal/" + axis + "/"
		for leaf, dir := range leafDirs(t, filepath.Join(root, axis)) {
			for _, imp := range dirImports(t, dir) {
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
	root := internalRoot(t)
	for _, axis := range axisNames {
		dir := filepath.Join(root, axis, "shared")
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		for _, imp := range dirImports(t, dir) {
			for _, other := range axisNames {
				base := modPath + "/internal/" + other + "/"
				if leaf, ok := strings.CutPrefix(imp, base); ok && topSegment(leaf) != "shared" {
					t.Errorf("shared core %s imports leaf %s", dir, imp)
				}
			}
		}
	}
}

func TestSelectionLayerNamesNoVariant(t *testing.T) {
	root := internalRoot(t)
	banned := map[string]bool{
		"IsolationNone": true, "IsolationBwrap": true, "IsolationDocker": true,
		"ProvisionNone": true, "ProvisionNix": true, "ProvisionCommand": true,
		"NetworkHost": true, "NetworkNone": true, "NetworkProxy": true,
		"ModeHost": true, "ModeNone": true, "ModeProxy": true,
	}
	// The selection / policy gate (registry.go: Select + Capabilities + EnforcePolicy)
	// must name no concrete variant; it gates on the Capabilities booleans only.
	file := filepath.Join(root, "sandbox", "registry.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && banned[id.Name] {
			t.Errorf("selection layer names concrete variant %q", id.Name)
		}
		return true
	})
}

func TestConcreteConstructorsOnlyInPlugins(t *testing.T) {
	root := internalRoot(t)
	// leaf import path -> its concrete constructor name.
	ctors := map[string]string{
		modPath + "/internal/isolation/none":    "NewIsolator",
		modPath + "/internal/isolation/bwrap":   "NewIsolator",
		modPath + "/internal/isolation/docker":  "NewIsolator",
		modPath + "/internal/provision/none":    "New",
		modPath + "/internal/provision/command": "New",
		modPath + "/internal/provision/nix":     "New",
	}
	pluginsFile := filepath.Join(root, "sandbox", "plugins", "plugins.go")

	fset := token.NewFileSet()
	for _, file := range productionFiles(t, root) {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		// alias -> import path for this file.
		alias := map[string]string{}
		for _, im := range f.Imports {
			path, err := strconv.Unquote(im.Path.Value)
			if err != nil {
				continue
			}
			name := topSegmentLast(path)
			if im.Name != nil {
				name = im.Name.Name
			}
			alias[name] = path
		}
		ast.Inspect(f, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			x, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			path, ok := alias[x.Name]
			if !ok {
				return true
			}
			if ctor, banned := ctors[path]; banned && sel.Sel.Name == ctor && file != pluginsFile {
				t.Errorf("concrete constructor %s.%s referenced outside plugins.go: %s", x.Name, ctor, file)
			}
			return true
		})
	}
}

// topSegmentLast returns the final path segment (the default import identifier).
func topSegmentLast(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func TestRegistryFactoriesAreLazy(t *testing.T) {
	if reflect.TypeOf(IsolatorFactory(nil)).Kind() != reflect.Func {
		t.Error("IsolatorFactory must be a func()-typed factory (lazy binding)")
	}
	if reflect.TypeOf(ProvisionerFactory(nil)).Kind() != reflect.Func {
		t.Error("ProvisionerFactory must be a func()-typed factory (lazy binding)")
	}
}

// productionFiles returns every non-test Go file under root.
func productionFiles(t *testing.T, root string) []string {
	var files []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk: %v", err)
	}
	return files
}

func TestOnlySystemAdapterTouchesTheOperatingSystem(t *testing.T) {
	root := internalRoot(t)
	systemDir := filepath.Join(root, "system") + string(filepath.Separator)
	banned := map[string]bool{"os": true, "os/exec": true, "runtime": true, "syscall": true}
	fset := token.NewFileSet()
	for _, file := range productionFiles(t, root) {
		if strings.HasPrefix(file, systemDir) {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, im := range f.Imports {
			if path, _ := strconv.Unquote(im.Path.Value); banned[path] {
				t.Errorf("%s imports %s; reach the machine through a port", file, path)
			}
		}
	}
}

func TestInternalPackagesDoNotPrint(t *testing.T) {
	root := internalRoot(t)
	fset := token.NewFileSet()
	for _, file := range productionFiles(t, root) {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				if x, ok := node.X.(*ast.Ident); ok && x.Name == "fmt" && strings.HasPrefix(node.Sel.Name, "Print") {
					t.Errorf("%s calls fmt.%s", file, node.Sel.Name)
				}
			case *ast.CallExpr:
				if id, ok := node.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
					t.Errorf("%s calls %s", file, id.Name)
				}
			}
			return true
		})
	}
}

func TestApplicationCoreNamesNoAdapter(t *testing.T) {
	root := internalRoot(t)
	adapters := []string{
		modPath + "/internal/system",
		modPath + "/internal/sandbox/plugins",
		modPath + "/internal/isolation/none",
		modPath + "/internal/isolation/bwrap",
		modPath + "/internal/isolation/docker",
		modPath + "/internal/provision/none",
		modPath + "/internal/provision/command",
		modPath + "/internal/provision/nix",
		modPath + "/internal/workspace/git",
		modPath + "/internal/workspace/jj",
		modPath + "/internal/terminal/tmux",
	}
	for pkg, extra := range map[string][]string{
		"launch": {"github.com/spf13/cobra", modPath + "/internal/cmd"},
		"cmd":    nil,
	} {
		banned := append(append([]string{}, adapters...), extra...)
		for _, imp := range dirImports(t, filepath.Join(root, pkg)) {
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
	root := internalRoot(t)
	roots := []string{root, filepath.Join(filepath.Dir(root), "cmd")}

	fset := token.NewFileSet()
	for _, dir := range roots {
		for _, file := range productionFiles(t, dir) {
			f, err := parser.ParseFile(fset, file, nil, 0)
			if err != nil {
				t.Fatalf("parse %s: %v", file, err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if ok && pkg.Name == "types" && strings.HasPrefix(sel.Sel.Name, "Agent") && !agentTypeSurface[sel.Sel.Name] {
					t.Errorf("%s names the concrete agent %s; compare the declared facts instead", file, sel.Sel.Name)
				}
				return true
			})
		}
	}
}
