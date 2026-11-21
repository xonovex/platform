package nix

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

const testRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

var helloSource = sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"hello"}}

func buildLine(src sharednix.NixSource, pkgs ...string) string {
	args := []string{"build", "--no-link", "--print-out-paths", "--no-write-lock-file"}
	for _, pkg := range pkgs {
		args = append(args, "github:NixOS/nixpkgs/"+src.Rev+"#"+pkg)
	}
	return oshosttest.CommandLine("nix", args...)
}

func pathInfoLine(paths ...string) string {
	return oshosttest.CommandLine("nix", append([]string{"path-info", "-r", "--no-write-lock-file"}, paths...)...)
}

func rootLine(home string, src sharednix.NixSource, storePath string) string {
	link := filepath.Join(gcRootBase(home, src), filepath.Base(storePath))
	return oshosttest.CommandLine("nix-store", "--realise", storePath, "--add-root", link, "--indirect")
}

func gcRootBase(home string, src sharednix.NixSource) string {
	return filepath.Join(home, ".local", "share", "agent-nix", "gcroots", sharednix.ComputeEnvID(src))
}

// helloHost records a successful packages resolution and GC-rooting of hello.
func helloHost() *oshosttest.Fake {
	host := oshosttest.New()
	host.Outcomes[buildLine(helloSource, "hello")] = oshosttest.Outcome{Stdout: "/nix/store/aaa-hello\n"}
	host.Outcomes[pathInfoLine("/nix/store/aaa-hello")] = oshosttest.Outcome{Stdout: "/nix/store/bbb-glibc\n/nix/store/aaa-hello\n/nix/store/bbb-glibc\n"}
	host.Outcomes[rootLine(host.Home, helloSource, "/nix/store/aaa-hello")] = oshosttest.Outcome{}
	return host
}

func TestContribute_BindsRequisitesNotWholeStore(t *testing.T) {
	host := helloHost()

	c, err := New(host).Contribute(provshared.Input{NixSource: helloSource})
	if err != nil {
		t.Fatalf("Contribute err = %v", err)
	}
	wantRequisites := []string{"/nix/store/aaa-hello", "/nix/store/bbb-glibc"}
	if !slices.Equal(c.RoBindPaths, wantRequisites) {
		t.Errorf("RoBindPaths = %v, want the requisites %v", c.RoBindPaths, wantRequisites)
	}
	if slices.Contains(c.RoBindPaths, "/nix/store") {
		t.Error("must bind only the requisites, never the whole /nix/store")
	}
	for _, b := range c.RoBindPaths {
		if strings.Contains(b, "daemon-socket") {
			t.Error("must not contribute the nix daemon socket")
		}
	}
	if !slices.Equal(c.PathEntries, []string{"/nix/store/aaa-hello/bin"}) {
		t.Errorf("PathEntries = %v, want the hello bin directory", c.PathEntries)
	}
	if !host.Exists(gcRootBase(host.Home, helloSource)) {
		t.Error("GC-root directory must be created before the closure is handed off")
	}
	last := host.Calls[len(host.Calls)-1]
	if last.Name != "nix-store" {
		t.Errorf("last call = %q, want the nix-store GC-root registration", last.Line())
	}
}

func TestContribute_FailsClosedOnInvalidSource(t *testing.T) {
	host := oshosttest.New()
	// packages source with no packages fails ValidateSource before resolving.
	in := provshared.Input{NixSource: sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision}}
	if _, err := New(host).Contribute(in); err == nil {
		t.Error("Contribute(no packages) = nil, want validation error")
	}
	if len(host.Calls) != 0 {
		t.Errorf("calls = %v, resolve must not run when the source is invalid (fail closed)", host.Calls)
	}
}

func TestContribute_ReportsResolutionAndRootFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(host *oshosttest.Fake)
		phrase string
	}{
		{
			name: "resolution",
			mutate: func(host *oshosttest.Fake) {
				host.Outcomes[buildLine(helloSource, "hello")] = oshosttest.Outcome{Err: errors.New("build failed")}
			},
			phrase: "resolve nix closure",
		},
		{
			name: "root",
			mutate: func(host *oshosttest.Fake) {
				host.Outcomes[rootLine(host.Home, helloSource, "/nix/store/aaa-hello")] = oshosttest.Outcome{Err: errors.New("root failed")}
			},
			phrase: "register gc-root",
		},
		{
			name:   "home",
			mutate: func(host *oshosttest.Fake) { host.HomeErr = errors.New("no home") },
			phrase: "resolve user home directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host := helloHost()
			test.mutate(host)
			_, err := New(host).Contribute(provshared.Input{NixSource: helloSource})
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("Contribute() error = %v, want phrase %q", err, test.phrase)
			}
		})
	}
}

func TestPinned(t *testing.T) {
	if !New(oshosttest.New()).Pinned() {
		t.Error("Pinned() = false, want true")
	}
}

func TestCheckAgent(t *testing.T) {
	agent := &types.AgentConfig{Binary: "claude", NixPackage: "claude-code"}
	host := oshosttest.New().
		AddExecutable("/nix/store/aaa-claude/bin/claude").
		AddFile("/nix/store/bbb-data/bin/claude", "data", 0o644).
		AddDir("/nix/store/ccc-empty/bin")
	p := New(host)

	tests := []struct {
		name    string
		entries []string
		handled bool
	}{
		{"found after a miss", []string{"/nix/store/ccc-empty/bin", "/nix/store/aaa-claude/bin"}, true},
		{"missing", []string{"/nix/store/ccc-empty/bin"}, false},
		{"not executable", []string{"/nix/store/bbb-data/bin"}, false},
		{"directory", []string{"/nix/store/ccc-empty"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handled, err := p.CheckAgent(provision.Contribution{PathEntries: test.entries}, agent)
			if handled != test.handled {
				t.Errorf("CheckAgent() handled = %v, want %v", handled, test.handled)
			}
			if test.handled && err != nil {
				t.Errorf("CheckAgent() error = %v, want nil", err)
			}
			if !test.handled && (err == nil || !strings.Contains(err.Error(), `include package "claude-code"`)) {
				t.Errorf("CheckAgent() error = %v, want the missing-agent error", err)
			}
		})
	}
}

// TestCheckAgentResolvesAPathAsItself covers a harness override whose binary
// is a path: it names one file, which the closure must hold.
func TestCheckAgentResolvesAPathAsItself(t *testing.T) {
	host := oshosttest.New().
		AddExecutable("/nix/store/aaa-claude/bin/claude").
		AddFile("/nix/store/bbb-data/bin/claude", "data", 0o644).
		AddExecutable("/usr/local/bin/claude").
		AddExecutable("/nix/store/zzz-other/bin/claude")
	p := New(host)
	closure := provision.Contribution{
		RoBindPaths: []string{"/nix/store/bbb-data", "/nix/store/aaa-claude"},
		PathEntries: []string{"/nix/store/aaa-claude/bin"},
	}

	tests := []struct {
		name    string
		binary  string
		handled bool
		phrase  string
	}{
		{"absolute inside the closure", "/nix/store/aaa-claude/bin/claude", true, ""},
		{"absolute inside the closure, not clean", "/nix/store/aaa-claude/bin/../bin/claude", true, ""},
		{"absolute not present", "/nix/store/aaa-claude/bin/codex", false, `include package "claude-code" or select`},
		{"absolute not executable", "/nix/store/bbb-data/bin/claude", false, `include package "claude-code" or select`},
		{"absolute outside the closure on the host", "/usr/local/bin/claude", false, "not inside the nix closure"},
		{"absolute store path outside the closure", "/nix/store/zzz-other/bin/claude", false, "not inside the nix closure"},
		{"absolute sharing a requisite prefix", "/nix/store/aaa-claude-evil/bin/claude", false, "not inside the nix closure"},
		{"relative path", "bin/claude", false, "not inside the nix closure"},
		{"bare name", "claude", true, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agent := &types.AgentConfig{Binary: test.binary, NixPackage: "claude-code"}
			handled, err := p.CheckAgent(closure, agent)
			if handled != test.handled {
				t.Errorf("CheckAgent(%q) handled = %v, want %v", test.binary, handled, test.handled)
			}
			if test.handled && err != nil {
				t.Errorf("CheckAgent(%q) error = %v, want nil", test.binary, err)
			}
			if !test.handled && (err == nil || !strings.Contains(err.Error(), test.phrase) || !strings.Contains(err.Error(), `include package "claude-code"`)) {
				t.Errorf("CheckAgent(%q) error = %v, want phrase %q", test.binary, err, test.phrase)
			}
		})
	}
}

