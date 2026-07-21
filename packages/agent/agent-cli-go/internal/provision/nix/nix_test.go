package nix

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	sharednix "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision/nix"
)

const testRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func installFakeNix(t *testing.T, body string) {
	t.Helper()

	binDir := t.TempDir()
	path := filepath.Join(binDir, "nix")
	script := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake nix: %v", err)
	}
	t.Setenv("PATH", binDir)
}

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

	in := provshared.Input{NixSource: sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"hello"}}}
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
	in := provshared.Input{NixSource: sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision}}
	if _, err := p.Contribute(in); err == nil {
		t.Error("Contribute(no packages) = nil, want validation error")
	}
	if resolveCalled {
		t.Error("resolve must not run when the source is invalid (fail closed)")
	}
}

func TestContribute_ReportsResolutionAndRootFailures(t *testing.T) {
	source := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"hello"}}
	tests := []struct {
		name    string
		resolve resolveFunc
		root    rootFunc
		phrase  string
	}{
		{
			name: "resolution",
			resolve: func(sharednix.NixSource) (sharednix.ClosureDescriptor, error) {
				return sharednix.ClosureDescriptor{}, errors.New("resolution failed")
			},
			root:   func(sharednix.NixSource, sharednix.ClosureDescriptor) error { return nil },
			phrase: "resolve nix closure",
		},
		{
			name: "root",
			resolve: func(sharednix.NixSource) (sharednix.ClosureDescriptor, error) {
				return sharednix.ClosureDescriptor{StorePaths: []string{"/nix/store/test"}}, nil
			},
			root: func(sharednix.NixSource, sharednix.ClosureDescriptor) error {
				return errors.New("root failed")
			},
			phrase: "register gc-root",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provisioner := &Provisioner{resolve: test.resolve, root: test.root}
			_, err := provisioner.Contribute(provshared.Input{NixSource: source})
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("Contribute() error = %v, want phrase %q", err, test.phrase)
			}
		})
	}
}

func TestPinned(t *testing.T) {
	if !New().Pinned() {
		t.Error("Pinned() = false, want true")
	}
}

func TestSourceFromFlags(t *testing.T) {
	pkgs, err := SourceFromFlags("packages", testRevision, []string{"ripgrep"}, "", "", "", "")
	if err != nil || pkgs.Kind != sharednix.NixSourcePackages || pkgs.Rev != testRevision {
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
	explicit, err := SourceFromFlags("flake", "", nil, "dev", "/flake", "/repo", "/work")
	if err != nil || explicit.FlakeRef != "/flake" || explicit.Shell != "dev" {
		t.Errorf("explicit flake source = %+v, error = %v", explicit, err)
	}

	if _, err := SourceFromFlags("bogus", "", nil, "", "", "", ""); err == nil {
		t.Error("SourceFromFlags(bogus) = nil, want error")
	}
	if _, err := SourceFromFlags("packages", "nixos-unstable", []string{"ripgrep"}, "", "", "", ""); err == nil {
		t.Error("SourceFromFlags(mutable revision) = nil, want immutable revision error")
	}
}

func TestGCRootDir_KeyedByEnvID(t *testing.T) {
	src := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"hello"}}
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

func TestResolveClosure_ValidatesKindsAndRequiredInputs(t *testing.T) {
	if _, err := ResolveClosure(sharednix.NixSource{Kind: "unknown"}); err == nil {
		t.Error("ResolveClosure(unknown) error = nil")
	}
	if _, err := resolvePackages(sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision}); err == nil {
		t.Error("resolvePackages(no packages) error = nil")
	}
	if _, err := resolveFlake(sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake}); err == nil {
		t.Error("resolveFlake(no ref) error = nil")
	}
	if requisites, err := pathInfoRequisites(nil); err != nil || requisites != nil {
		t.Errorf("pathInfoRequisites(nil) = (%v, %v), want (nil, nil)", requisites, err)
	}
}

