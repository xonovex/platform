//go:build arch

package arch

import (
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/archtest"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
)

// receiverName is the type name of the receiver of a method declaration.
func receiverName(expr ast.Expr) string {
	switch node := expr.(type) {
	case *ast.StarExpr:
		return receiverName(node.X)
	case *ast.IndexExpr:
		return receiverName(node.X)
	case *ast.IndexListExpr:
		return receiverName(node.X)
	case *ast.Ident:
		return node.Name
	}
	return ""
}

// confinements returns each non-test type of the module that declares every
// method of the shared Confinement port, as its module-relative directory and
// its name.
func confinements(t *testing.T) []string {
	t.Helper()
	port := reflect.TypeFor[isolation.Confinement]()
	methods := map[string]map[string]bool{}
	for _, file := range archtest.Production(moduleFiles(t)) {
		for _, decl := range file.Syntax.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			key := file.Dir + "." + receiverName(fn.Recv.List[0].Type)
			if methods[key] == nil {
				methods[key] = map[string]bool{}
			}
			methods[key][fn.Name.Name] = true
		}
	}
	var found []string
	for key, declared := range methods {
		complete := true
		for i := range port.NumMethod() {
			complete = complete && declared[port.Method(i).Name]
		}
		if complete {
			found = append(found, key)
		}
	}
	slices.Sort(found)
	return found
}

// The operator runs every agent as a Job pod, so kube is the only isolation
// it implements: one Confinement, in the kube leaf.
func TestKubeIsTheOnlyIsolation(t *testing.T) {
	if got, want := confinements(t), []string{"internal/isolation/kube.Confinement"}; !slices.Equal(got, want) {
		t.Errorf("types that implement the Confinement port = %v, want %v", got, want)
	}
	entries, err := os.ReadDir(filepath.Join(moduleRoot(t), "internal", "isolation"))
	if err != nil {
		t.Fatalf("read the isolation axis: %v", err)
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	if want := []string{"kube", "shared"}; !slices.Equal(dirs, want) {
		t.Errorf("isolation axis holds %s, want %s: kube is the only isolation leaf", strings.Join(dirs, ", "), strings.Join(want, ", "))
	}
}