func TestGCRootDir_KeyedByEnvID(t *testing.T) {
	host := oshosttest.New()
	got, err := New(host).gcRootDir(helloSource)
	if err != nil {
		t.Fatalf("gcRootDir() error = %v", err)
	}
	if want := gcRootBase("/home/user", helloSource); got != want {
		t.Errorf("gcRootDir = %q, want %q", got, want)
	}
}

func TestGCRootDirReturnsHomeLookupError(t *testing.T) {
	host := oshosttest.New()
	host.HomeErr = errors.New("no home")

	_, err := New(host).gcRootDir(helloSource)

	if err == nil || !strings.Contains(err.Error(), "resolve user home directory") {
		t.Fatalf("gcRootDir() error = %v, want home lookup error", err)
	}
}

func TestRegisterGCRoot_RequiresStorePaths(t *testing.T) {
	if err := New(oshosttest.New()).registerGCRoot(helloSource, sharednix.ClosureDescriptor{}); err == nil {
		t.Error("registerGCRoot(empty closure) error = nil")
	}
}

func TestRegisterGCRoot_FallsBackToRequisites(t *testing.T) {
	host := oshosttest.New()
	host.Outcomes[rootLine(host.Home, helloSource, "/nix/store/bbb-glibc")] = oshosttest.Outcome{}

	err := New(host).registerGCRoot(helloSource, sharednix.ClosureDescriptor{Requisites: []string{"/nix/store/bbb-glibc"}})

	if err != nil {
		t.Fatalf("registerGCRoot(requisites only) error = %v", err)
	}
	if len(host.Calls) != 1 || host.Calls[0].Args[1] != "/nix/store/bbb-glibc" {
		t.Errorf("calls = %v, want one root of the requisite", host.Calls)
	}
}

func TestResolveClosure_ValidatesKindsAndRequiredInputs(t *testing.T) {
	p := New(oshosttest.New())
	if _, err := p.ResolveClosure(sharednix.NixSource{Kind: "unknown"}); err == nil {
		t.Error("ResolveClosure(unknown) error = nil")
	}
	if _, err := p.resolvePackages(sharednix.NixSource{Kind: sharednix.NixSourcePackages, Packages: []string{"hello"}}); err == nil {
		t.Error("resolvePackages without a rev = nil, want fail-closed error")
	}
	if _, err := p.resolvePackages(sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision}); err == nil {
		t.Error("resolvePackages(no packages) error = nil")
	}
	if _, err := p.resolveFlake(sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake}); err == nil {
		t.Error("resolveFlake(no ref) error = nil")
	}
	if requisites, err := p.pathInfoRequisites(nil); err != nil || requisites != nil {
		t.Errorf("pathInfoRequisites(nil) = (%v, %v), want (nil, nil)", requisites, err)
	}
}

func TestResolveClosure_Packages(t *testing.T) {
	src := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"ripgrep"}}
	host := oshosttest.New()
	host.Outcomes[buildLine(src, "ripgrep")] = oshosttest.Outcome{Stdout: "/nix/store/aaa-ripgrep\n"}
	host.Outcomes[pathInfoLine("/nix/store/aaa-ripgrep")] = oshosttest.Outcome{Stdout: "/nix/store/bbb-glibc\n/nix/store/aaa-ripgrep\n/nix/store/bbb-glibc\n"}

	closure, err := New(host).ResolveClosure(src)

	if err != nil {
		t.Fatalf("ResolveClosure(packages) error = %v", err)
	}
	if !slices.Equal(closure.StorePaths, []string{"/nix/store/aaa-ripgrep"}) {
		t.Errorf("package store paths = %v", closure.StorePaths)
	}
	if !slices.Equal(closure.Requisites, []string{"/nix/store/aaa-ripgrep", "/nix/store/bbb-glibc"}) {
		t.Errorf("package requisites = %v", closure.Requisites)
	}
	if !slices.Equal(closure.PathEntries, []string{"/nix/store/aaa-ripgrep/bin"}) {
		t.Errorf("package path entries = %v", closure.PathEntries)
	}
}

