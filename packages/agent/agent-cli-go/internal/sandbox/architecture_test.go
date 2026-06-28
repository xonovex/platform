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

const modPath = "github.com/xonovex/platform/packages/cli/agent-cli-go"

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
	for _, axis := range []string{"isolation", "provision", "network"} {
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
	for _, axis := range []string{"isolation", "provision", "network", "workspace"} {
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
	axes := []string{"isolation", "provision", "network", "workspace"}
	for _, axis := range []string{"isolation", "provision", "network"} {
		dir := filepath.Join(root, axis, "shared")
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		for _, imp := range dirImports(t, dir) {
			for _, other := range axes {
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

	fset := token.NewFileSet()
	for _, file := range files {
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
