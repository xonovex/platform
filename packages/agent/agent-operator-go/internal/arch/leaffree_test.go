package arch

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

const opModule = "github.com/xonovex/platform/packages/agent/agent-operator-go"

// prodImports returns the import paths of the non-test Go files in dir.
func prodImports(t *testing.T, dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	fset := token.NewFileSet()
	var imps []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !hasSuffix(name, ".go") || hasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, im := range f.Imports {
			if p, err := strconv.Unquote(im.Path.Value); err == nil {
				imps = append(imps, p)
			}
		}
	}
	return imps
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// TestSharedCoresAreLeafFree locks the invariant that each operator AXIS-PORT
// shared/ core names no concrete leaf, directly OR via the composition root — the
// registries that wire leaves live only in internal/plugins, keeping the ports
// neutral (symmetric with the CLI's leaf-free shared cores).
//
// isolation/shared is intentionally EXCLUDED: it is the pod composition layer (the
// single pod realizer), not an axis port. It composes every axis, so it imports
// internal/plugins to resolve the harness command / VCS strategy / toolchain —
// exactly as the CLI's cmd package imports sandbox/plugins. The composition layer
// depending on the registries is correct; only the ports below it must stay neutral.
func TestSharedCoresAreLeafFree(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	opInternal := filepath.Dir(filepath.Dir(file)) // .../internal/arch -> .../internal

	leaves := map[string]bool{
		opModule + "/internal/provision/nix":    true,
		opModule + "/internal/workspace/git":    true,
		opModule + "/internal/workspace/jj":     true,
		opModule + "/internal/harness/claude":   true,
		opModule + "/internal/harness/opencode": true,
	}
	// The composition root is the only package permitted to reach the leaves, so a
	// port importing it would transitively acquire a leaf — also a violation.
	const compositionRoot = opModule + "/internal/plugins"
	ports := []string{"provision/shared", "workspace/shared", "harness/shared", "network/shared"}

	for _, core := range ports {
		dir := filepath.Join(opInternal, filepath.FromSlash(core))
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		for _, imp := range prodImports(t, dir) {
			if leaves[imp] || imp == compositionRoot {
				t.Errorf("axis port %s imports %s; leaf wiring must live only in internal/plugins, not in a port", core, imp)
			}
		}
	}
}
