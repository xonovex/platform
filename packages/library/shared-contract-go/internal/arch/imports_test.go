//go:build arch

// Package arch holds the architecture fitness tests of the library. They read
// the source tree, so they carry the arch build tag and run in the go-test-arch
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

const module = "github.com/xonovex/platform/packages/library/shared-contract-go"

// allowed lists, for each package, the packages of this module it may import
// from its non-test files. A package that is missing from the table, or that
// imports more than its entry lists, fails the test, so a new dependency edge is
// always a decision in this file.
var allowed = map[string][]string{
	"pkg/condition":         {},
	"pkg/discovery":         {"pkg/fs"},
	"pkg/findings":          {},
	"pkg/fs":                {},
	"pkg/fs/memfs":          {"pkg/fs"},
	"pkg/fs/osfs":           {"pkg/fs"},
	"pkg/jsonschema":        {},
	"pkg/jsonschema/tekuri": {"pkg/jsonschema"},
	"pkg/report":            {},
	"pkg/sarif":             {"pkg/findings"},
	"pkg/schemas":           {},
	"pkg/sources":           {"pkg/findings", "pkg/fs", "pkg/jsonschema", "pkg/schemas"},
	"pkg/wiring":            {"pkg/fs", "pkg/jsonschema", "pkg/jsonschema/tekuri", "pkg/schemas", "pkg/sources"},
}

// adapters are the packages that bind a port to a concrete technology. Only the
// composition root imports them from non-test code.
var adapters = []string{"pkg/fs/memfs", "pkg/fs/osfs", "pkg/jsonschema/tekuri"}

// thirdParty lists the packages that may import a module outside the standard
// library, with the modules each may use.
var thirdParty = map[string]string{
	"pkg/jsonschema/tekuri": "github.com/santhosh-tekuri/jsonschema/v6",
}

// moduleFiles parses the module's source, without the fitness tests themselves.
func moduleFiles(t *testing.T) []archtest.File {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	files, err := archtest.Load(os.DirFS(filepath.Dir(filepath.Dir(filepath.Dir(here)))), "pkg")
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

// hostAdapters may import hostImports.
var hostAdapters = map[string]bool{"pkg/fs/osfs": true}

// TestEveryPackageIsDeclared fails for a package missing from the table and for
// a table entry with no package.
func TestEveryPackageIsDeclared(t *testing.T) {
	present := map[string]bool{}
	for _, f := range moduleFiles(t) {
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
// hexagon: the core packages know the ports, adapters know one port each, and
// only the composition root knows them all.
func TestInternalImportsFollowTheTable(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		permitted := allowed[f.Dir]
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if !inside {
				continue
			}
			if !slices.Contains(permitted, rel) {
				t.Errorf("%s imports %s; %s may import only %v", f.Path, rel, f.Dir, permitted)
			}
		}
	}
}

// TestOnlyTheCompositionRootImportsAdapters keeps the concrete technology
// behind the ports.
func TestOnlyTheCompositionRootImportsAdapters(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		if f.Dir == "pkg/wiring" {
			continue
		}
		for _, imp := range f.Imports {
			rel, inside := strings.CutPrefix(imp, module+"/")
			if inside && slices.Contains(adapters, rel) {
				t.Errorf("%s imports the adapter %s; only pkg/wiring may", f.Path, rel)
			}
		}
	}
}

// TestOnlyDeclaredAdaptersImportThirdParty keeps the core on the standard
// library.
func TestOnlyDeclaredAdaptersImportThirdParty(t *testing.T) {
	for _, f := range moduleFiles(t) {
		for _, imp := range f.Imports {
			if archtest.IsStandard(imp) || strings.HasPrefix(imp, module) {
				continue
			}
			want, ok := thirdParty[f.Dir]
			if !ok || !strings.HasPrefix(imp, want) {
				t.Errorf("%s imports %s; %s may import no module outside the standard library except %q", f.Path, imp, f.Dir, want)
			}
		}
	}
}

// TestOnlyTheOSAdapterTouchesTheHost keeps the real disk, environment, process
// and network out of the core and out of every unit test.
func TestOnlyTheOSAdapterTouchesTheHost(t *testing.T) {
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

// TestNoPackageReadsTheClock keeps time an input: a function takes the instant
// it needs as an argument.
func TestNoPackageReadsTheClock(t *testing.T) {
	for _, f := range moduleFiles(t) {
		if archtest.ReadsClock(f.Syntax) {
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

// TestNoPackagePrints keeps the standard streams out of the library: a package
// writes to the io.Writer its caller passes.
func TestNoPackagePrints(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, use := range archtest.Prints(f.Syntax) {
			t.Errorf("%s uses %s; write to the io.Writer the caller passes", f.Path, use)
		}
	}
}

// TestNoPackageKeepsState keeps every package free of state that one caller
// could change under another.
func TestNoPackageKeepsState(t *testing.T) {
	for _, f := range archtest.Production(moduleFiles(t)) {
		for _, state := range archtest.PackageState(f.Syntax) {
			t.Errorf("%s declares %s; pass state as a value", f.Path, state)
		}
	}
}