func TestResolveClosure_WithFakeNix(t *testing.T) {
	installFakeNix(t, `case "$1" in
  build) printf '/nix/store/aaa-ripgrep\n' ;;
  path-info) printf '/nix/store/bbb-glibc\n/nix/store/aaa-ripgrep\n/nix/store/bbb-glibc\n' ;;
  print-dev-env) printf '%s\n' '{"variables":{"PATH":{"type":"exported","value":"/nix/store/ccc-tool/bin:/usr/bin"}}}' ;;
  *) exit 9 ;;
esac`)

	packages, err := ResolveClosure(sharednix.NixSource{
		Kind:     sharednix.NixSourcePackages,
		Rev:      testRevision,
		Packages: []string{"ripgrep"},
	})
	if err != nil {
		t.Fatalf("ResolveClosure(packages) error = %v", err)
	}
	if !slices.Equal(packages.StorePaths, []string{"/nix/store/aaa-ripgrep"}) {
		t.Errorf("package store paths = %v", packages.StorePaths)
	}
	if !slices.Equal(packages.Requisites, []string{"/nix/store/aaa-ripgrep", "/nix/store/bbb-glibc"}) {
		t.Errorf("package requisites = %v", packages.Requisites)
	}

	flake, err := ResolveClosure(sharednix.NixSource{
		Kind:     sharednix.NixSourceProjectFlake,
		FlakeRef: "/repo",
		Shell:    "default",
	})
	if err != nil {
		t.Fatalf("ResolveClosure(flake) error = %v", err)
	}
	if !slices.Equal(flake.StorePaths, []string{"/nix/store/ccc-tool"}) {
		t.Errorf("flake store paths = %v", flake.StorePaths)
	}
	if !slices.Equal(flake.PathEntries, []string{"/nix/store/ccc-tool/bin"}) {
		t.Errorf("flake path entries = %v", flake.PathEntries)
	}
}

func TestResolveFlake_RejectsInvalidOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		phrase string
	}{
		{"invalid JSON", "not-json", "parse print-dev-env JSON"},
		{"missing PATH", `{ "variables": {} }`, "exposes no PATH"},
		{"PATH outside store", `{ "variables": { "PATH": { "value": "/usr/bin" } } }`, "has no /nix/store entries"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			installFakeNix(t, "printf '%s\\n' '"+test.output+"'")
			_, err := resolveFlake(sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"})
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("resolveFlake() error = %v, want phrase %q", err, test.phrase)
			}
		})
	}
}

func TestRunCmd(t *testing.T) {
	out, err := runCmd("/bin/sh", "-c", "printf success")
	if err != nil || out != "success" {
		t.Errorf("runCmd(success) = (%q, %v)", out, err)
	}
	_, err = runCmd("/bin/sh", "-c", "printf failure >&2; exit 3")
	if err == nil || !strings.Contains(err.Error(), "failure") {
		t.Errorf("runCmd(failure) error = %v", err)
	}
}

func TestRegisterGCRoot_RequiresStorePaths(t *testing.T) {
	source := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"hello"}}
	if err := registerGCRoot(source, sharednix.ClosureDescriptor{}); err == nil {
		t.Error("registerGCRoot(empty closure) error = nil")
	}
}

func TestStorePathHelpers(t *testing.T) {
	if got := storeRoot("/nix/store/abc-hello/bin"); got != "/nix/store/abc-hello" {
		t.Errorf("storeRoot = %q, want /nix/store/abc-hello", got)
	}
	if got := storeRoot("/usr/bin"); got != "" {
		t.Errorf("storeRoot(/usr/bin) = %q, want empty", got)
	}
	if got := storeRoot("/nix/store/"); got != "" {
		t.Errorf("storeRoot(incomplete path) = %q, want empty", got)
	}
	entries := storePathEntries("/nix/store/a-x/bin:/usr/bin:/nix/store/b-y/bin")
	if !slices.Equal(entries, []string{"/nix/store/a-x/bin", "/nix/store/b-y/bin"}) {
		t.Errorf("storePathEntries = %v", entries)
	}
	roots := storeRoots([]string{"/nix/store/a-x/bin", "/nix/store/a-x/sbin", "/nix/store/b-y/bin"})
	if !slices.Equal(roots, []string{"/nix/store/a-x", "/nix/store/b-y"}) {
		t.Errorf("storeRoots = %v", roots)
	}
	lines := nonEmptyLines(" first\n\n second \n")
	if !slices.Equal(lines, []string{"first", "second"}) {
		t.Errorf("nonEmptyLines = %v", lines)
	}
	if got := dedupe([]string{"a", "b", "a"}); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("dedupe = %v", got)
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
