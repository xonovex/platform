//go:build integration

package nix

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	sharednix "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision/nix"
)

// TestNixprov_Integration resolves a rev-pinned package on the host, GC-roots the
// closure, and verifies the closure is actually rooted (survives GC), a
// non-destructive proxy for the GC-root regression. It runs nix and nix-store, so
// it carries the integration build tag on top of its own gate: set
// NIXPROV_INTEGRATION=1 and NIXPROV_INTEGRATION_REV=<nixpkgs rev>.
func TestNixprov_Integration(t *testing.T) {
	for _, bin := range []string{"nix", "nix-store"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not available", bin)
		}
	}
	rev := os.Getenv("NIXPROV_INTEGRATION_REV")
	if os.Getenv("NIXPROV_INTEGRATION") != "1" || rev == "" {
		t.Skip("set NIXPROV_INTEGRATION=1 and NIXPROV_INTEGRATION_REV=<nixpkgs rev> to run")
	}

	src := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: rev, Packages: []string{"hello"}}
	closure, err := ResolveClosure(src)
	if err != nil {
		t.Fatalf("ResolveClosure: %v", err)
	}
	if len(closure.Requisites) == 0 {
		t.Fatal("closure has no requisites")
	}
	for _, r := range closure.Requisites {
		if !strings.HasPrefix(r, "/nix/store/") {
			t.Errorf("requisite %q is not a store path", r)
		}
	}

	if err := registerGCRoot(src, closure); err != nil {
		t.Fatalf("registerGCRoot: %v", err)
	}
	rootDir, err := gcRootDir(src)
	if err != nil {
		t.Fatalf("gcRootDir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(rootDir) })

	out, err := exec.Command("nix-store", "--query", "--roots", closure.StorePaths[0]).CombinedOutput()
	if err != nil {
		t.Fatalf("nix-store --query --roots: %v: %s", err, out)
	}
	if !strings.Contains(string(out), rootDir) {
		t.Errorf("closure not rooted under %s; roots:\n%s", rootDir, out)
	}
}
