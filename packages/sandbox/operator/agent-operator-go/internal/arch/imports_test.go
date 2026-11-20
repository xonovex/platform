//go:build arch

// Package arch holds the architecture fitness tests of the operator. They read
// the source tree, so they carry the arch build tag and run in the go-test-arch
// task, not in the unit run.
package arch

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go"

// leaves are the concrete variants of the operator axes.
var leaves = []string{
	"internal/harness/claude",
	"internal/harness/opencode",
	"internal/provision/nix",
	"internal/workspace/git",
	"internal/workspace/jj",
}

// sourceFile is one Go file of the module: the module-relative directory of
// its package, whether it is a test file, and its import paths.
type sourceFile struct {
	path    string
	dir     string
	test    bool
	imports []string
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

// moduleFiles parses the imports of every Go file of the module, whatever its
// build tags.
func moduleFiles(t *testing.T) []sourceFile {
	t.Helper()
	root := moduleRoot(t)
	fset := token.NewFileSet()
	var files []sourceFile
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testdata" || (strings.HasPrefix(entry.Name(), ".") && path != root) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		parsed, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		dir, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		file := sourceFile{path: path, dir: filepath.ToSlash(dir), test: strings.HasSuffix(path, "_test.go")}
		for _, spec := range parsed.Imports {
			if imported, err := strconv.Unquote(spec.Path.Value); err == nil {
				file.imports = append(file.imports, imported)
			}
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if len(files) == 0 {
		t.Fatalf("no Go files under %s", root)
	}
	return files
}

// importers returns the files that import the module package at dir.
func importers(files []sourceFile, dir string) []sourceFile {
	var found []sourceFile
	for _, file := range files {
		for _, imported := range file.imports {
			if imported == module+"/"+dir {
				found = append(found, file)
			}
		}
	}
	return found
}

// Each leaf is reached only through the registry, so no core, adapter or test
// outside the registry depends on a concrete variant.
func TestLeavesAreImportedOnlyByTheRegistry(t *testing.T) {
	files := moduleFiles(t)
	for _, leaf := range leaves {
		for _, file := range importers(files, leaf) {
			if file.dir != "internal/plugins" && file.dir != leaf {
				t.Errorf("%s imports the leaf %s; only internal/plugins may", file.path, leaf)
			}
		}
	}
}

// The registry and the controller-runtime adapters are bound to the ports in
// one place. The sample tests of config admit the shipped samples against the
// real registry.
func TestTheCompositionRootAloneImportsTheRegistryAndTheAdapters(t *testing.T) {
	files := moduleFiles(t)
	allowed := map[string]map[string]bool{
		"internal/plugins": {"internal/plugins": true, "internal/wiring": true, "config": true},
		"internal/kube":    {"internal/kube": true, "internal/wiring": true},
	}
	for dir, importersAllowed := range allowed {
		for _, file := range importers(files, dir) {
			if !importersAllowed[file.dir] || (file.dir == "config" && !file.test) {
				t.Errorf("%s imports %s; only the composition root internal/wiring may", file.path, dir)
			}
		}
	}
}

// Only the driving adapters (controller, webhook), the driven adapter (kube)
// and the composition root (wiring) depend on controller-runtime; every other
// core stays free of it.
func TestOnlyTheAdaptersImportControllerRuntime(t *testing.T) {
	adapters := map[string]bool{
		"internal/controller": true,
		"internal/webhook":    true,
		"internal/kube":       true,
		"internal/wiring":     true,
	}
	for _, file := range moduleFiles(t) {
		if file.test || !strings.HasPrefix(file.dir, "internal/") || adapters[file.dir] {
			continue
		}
		for _, imported := range file.imports {
			if strings.HasPrefix(imported, "sigs.k8s.io/controller-runtime") {
				t.Errorf("%s imports %s; a core names only its own ports", file.path, imported)
			}
		}
	}
}

// launchLibrary is the module path of the shared run kernel.
const launchLibrary = "github.com/xonovex/platform/packages/library/shared-launch-go"

// launchCore are the packages of the shared run kernel that start nothing on a
// machine. The operator runs an agent as a Job, so it never reaches a launcher
// leaf or the operating-system adapter of the library.
var launchCore = map[string]bool{
	"pkg/credential": true,
	"pkg/isolation":  true,
	"pkg/launch":     true,
	"pkg/provision":  true,
	"pkg/record":     true,
	"pkg/sandbox":    true,
}

func TestOnlyTheLaunchCoreIsImported(t *testing.T) {
	for _, file := range moduleFiles(t) {
		for _, imported := range file.imports {
			if pkg, inside := strings.CutPrefix(imported, launchLibrary+"/"); inside && !launchCore[pkg] {
				t.Errorf("%s imports %s; the operator imports only the launch core", file.path, imported)
			}
		}
	}
}
