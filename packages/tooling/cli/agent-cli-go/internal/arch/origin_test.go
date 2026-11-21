//go:build arch

package arch

// Fitness tests of the origin axis: a shared core, the leaves directory and git,
// and one composition root. They lock the same invariants as the sandbox axes.

import (
	"go/ast"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"
)

// originLeaves are the concrete origins below internal/origin.
var originLeaves = []string{"directory", "git"}

// originRoot is the registry package, the one place that names a concrete
// origin.
const originRoot = "internal/origin/plugins"

func TestOriginLeavesKnowOnlyTheSharedCore(t *testing.T) {
	files := moduleFiles(t)
	for _, leaf := range originLeaves {
		for _, imp := range dirImports(files, "internal/origin/"+leaf) {
			if rest, ok := strings.CutPrefix(imp, modPath+"/internal/origin/"); ok && rest != "shared" {
				t.Errorf("leaf %s imports %s", leaf, imp)
			}
		}
	}
}

func TestTheOriginSharedCoreKnowsNoLeafAndNoPluginStore(t *testing.T) {
	for _, imp := range dirImports(moduleFiles(t), "internal/origin/shared") {
		if strings.HasPrefix(imp, modPath+"/internal/origin/") || imp == modPath+"/internal/plugin" {
			t.Errorf("the shared core imports %s", imp)
		}
	}
}

// TestOnlyTheOriginCompositionRootImportsTheLeaves keeps the registry the one
// place that imports a concrete origin.
func TestOnlyTheOriginCompositionRootImportsTheLeaves(t *testing.T) {
	for _, f := range filesUnder(moduleFiles(t), "internal") {
		if f.Dir == originRoot || strings.HasPrefix(f.Dir, originRoot+"/") {
			continue
		}
		for _, imp := range f.Imports {
			for _, leaf := range originLeaves {
				if imp == modPath+"/internal/origin/"+leaf {
					t.Errorf("%s imports the origin %s; only origin/plugins may", f.Path, leaf)
				}
			}
		}
	}
}

// TestOriginLeafConstructorsAreCalledOnlyByTheCompositionRoot checks the call
// sites as well: a leaf named through a local alias is still found.
func TestOriginLeafConstructorsAreCalledOnlyByTheCompositionRoot(t *testing.T) {
	files := filesUnder(moduleFiles(t), "internal")
	for _, leaf := range originLeaves {
		found := 0
		for _, f := range files {
			for _, call := range constructorCalls(f, leaf) {
				found++
				if f.Path != originRoot+"/plugins.go" {
					t.Errorf("%s calls %s", f.Path, call)
				}
			}
		}
		if found != 1 {
			t.Errorf("the composition root calls %s.New %d times, want 1", leaf, found)
		}
	}
}

// constructorCalls lists each call of pkg.New in the file.
func constructorCalls(f archtest.File, pkg string) []string {
	var calls []string
	ast.Inspect(f.Syntax, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "New" {
			return true
		}
		if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == pkg {
			calls = append(calls, pkg+".New")
		}
		return true
	})
	return calls
}
