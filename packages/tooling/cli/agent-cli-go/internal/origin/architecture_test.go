package origin

// Fitness tests of the origin axis: a shared core, the leaves directory and git,
// and one composition root. They lock the same invariants as the sandbox axes.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const modPath = "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go"

var leaves = []string{"directory", "git"}

func originRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func importsOf(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var imports []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			path, _ := strconv.Unquote(spec.Path.Value)
			imports = append(imports, path)
		}
	}
	return imports
}

func TestLeavesKnowOnlyTheSharedCore(t *testing.T) {
	root := originRoot(t)
	for _, leaf := range leaves {
		for _, path := range importsOf(t, filepath.Join(root, leaf)) {
			if rest, ok := strings.CutPrefix(path, modPath+"/internal/origin/"); ok && rest != "shared" {
				t.Errorf("leaf %s imports %s", leaf, path)
			}
		}
	}
}

func TestTheSharedCoreKnowsNoLeafAndNoPluginStore(t *testing.T) {
	for _, path := range importsOf(t, filepath.Join(originRoot(t), "shared")) {
		if strings.HasPrefix(path, modPath+"/internal/origin/") || path == modPath+"/internal/plugin" {
			t.Errorf("the shared core imports %s", path)
		}
	}
}

// TestOnlyTheCompositionRootImportsTheLeaves keeps the registry the one place
// that names a concrete origin.
func TestOnlyTheCompositionRootImportsTheLeaves(t *testing.T) {
	internal := filepath.Dir(originRoot(t))
	allowed := filepath.Join(originRoot(t), "plugins") + string(filepath.Separator)
	err := filepath.WalkDir(internal, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") || strings.HasPrefix(path, allowed) {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imported, _ := strconv.Unquote(spec.Path.Value)
			for _, leaf := range leaves {
				if imported == modPath+"/internal/origin/"+leaf {
					t.Errorf("%s imports the origin %s; only origin/plugins may", path, leaf)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestLeafConstructorsAreCalledOnlyByTheCompositionRoot checks the call sites as
// well: a leaf named through a local alias is still found.
func TestLeafConstructorsAreCalledOnlyByTheCompositionRoot(t *testing.T) {
	root := originRoot(t)
	for _, leaf := range leaves {
		found := 0
		err := filepath.WalkDir(filepath.Dir(root), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "New" {
					return true
				}
				if pkg, ok := selector.X.(*ast.Ident); ok && pkg.Name == leaf {
					found++
					if !strings.HasSuffix(path, filepath.Join("origin", "plugins", "plugins.go")) {
						t.Errorf("%s calls %s.New", path, leaf)
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if found != 1 {
			t.Errorf("the composition root calls %s.New %d times, want 1", leaf, found)
		}
	}
}