func TestResolvePackages_ReportsBuildAndPathInfoFailures(t *testing.T) {
	src := sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: testRevision, Packages: []string{"ripgrep"}}
	tests := []struct {
		name     string
		build    oshosttest.Outcome
		pathInfo *oshosttest.Outcome
		phrase   string
	}{
		{name: "build fails", build: oshosttest.Outcome{Err: errors.New("build failed")}, phrase: "build failed"},
		{name: "no store paths", build: oshosttest.Outcome{Stdout: "\n"}, phrase: "produced no store paths"},
		{name: "path-info fails", build: oshosttest.Outcome{Stdout: "/nix/store/aaa-ripgrep\n"}, pathInfo: &oshosttest.Outcome{Err: errors.New("path-info failed")}, phrase: "path-info failed"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host := oshosttest.New()
			host.Outcomes[buildLine(src, "ripgrep")] = test.build
			if test.pathInfo != nil {
				host.Outcomes[pathInfoLine("/nix/store/aaa-ripgrep")] = *test.pathInfo
			}
			_, err := New(host).resolvePackages(src)
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("resolvePackages() error = %v, want phrase %q", err, test.phrase)
			}
		})
	}
}

func flakeLine(ref, shell string) string {
	return oshosttest.CommandLine("nix", "print-dev-env", "--json", "--no-write-lock-file", ref+"#"+shell)
}

func TestResolveClosure_Flake(t *testing.T) {
	host := oshosttest.New()
	// The other variables carry the value shapes of a real devShell: an array
	// and an associative array.
	host.Outcomes[flakeLine("/repo", "default")] = oshosttest.Outcome{Stdout: `{"variables":{"PATH":{"type":"exported","value":"/nix/store/ccc-tool/bin:/usr/bin"},"envBuildBuildHooks":{"type":"array","value":["hook"]},"outputs":{"type":"associative","value":{"out":"/nix/store/ddd-out"}}}}`}
	host.Outcomes[pathInfoLine("/nix/store/ccc-tool")] = oshosttest.Outcome{Stdout: "/nix/store/ccc-tool\n"}

	flake, err := New(host).ResolveClosure(sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"})

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
		name    string
		outcome oshosttest.Outcome
		phrase  string
	}{
		{"command fails", oshosttest.Outcome{Err: errors.New("print-dev-env failed")}, "print-dev-env failed"},
		{"invalid JSON", oshosttest.Outcome{Stdout: "not-json"}, "parse print-dev-env JSON"},
		{"missing PATH", oshosttest.Outcome{Stdout: `{ "variables": {} }`}, "exposes no PATH"},
		{"PATH that is no string", oshosttest.Outcome{Stdout: `{ "variables": { "PATH": { "type": "array", "value": ["/nix/store/ccc-tool/bin"] } } }`}, "exposes no PATH"},
		{"PATH outside store", oshosttest.Outcome{Stdout: `{ "variables": { "PATH": { "value": "/usr/bin" } } }`}, "has no /nix/store entries"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host := oshosttest.New()
			host.Outcomes[flakeLine("/repo", "default")] = test.outcome
			_, err := New(host).resolveFlake(sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"})
			if err == nil || !strings.Contains(err.Error(), test.phrase) {
				t.Fatalf("resolveFlake() error = %v, want phrase %q", err, test.phrase)
			}
		})
	}
}

func TestResolveFlake_ReportsPathInfoFailure(t *testing.T) {
	host := oshosttest.New()
	host.Outcomes[flakeLine("/repo", "default")] = oshosttest.Outcome{Stdout: `{"variables":{"PATH":{"value":"/nix/store/ccc-tool/bin"}}}`}

	if _, err := New(host).resolveFlake(sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: "/repo", Shell: "default"}); err == nil {
		t.Fatal("resolveFlake() error = nil, want the path-info failure")
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
