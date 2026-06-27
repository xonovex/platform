package nixflake

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/nixenv"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// hasBind reports whether args contains the exact three-token sequence
// `flag src dst` (so a forbidden directory bind cannot false-match a longer
// file path that merely shares its prefix).
func hasBind(args []string, flag, src, dst string) bool {
	for i := 0; i+2 < len(args); i++ {
		if args[i] == flag && args[i+1] == src && args[i+2] == dst {
			return true
		}
	}
	return false
}

// mountsDir reports whether args mounts dir over dir via any bwrap bind flag
// (mount-form-agnostic, so a regression using --bind instead of --ro-bind is
// still caught).
func mountsDir(args []string, dir string) bool {
	for _, flag := range []string{"--ro-bind", "--bind", "--ro-bind-try", "--bind-try", "--dev-bind"} {
		if hasBind(args, flag, dir, dir) {
			return true
		}
	}
	return false
}

// setenvValue returns the value bwrap would set for key, if present.
func setenvValue(args []string, key string) (string, bool) {
	for i := 0; i+2 < len(args); i++ {
		if args[i] == "--setenv" && args[i+1] == key {
			return args[i+2], true
		}
	}
	return "", false
}

func TestParseFlakeConfigDefaults(t *testing.T) {
	e := NewExecutor()
	fc := e.parseFlakeConfig(&types.SandboxConfig{RepoDir: "/repo"})
	if fc.FlakeRef != "/repo" {
		t.Fatalf("FlakeRef = %q, want /repo", fc.FlakeRef)
	}
	if fc.Shell != "default" {
		t.Fatalf("Shell = %q, want default", fc.Shell)
	}
}

func TestParseFlakeConfigWorkDirFallback(t *testing.T) {
	e := NewExecutor()
	fc := e.parseFlakeConfig(&types.SandboxConfig{WorkDir: "/work"})
	if fc.FlakeRef != "/work" {
		t.Fatalf("FlakeRef = %q, want /work (WorkDir fallback)", fc.FlakeRef)
	}
}

func TestParseFlakeConfigExplicit(t *testing.T) {
	e := NewExecutor()
	fc := e.parseFlakeConfig(&types.SandboxConfig{
		Image:   "nixflake:ref=git+https://x#abc,shell=ci",
		WorkDir: "/work",
	})
	if fc.FlakeRef != "git+https://x#abc" || fc.Shell != "ci" {
		t.Fatalf("got %+v", fc)
	}
}

