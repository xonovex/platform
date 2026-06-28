package nix

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	sharednix "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision/nix"
)

func TestContribute_BindsRequisitesNotWholeStore(t *testing.T) {
	closure := sharednix.ClosureDescriptor{
		StorePaths:  []string{"/nix/store/aaa-hello"},
		Requisites:  []string{"/nix/store/aaa-hello", "/nix/store/bbb-glibc"},
		PathEntries: []string{"/nix/store/aaa-hello/bin"},
		Env:         map[string]string{"FOO": "bar"},
	}
	rooted := false
	p := &Provisioner{
		resolve: func(sharednix.NixSource) (sharednix.ClosureDescriptor, error) { return closure, nil },
		root:    func(sharednix.NixSource, sharednix.ClosureDescriptor) error { rooted = true; return nil },
	}

	in := provshared.Input{NixSource: sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: "abc123", Packages: []string{"hello"}}}
	c, err := p.Contribute(in)
	if err != nil {
		t.Fatalf("Contribute err = %v", err)
	}
	if !slices.Equal(c.RoBindPaths, closure.Requisites) {
		t.Errorf("RoBindPaths = %v, want the requisites %v", c.RoBindPaths, closure.Requisites)
	}
	if slices.Contains(c.RoBindPaths, "/nix/store") {
		t.Error("must bind only the requisites, never the whole /nix/store")
	}
	for _, b := range c.RoBindPaths {
		if strings.Contains(b, "daemon-socket") {
			t.Error("must not contribute the nix daemon socket")
		}
	}
	if !slices.Equal(c.PathEntries, closure.PathEntries) {
		t.Errorf("PathEntries = %v, want %v", c.PathEntries, closure.PathEntries)
	}
	if c.Env["FOO"] != "bar" {
		t.Error("closure env not contributed")
	}
	if !rooted {
		t.Error("GC-root must be registered before handing off the closure")
	}
}

func TestContribute_FailsClosedOnInvalidSource(t *testing.T) {
	resolveCalled := false
	p := &Provisioner{
		resolve: func(sharednix.NixSource) (sharednix.ClosureDescriptor, error) {
			resolveCalled = true
			return sharednix.ClosureDescriptor{}, nil
		},
		root: func(sharednix.NixSource, sharednix.ClosureDescriptor) error { return nil },
	}
	// packages source with no packages fails ValidateSource before resolving.
	in := provshared.Input{NixSource: sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: "abc"}}
	if _, err := p.Contribute(in); err == nil {
		t.Error("Contribute(no packages) = nil, want validation error")
	}
	if resolveCalled {
		t.Error("resolve must not run when the source is invalid (fail closed)")
	}
}

func TestSourceFromFlags(t *testing.T) {
	pkgs, err := SourceFromFlags("packages", "abc", []string{"ripgrep"}, "", "", "", "")
	if err != nil || pkgs.Kind != sharednix.NixSourcePackages || pkgs.Rev != "abc" {
		t.Fatalf("packages source = %+v, err = %v", pkgs, err)
	}

	// flake source defaults FlakeRef to repoDir (then workDir) and Shell to default.
	flake, err := SourceFromFlags("flake", "", nil, "", "", "/repo", "/work")
	if err != nil || flake.Kind != sharednix.NixSourceProjectFlake || flake.FlakeRef != "/repo" || flake.Shell != defaultFlakeShell {
		t.Fatalf("flake source = %+v, err = %v", flake, err)
	}
	flakeWork, _ := SourceFromFlags("flake", "", nil, "", "", "", "/work")
	if flakeWork.FlakeRef != "/work" {
		t.Errorf("flake source without repoDir = %q, want /work", flakeWork.FlakeRef)
	}

	if _, err := SourceFromFlags("bogus", "", nil, "", "", "", ""); err == nil {
		t.Error("SourceFromFlags(bogus) = nil, want error")
	}
}

func TestGCRootDir_KeyedByEnvID(t *testing.T) {
	src := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: "abc", Packages: []string{"hello"}}
	want := filepath.Join(agentNixDir(), "gcroots", sharednix.ComputeEnvID(src))
	if got := gcRootDir(src); got != want {
		t.Errorf("gcRootDir = %q, want %q", got, want)
	}
}

func TestResolvePackages_RequiresRev(t *testing.T) {
	if _, err := resolvePackages(sharednix.NixSource{Kind: sharednix.NixSourcePackages, Packages: []string{"hello"}}); err == nil {
		t.Error("resolvePackages without a rev = nil, want fail-closed error")
	}
}

func TestStorePathHelpers(t *testing.T) {
	if got := storeRoot("/nix/store/abc-hello/bin"); got != "/nix/store/abc-hello" {
		t.Errorf("storeRoot = %q, want /nix/store/abc-hello", got)
	}
	if got := storeRoot("/usr/bin"); got != "" {
		t.Errorf("storeRoot(/usr/bin) = %q, want empty", got)
	}
	entries := storePathEntries("/nix/store/a-x/bin:/usr/bin:/nix/store/b-y/bin")
	if !slices.Equal(entries, []string{"/nix/store/a-x/bin", "/nix/store/b-y/bin"}) {
		t.Errorf("storePathEntries = %v", entries)
	}
	roots := storeRoots([]string{"/nix/store/a-x/bin", "/nix/store/a-x/sbin", "/nix/store/b-y/bin"})
	if !slices.Equal(roots, []string{"/nix/store/a-x", "/nix/store/b-y"}) {
		t.Errorf("storeRoots = %v", roots)
	}
}

// TestNixprov_Integration resolves a rev-pinned package on the host, GC-roots the
// closure, and verifies the closure is actually rooted (survives GC) — a
// non-destructive proxy for the GC-root regression. Gated on real nix; set
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
	t.Cleanup(func() { _ = os.RemoveAll(gcRootDir(src)) })

	out, err := exec.Command("nix-store", "--query", "--roots", closure.StorePaths[0]).CombinedOutput()
	if err != nil {
		t.Fatalf("nix-store --query --roots: %v: %s", err, out)
	}
	if !strings.Contains(string(out), gcRootDir(src)) {
		t.Errorf("closure not rooted under %s; roots:\n%s", gcRootDir(src), out)
	}
}
