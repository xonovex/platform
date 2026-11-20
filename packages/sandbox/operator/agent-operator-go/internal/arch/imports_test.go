//go:build arch

// Package arch holds the architecture fitness tests of the operator. They read
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

const module = "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go"

// leaves are the concrete variants of the operator axes.
var leaves = []string{
	"internal/isolation/kube",
	"internal/provision/nix",
	"internal/provision/none",
	"internal/workspace/git",
	"internal/workspace/jj",
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
	files, err := archtest.Load(os.DirFS(moduleRoot(t)), ".")
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

// importers returns the files that import the module package at dir.
func importers(files []archtest.File, dir string) []archtest.File {
	var found []archtest.File
	for _, file := range files {
		for _, imported := range file.Imports {
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
			if file.Dir != "internal/plugins" && file.Dir != leaf {
				t.Errorf("%s imports the leaf %s; only internal/plugins may", file.Path, leaf)
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
			if !importersAllowed[file.Dir] || (file.Dir == "config" && !file.Test) {
				t.Errorf("%s imports %s; only the composition root internal/wiring may", file.Path, dir)
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
		if file.Test || !strings.HasPrefix(file.Dir, "internal/") || adapters[file.Dir] {
			continue
		}
		for _, imported := range file.Imports {
			if strings.HasPrefix(imported, "sigs.k8s.io/controller-runtime") {
				t.Errorf("%s imports %s; a core names only its own ports", file.Path, imported)
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
		for _, imported := range file.Imports {
			if pkg, inside := strings.CutPrefix(imported, launchLibrary+"/"); inside && !launchCore[pkg] {
				t.Errorf("%s imports %s; the operator imports only the launch core", file.Path, imported)
			}
		}
	}
}

// compositionRoots are the main packages: they alone read the environment,
// print and bind the adapters.
var compositionRoots = map[string]bool{"cmd/operator": true, "cmd/e2e-summary": true}

// endToEnd reports a file that builds only for an end-to-end tier, which drives
// a real cluster from the machine that runs the tests.
func endToEnd(f archtest.File) bool {
	return slices.ContainsFunc([]string{"e2e", "e2e_gvisor", "e2e_kata", "e2e_coco"}, f.BuildsOnlyWith)
}

// TestOnlyTheCompositionRootsTouchTheHost keeps the disk, the environment, the
// process and the network out of every core and adapter: the operator reaches
// the cluster through controller-runtime, never through the machine it runs on.
func TestOnlyTheCompositionRootsTouchTheHost(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if compositionRoots[f.Dir] || endToEnd(f) {
			continue
		}
		for _, imp := range f.Imports {
			if archtest.IsHost(imp) {
				t.Errorf("%s imports %s; only %v may", f.Path, imp, []string{"cmd/e2e-summary", "cmd/operator"})
			}
		}
	}
}

// TestOnlyTheCompositionRootsPrint keeps the standard streams in the main
// packages: a reconciler and a webhook report through their results.
func TestOnlyTheCompositionRootsPrint(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if compositionRoots[f.Dir] {
			continue
		}
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; report through a result or the logger", f.Path, use)
		}
	}
}

// TestNoPackageKeepsState keeps every package free of state that one
// reconcile could leave behind for the next.
func TestNoPackageKeepsState(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if compositionRoots[f.Dir] {
			continue
		}
		for _, state := range archtest.PackageState(f.Syntax) {
			t.Errorf("%s declares %s; pass state as a value", f.Path, state)
		}
	}
}