func TestParseLoaderPaths(t *testing.T) {
	out := strings.Join([]string{
		"\tlinux-vdso.so.1 (0x00007fff)",
		"\tlibc.so.6 => /usr/lib64/libc.so.6 (0x00007f00)",
		"\tlibgcc_s.so.1 => /nix/store/abc/lib/libgcc_s.so.1 (0x00007f01)",
		"\t/lib64/ld-linux-x86-64.so.2 (0x00007f02)",
	}, "\n")

	paths, err := parseLoaderPaths(out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"/usr/lib64/libc.so.6",
		"/nix/store/abc/lib/libgcc_s.so.1", // store paths are filtered later, not here
		"/lib64/ld-linux-x86-64.so.2",
	}
	if strings.Join(paths, "|") != strings.Join(want, "|") {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
}

func TestParseLoaderPathsUnresolvedFailsClosed(t *testing.T) {
	out := "\tlibc.so.6 => /usr/lib64/libc.so.6 (0x1)\n\tlibnixutil.so.2 => not found\n"
	if _, err := parseLoaderPaths(out); err == nil {
		t.Fatal("expected an error for an unresolved NEEDED library")
	}
}

func TestDedupeNonStore(t *testing.T) {
	in := []string{
		"/lib64/ld-linux-x86-64.so.2",
		"/usr/lib64/libc.so.6",
		"/nix/store/abc/lib/libgcc_s.so.1", // dropped: already under the bound store
		"/usr/lib64/libc.so.6",             // dropped: duplicate
		"",                                 // dropped: empty
		"/etc/ld.so.cache",
	}
	got := dedupeNonStore(in)
	want := []string{
		"/lib64/ld-linux-x86-64.so.2",
		"/usr/lib64/libc.so.6",
		"/etc/ld.so.cache",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("dedupeNonStore = %v, want %v", got, want)
	}
}

func TestLibcDirOf(t *testing.T) {
	if dir := libcDirOf([]string{"/usr/lib64/libm.so.6", "/usr/lib64/libc.so.6"}); dir != "/usr/lib64" {
		t.Fatalf("libcDirOf = %q, want /usr/lib64", dir)
	}
	if dir := libcDirOf([]string{"/usr/lib64/libm.so.6"}); dir != "" {
		t.Fatalf("libcDirOf = %q, want empty when libc absent", dir)
	}
	if libs := nssResolverLibs(""); libs != nil {
		t.Fatalf("nssResolverLibs(\"\") = %v, want nil", libs)
	}
}

func TestSystemNixMount(t *testing.T) {
	libBinds := []string{"--ro-bind", "/usr/lib64/libc.so.6", "/usr/lib64/libc.so.6"}
	nm := systemNixMount("/opt/customnix/bin/nix", libBinds)

	if nm.binDir != sandboxNixDir {
		t.Fatalf("binDir = %q, want %q", nm.binDir, sandboxNixDir)
	}
	if nm.nixExe != sandboxNixDir+"/nix" {
		t.Fatalf("nixExe = %q, want %q", nm.nixExe, sandboxNixDir+"/nix")
	}
	// The nix binary is bound as a single file into the private dir...
	if !hasBind(nm.binds, "--ro-bind", "/opt/customnix/bin/nix", sandboxNixDir+"/nix") {
		t.Fatal("nix binary not bound into the private dir")
	}
	// ...and the precomputed lib binds are carried through.
	if !hasBind(nm.binds, "--ro-bind", "/usr/lib64/libc.so.6", "/usr/lib64/libc.so.6") {
		t.Fatal("library bind not carried through")
	}
	// The host bin dir is never bound.
	if hasBind(nm.binds, "--ro-bind", "/opt/customnix/bin", "/opt/customnix/bin") {
		t.Fatal("host bin dir was bound")
	}
}

func TestResolveNixMountStore(t *testing.T) {
	nm, err := resolveNixMount("/nix/store/abc-nix-2.0/bin/nix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nm.binDir != "/nix/store/abc-nix-2.0/bin" {
		t.Fatalf("binDir = %q", nm.binDir)
	}
	if nm.nixExe != "/nix/store/abc-nix-2.0/bin/nix" {
		t.Fatalf("nixExe = %q", nm.nixExe)
	}
	if len(nm.binds) != 0 {
		t.Fatalf("store nix needs no extra binds, got %v", nm.binds)
	}
}

func TestBuildDevelopArgs(t *testing.T) {
	fc := nixenv.FlakeSandboxConfig{FlakeRef: "/repo", Shell: "go"}
	agentCmd := []string{"claude", "--flag"}

	for _, nm := range []nixMount{
		{binDir: "/nix/store/x/bin", nixExe: "/nix/store/x/bin/nix"},
		systemNixMount("/opt/customnix/bin/nix", nil),
	} {
		args := buildDevelopArgs(nm, fc, agentCmd)
		if args[0] != nm.nixExe {
			t.Fatalf("argv[0] = %q, want %q", args[0], nm.nixExe)
		}
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "develop --no-write-lock-file /repo#go --command claude --flag") {
			t.Fatalf("develop args malformed: %v", args)
		}
	}
}

func TestBwrapArgsDenyDefault(t *testing.T) {
	e := NewExecutor()
	dirs := e.ensureAgentDirs("test")
	fc := nixenv.FlakeSandboxConfig{FlakeRef: "/repo", Shell: "default"}
	config := &types.SandboxConfig{WorkDir: "/work"}

	storeMount, err := resolveNixMount("/nix/store/x-nix/bin/nix")
	if err != nil {
		t.Fatalf("store mount: %v", err)
	}
	systemMount := systemNixMount("/opt/customnix/bin/nix", []string{
		"--ro-bind", "/usr/lib64/libc.so.6", "/usr/lib64/libc.so.6",
	})

	cases := []struct {
		name     string
		nm       nixMount
		wantPATH string
	}{
		{"store", storeMount, "/nix/store/x-nix/bin"},
		{"system", systemMount, sandboxNixDir},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := e.buildBwrapArgs(config, fc, dirs, tc.nm)

			path, ok := setenvValue(args, "PATH")
			if !ok || path != tc.wantPATH {
				t.Fatalf("PATH = %q (ok=%v), want %q", path, ok, tc.wantPATH)
			}

			// Deny-default: no host executable directory may be mounted, via any
			// bind form.
			for _, dir := range []string{"/usr/bin", "/bin", "/sbin", "/usr/sbin", "/usr/local/bin"} {
				if mountsDir(args, dir) {
					t.Fatalf("host bin dir %s leaked into sandbox", dir)
				}
			}

			// The flake dir and /nix/store are ro-bound.
			if !hasBind(args, "--ro-bind", "/repo", "/repo") {
				t.Fatal("flake dir not ro-bound")
			}
			if !hasBind(args, "--ro-bind", "/nix/store", "/nix/store") {
				t.Fatal("/nix/store not ro-bound")
			}

			// Inherited loader/channel env is neutralized.
			for _, k := range []string{"LD_PRELOAD", "LD_LIBRARY_PATH", "NIX_PATH"} {
				if v, ok := setenvValue(args, k); !ok || v != "" {
					t.Fatalf("%s = %q (ok=%v), want \"\"", k, v, ok)
				}
			}
		})
	}

	// The system mount binds the nix binary as a file into the private dir.
	args := e.buildBwrapArgs(config, fc, dirs, systemMount)
	if !hasBind(args, "--ro-bind", "/opt/customnix/bin/nix", sandboxNixDir+"/nix") {
		t.Fatal("system nix binary not bound into the private dir")
	}
}

func TestNeedsGPU(t *testing.T) {
	if needsGPU(&types.SandboxConfig{}) {
		t.Fatal("no device bindings should not need GPU")
	}
	if !needsGPU(&types.SandboxConfig{BindPaths: []string{"/dev/dri/card0"}}) {
		t.Fatal("/dev/dri bind should be detected as GPU need")
	}
	if !needsGPU(&types.SandboxConfig{RoBindPaths: []string{"/dev/nvidia0"}}) {
		t.Fatal("/dev/nvidia bind should be detected as GPU need")
	}
}

// TestResolveNixMountRealNix exercises resolveNixMount against the host's actual
// nix, asserting the invariant that holds regardless of distro: no ro-bind
// source is a host executable directory and every source exists.
func TestResolveNixMountRealNix(t *testing.T) {
	nixBin, err := exec.LookPath("nix")
	if err != nil {
		t.Skip("nix not available")
	}
	realNix, err := filepath.EvalSymlinks(nixBin)
	if err != nil {
		t.Fatalf("resolve nix: %v", err)
	}

	nm, err := resolveNixMount(realNix)
	if err != nil {
		t.Fatalf("resolveNixMount: %v", err)
	}

	hostBinDirs := map[string]bool{
		"/usr/bin": true, "/bin": true, "/sbin": true,
		"/usr/sbin": true, "/usr/local/bin": true,
	}
	if hostBinDirs[nm.binDir] {
		t.Fatalf("binDir %q is a host bin dir", nm.binDir)
	}

	for i := 0; i+2 < len(nm.binds); i += 3 {
		if nm.binds[i] != "--ro-bind" {
			continue
		}
		src := nm.binds[i+1]
		if hostBinDirs[src] {
			t.Fatalf("bind source %q is a host bin dir", src)
		}
		if _, statErr := os.Stat(src); statErr != nil {
			t.Fatalf("bind source %q does not exist: %v", src, statErr)
		}
	}

	if strings.HasPrefix(realNix, "/nix/store/") {
		if len(nm.binds) != 0 {
			t.Fatalf("store nix should need no binds, got %v", nm.binds)
		}
	} else {
		// A system nix must carry the loader cache.
		if _, statErr := os.Stat("/etc/ld.so.cache"); statErr == nil {
			if !hasBind(nm.binds, "--ro-bind", "/etc/ld.so.cache", "/etc/ld.so.cache") {
				t.Fatal("system nix did not bind /etc/ld.so.cache")
			}
		}
	}
}

// TestIntegrationRealNixDevShell drives the real resolveNixMount/buildBwrapArgs/
// buildDevelopArgs under bwrap and asserts the devShell runs while the host bin
// dir stays unmounted. Env-gated so it never runs in CI.
//
//	NIXFLAKE_INTEGRATION=1 NIXFLAKE_INTEGRATION_FLAKE=/path/to/repo \
//	NIXFLAKE_INTEGRATION_SHELL=go go test ./internal/sandbox/nixflake -run Integration -v
func TestIntegrationRealNixDevShell(t *testing.T) {
	if os.Getenv("NIXFLAKE_INTEGRATION") != "1" {
		t.Skip("set NIXFLAKE_INTEGRATION=1 to run the live bwrap+nix-develop smoke")
	}
	flakeDir := os.Getenv("NIXFLAKE_INTEGRATION_FLAKE")
	if flakeDir == "" {
		t.Skip("set NIXFLAKE_INTEGRATION_FLAKE to a dir whose flake.nix exposes devShells")
	}
	shell := os.Getenv("NIXFLAKE_INTEGRATION_SHELL")
	if shell == "" {
		shell = "default"
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not available")
	}
	nixBin, err := exec.LookPath("nix")
	if err != nil {
		t.Skip("nix not available")
	}
	realNix, err := filepath.EvalSymlinks(nixBin)
	if err != nil {
		t.Fatalf("resolve nix: %v", err)
	}

	nm, err := resolveNixMount(realNix)
	if err != nil {
		t.Fatalf("resolveNixMount: %v", err)
	}

	e := NewExecutor()
	fc := nixenv.FlakeSandboxConfig{FlakeRef: flakeDir, Shell: shell}
	config := &types.SandboxConfig{WorkDir: flakeDir, Network: types.NetworkHost}
	dirs := e.ensureAgentDirs("integration")

	args := e.buildBwrapArgs(config, fc, dirs, nm)
	args = append(args, "--")
	probe := []string{"sh", "-c",
		`if [ -d /usr/bin ]; then echo "USRBIN_PRESENT:$(ls -1 /usr/bin | tr '\n' ',')"; else echo USRBIN_ABSENT; fi; ` +
			`command -v go >/dev/null 2>&1 && echo HAVE_GO; echo NIXFLAKE_INTEGRATION_OK`}
	args = append(args, buildDevelopArgs(nm, fc, probe)...)

	cmd := exec.Command("bwrap", args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	t.Logf("smoke output:\n%s", out)
	if err != nil {
		t.Fatalf("bwrap run failed: %v", err)
	}

	s := string(out)
	if !strings.Contains(s, "NIXFLAKE_INTEGRATION_OK") {
		t.Fatal("nix develop --command did not complete")
	}
	if !strings.Contains(s, "USRBIN_ABSENT") {
		t.Fatalf("deny-default breach: host /usr/bin is present in the sandbox")
	}
}
