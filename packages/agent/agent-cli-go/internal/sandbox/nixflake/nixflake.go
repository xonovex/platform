// Package nixflake runs an agent inside a project's own flake.nix devShell.
//
// It enters `nix develop <flakeRef>#<shell> --command <agent>` under bubblewrap,
// reusing the bind/namespace machinery of the nix tier but swapping the env
// source: instead of nix-building a synthesized buildEnv, it ro-binds the
// project flake directory so its flake.lock pins the closure. The sandbox PATH
// is the nix bin dir only (deny-default); `nix develop` then prepends the
// devShell's own bin paths.
//
// The host `nix` CLI is exposed two ways depending on the install:
//   - A store-resolved nix (a /nix/store symlink, the recommended deployment)
//     needs nothing extra — its whole closure lives under the already-bound
//     /nix/store and its bin dir holds only nix tooling.
//   - A distro nix (a real ELF under e.g. /usr/bin) is relocated to a private
//     dir holding only the nix binary, and its host-side runtime closure (ELF
//     interpreter + shared libraries) is bound as individual read-only files.
//     The host bin dir is never mounted and never placed on PATH.
//
// Scope: CPU/tooling agents. A pure flake devShell hides the host graphics
// stack (no /usr/lib GL/Vulkan loaders are bound), so an agent that needs a
// GPU/display device is refused — use the nix or bwrap method for those. The
// restored invariant is "PATH is the nix bin dir only, no host bin dir mounted";
// host libraries (and pre-existing UserConfigPaths/daemon-socket binds) are not
// in scope here.
package nixflake

import (
	"debug/elf"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/nixenv"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

// sandboxNixDir is the private in-sandbox directory that holds only the nix
// binary when nix is a non-store (distro) install. Keeping nix here — rather
// than on the host bin dir — guarantees the sandbox PATH exposes nix and nothing
// else.
const sandboxNixDir = "/.nixflake-bin"

// Executor enters a project flake devShell via `nix develop ... --command <agent>`.
//
// Isolation: host tools unreachable. buildBwrapArgs binds /nix/store and the
// ro-bound flake dir but no host /usr,/lib,/bin, and sets PATH to the nix bin dir
// only; the flake.lock devShell then supplies the toolchain. It satisfies
// RequirePinnedToolchain.
type Executor struct{}

// NewExecutor creates a new nixflake executor.
func NewExecutor() *Executor { return &Executor{} }

// IsAvailable checks that nix (with flake support), bwrap, and /nix/store exist.
func (e *Executor) IsAvailable() (bool, error) {
	if _, err := exec.LookPath("nix"); err != nil {
		scriptlib.LogError("nix is not available")
		return false, nil
	}
	if _, err := exec.LookPath("bwrap"); err != nil {
		scriptlib.LogError("bubblewrap (bwrap) is not available")
		return false, nil
	}
	if _, err := os.Stat("/nix/store"); os.IsNotExist(err) {
		scriptlib.LogError("/nix/store does not exist")
		return false, nil
	}
	return true, nil
}

// Execute enters the resolved flake devShell under bwrap and runs the agent.
func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
	// CPU/tooling scope: refuse agents that request a GPU/display device, since
	// a pure flake devShell does not bind the host graphics stack.
	if needsGPU(config) {
		return 1, fmt.Errorf("nixflake: GPU/display agents are unsupported (devShell hides host graphics libs); use the nix or bwrap method")
	}

	fc := e.parseFlakeConfig(config)

	nixBin, err := exec.LookPath("nix")
	if err != nil {
		return 1, fmt.Errorf("nixflake: nix not found: %w", err)
	}
	realNix, err := filepath.EvalSymlinks(nixBin)
	if err != nil {
		return 1, fmt.Errorf("nixflake: cannot resolve nix: %w", err)
	}

	nm, err := resolveNixMount(realNix)
	if err != nil {
		return 1, fmt.Errorf("nixflake: resolve nix runtime: %w", err)
	}

	dirs := e.ensureAgentDirs(config.AgentID)
	bwrapArgs := e.buildBwrapArgs(config, fc, dirs, nm)

	// Deny-default: the agent binary is resolved by the devShell PATH (no /env prefix).
	agentCmd := sandboxutil.BuildAgentCommand(config, "")
	fullCmd := sandboxutil.WrapWithInitCommands(agentCmd, config.SandboxInitCommands)

	bwrapArgs = append(bwrapArgs, "--")
	bwrapArgs = append(bwrapArgs, buildDevelopArgs(nm, fc, fullCmd)...)

	if config.Debug {
		scriptlib.LogDebug(config.Debug, "bwrap "+strings.Join(bwrapArgs, " "))
	}
	if config.Verbose {
		scriptlib.LogInfo("Entering flake devShell " + fc.FlakeRef + "#" + fc.Shell)
	}

	agentEnv, _ := sandboxutil.BuildProviderEnv(config)
	merged := sandboxutil.MergeEnvMaps(agentEnv, sandboxutil.ParseCustomEnv(config.CustomEnv))
	env := append(os.Environ(), sandboxutil.EnvMapToSlice(merged)...)

	return sandboxutil.SpawnSandbox("bwrap", bwrapArgs, env, "Nix flake sandbox", config.Verbose)
}

// GetCommand returns a human-readable description of the nixflake command.
func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
	fc := e.parseFlakeConfig(config)
	return []string{
		"bwrap --ro-bind /nix/store /nix/store \\",
		"  [--ro-bind <nix binary> + its runtime closure for a non-store nix] \\",
		"  --ro-bind " + fc.FlakeRef + " " + fc.FlakeRef + " ... -- \\",
		"  nix develop --no-write-lock-file " + fc.FlakeRef + "#" + fc.Shell + " --command <agent>",
	}
}

// nixMount describes how the host `nix` CLI is exposed inside the sandbox.
// binds are extra `--ro-bind src dst` argument triples for the nix binary and
// (for a non-store nix) its runtime closure. binDir is the sole PATH entry.
// nixExe is the in-sandbox path used as argv[0] of `nix develop`.
type nixMount struct {
	binds  []string
	binDir string
	nixExe string
}

// resolveNixMount decides how to expose `nix` inside the sandbox without leaking
// host tools. A store-resolved nix needs nothing extra: its whole closure lives
// under the already-bound /nix/store and its bin dir holds only nix tooling. A
// distro nix (a real ELF under e.g. /usr/bin) is relocated to a private dir with
// only the nix binary, and its host-side runtime closure (ELF interpreter +
// shared libraries) is bound as individual read-only files. It fails closed
// rather than binding a partial closure.
func resolveNixMount(realNix string) (nixMount, error) {
	if strings.HasPrefix(realNix, "/nix/store/") {
		return nixMount{binDir: filepath.Dir(realNix), nixExe: realNix}, nil
	}
	libBinds, err := systemNixRuntimeBinds(realNix)
	if err != nil {
		return nixMount{}, err
	}
	return systemNixMount(realNix, libBinds), nil
}

// systemNixMount assembles the nixMount for a non-store nix: the nix binary at a
// private path plus its precomputed runtime-library binds. Pure: no I/O.
func systemNixMount(realNix string, libBinds []string) nixMount {
	nixExe := sandboxNixDir + "/nix"
	binds := make([]string, 0, len(libBinds)+3)
	binds = append(binds, "--ro-bind", realNix, nixExe)
	binds = append(binds, libBinds...)
	return nixMount{binds: binds, binDir: sandboxNixDir, nixExe: nixExe}
}

// systemNixRuntimeBinds returns the read-only file binds a non-store nix needs to
// run inside the sandbox: its ELF interpreter, its shared-library closure, the
// dlopen-only NSS/resolver libraries (so getaddrinfo/DNS works), and the loader
// cache. The closure is enumerated with the binary's own loader in --list mode
// under a cleared environment, so inherited LD_*/NIX_* settings cannot misdirect
// it. An unresolved NEEDED library is a hard error (fail closed). Every bind is
// an individual file: bwrap auto-creates the parent as an empty tmpfs, so no host
// directory (and no host executable) is exposed.
func systemNixRuntimeBinds(realNix string) ([]string, error) {
	interp, err := elfInterpreter(realNix)
	if err != nil {
		return nil, fmt.Errorf("read ELF interpreter of %s: %w", realNix, err)
	}

	cmd := exec.Command(interp, "--list", realNix)
	cmd.Env = []string{} // cleared: inherited LD_*/NIX_* must not steer resolution
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("enumerate runtime closure of %s: %w", realNix, err)
	}

	paths, err := parseLoaderPaths(string(out))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", realNix, err)
	}

	candidates := make([]string, 0, len(paths)+8)
	candidates = append(candidates, interp)
	candidates = append(candidates, paths...)
	// glibc dlopens the NSS and resolver backends; they are not in the DT_NEEDED
	// list, so add them explicitly from libc's directory or the sandboxed nix
	// (and any agent it spawns) cannot resolve host names.
	candidates = append(candidates, nssResolverLibs(libcDirOf(paths))...)
	// The loader cache maps sonames to absolute host paths (libs in non-default
	// dirs such as /usr/lib64 are unreachable without it).
	candidates = append(candidates, "/etc/ld.so.cache")

	return existingFileBinds(dedupeNonStore(candidates)), nil
}

// dedupeNonStore drops empty entries, /nix/store paths (already bound via the
// whole-store bind), and duplicates, preserving first-seen order. Pure.
func dedupeNonStore(paths []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" || strings.HasPrefix(p, "/nix/store/") || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// existingFileBinds turns each path into a `--ro-bind p p` triple, skipping
// paths that do not exist. Every entry is an individual file, so bwrap creates
// the parent as an empty tmpfs and no host directory is exposed.
func existingFileBinds(paths []string) []string {
	var binds []string
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		binds = append(binds, "--ro-bind", p, p)
	}
	return binds
}

// libcDirOf returns the directory holding libc.so.6 in a resolved path list, or
// "" if absent. Pure.
func libcDirOf(paths []string) string {
	for _, p := range paths {
		if strings.HasSuffix(p, "/libc.so.6") {
			return filepath.Dir(p)
		}
	}
	return ""
}

// nssResolverLibs returns the NSS/resolver shared libraries in libcDir that
// glibc dlopens for name resolution. Empty when libcDir is "".
func nssResolverLibs(libcDir string) []string {
	if libcDir == "" {
		return nil
	}
	var libs []string
	for _, prefix := range []string{"libnss_dns", "libnss_files", "libnss_compat", "libresolv"} {
		matches, _ := filepath.Glob(filepath.Join(libcDir, prefix+".so*"))
		libs = append(libs, matches...)
	}
	return libs
}

// elfInterpreter returns the PT_INTERP (dynamic loader) path of an ELF binary.
func elfInterpreter(binary string) (string, error) {
	f, err := elf.Open(binary)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	sec := f.Section(".interp")
	if sec == nil {
		return "", fmt.Errorf("%s has no .interp section", binary)
	}
	data, err := sec.Data()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\x00"), nil
}

// parseLoaderPaths extracts absolute library paths from `ld.so --list` (or ldd)
// output. It keeps the first /-rooted token on each line, skips the linux-vdso
// pseudo-entry, and returns an error if any NEEDED library is unresolved
// ("=> not found") so the caller fails closed instead of binding a partial
// closure. Pure: operates only on the given text.
func parseLoaderPaths(out string) ([]string, error) {
	var paths []string
	var unresolved []string

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "not found") {
			if fields := strings.Fields(line); len(fields) > 0 {
				unresolved = append(unresolved, fields[0])
			}
			continue
		}
		if strings.Contains(line, "linux-vdso") {
			continue
		}
		for _, f := range strings.Fields(line) {
			if strings.HasPrefix(f, "/") {
				paths = append(paths, f)
				break
			}
		}
	}

	if len(unresolved) > 0 {
		return nil, fmt.Errorf("unresolved runtime libraries: %s", strings.Join(unresolved, ", "))
	}
	return paths, nil
}

// buildDevelopArgs assembles the `nix develop` invocation that runs inside bwrap:
// it enters the flake devShell with a read-only lock and runs the agent command.
func buildDevelopArgs(nm nixMount, fc nixenv.FlakeSandboxConfig, agentCmd []string) []string {
	develop := []string{
		nm.nixExe, "develop", "--no-write-lock-file",
		fc.FlakeRef + "#" + fc.Shell, "--command",
	}
	return append(develop, agentCmd...)
}

// parseFlakeConfig parses the flake ref + shell from config.Image of form
// "nixflake:ref=...,shell=...". When no ref is given, it defaults to the bound
// project flake dir (RepoDir, then WorkDir) so the in-sandbox flake.lock pins
// the closure.
func (e *Executor) parseFlakeConfig(config *types.SandboxConfig) nixenv.FlakeSandboxConfig {
	fc := nixenv.FlakeSandboxConfig{Shell: nixenv.DefaultFlakeShell}

	spec := strings.TrimPrefix(config.Image, "nixflake:")
	if spec != config.Image { // had the prefix
		for _, kv := range strings.Split(spec, ",") {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				continue
			}
			switch strings.TrimSpace(k) {
			case "ref":
				fc.FlakeRef = strings.TrimSpace(v)
			case "shell":
				fc.Shell = strings.TrimSpace(v)
			}
		}
	}

	if fc.FlakeRef == "" {
		if config.RepoDir != "" {
			fc.FlakeRef = config.RepoDir
		} else {
			fc.FlakeRef = config.WorkDir
		}
	}
	if fc.Shell == "" {
		fc.Shell = nixenv.DefaultFlakeShell
	}
	return fc
}

// buildBwrapArgs builds bubblewrap args for the flake devShell. Unlike the nix
// tier it binds no /env: the flake dir is ro-bound (flake.lock pins the closure)
// and PATH is the nix bin dir only.
func (e *Executor) buildBwrapArgs(config *types.SandboxConfig, fc nixenv.FlakeSandboxConfig, dirs *AgentDirs, nm nixMount) []string {
	homeDir, _ := os.UserHomeDir()
	e.ensureSandboxMountPoint(dirs.Home, config.WorkDir)

	args := []string{
		"--ro-bind", "/nix/store", "/nix/store",
	}
	// The nix binary plus, for a non-store nix, its host-side runtime closure.
	args = append(args, nm.binds...)
	args = append(args,
		"--bind", dirs.Work, "/work",
		"--bind", dirs.Tmp, "/tmp",
		"--bind", dirs.Home, homeDir,
	)

	// Daemon socket so the sandboxed `nix develop` resolves the flake against
	// the host store.
	if _, err := os.Stat("/nix/var/nix/daemon-socket/socket"); err == nil {
		args = append(args, "--ro-bind", "/nix/var/nix/daemon-socket/socket", "/nix/var/nix/daemon-socket/socket")
	}

	// Flake dir read-only so flake.lock pins the closure (paired with
	// --no-write-lock-file). Only bind a local absolute path that is not the
	// work dir; a remote ref (git+https://...) is fetched over the network.
	if strings.HasPrefix(fc.FlakeRef, "/") && fc.FlakeRef != config.WorkDir {
		args = append(args, "--ro-bind", fc.FlakeRef, fc.FlakeRef)
	}

	for _, configPath := range sandbox.UserConfigPaths {
		sourcePath := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(sourcePath); err == nil {
			args = append(args, "--bind", sourcePath, sourcePath)
		}
	}
	args = append(args, "--bind", config.WorkDir, config.WorkDir)

	args = append(args, "--proc", "/proc", "--dev", "/dev")
	args = append(args, "--unshare-uts", "--unshare-ipc", "--unshare-pid", "--unshare-cgroup")
	if config.Network {
		args = append(args, "--share-net")
	} else {
		args = append(args, "--unshare-net")
	}

	env := map[string]string{
		"HOME":              homeDir,
		"TMPDIR":            "/tmp",
		"PATH":              nm.binDir, // deny-default: nix devShell prepends its own bin
		"LD_PRELOAD":        "",        // neutralize host loader injection
		"LD_LIBRARY_PATH":   "",        // host lib paths must not steer the sandboxed nix
		"NIX_PATH":          "",        // flakes only; ignore any inherited channel path
		"NIX_REMOTE":        "daemon",
		"NIX_SSL_CERT_FILE": "/etc/ssl/certs/ca-certificates.crt",
		"NIX_CONFIG":        "experimental-features = nix-command flakes",
	}
	for k, v := range sandboxutil.ParseCustomEnv(config.CustomEnv) {
		env[k] = v
	}
	for k, v := range env {
		args = append(args, "--setenv", k, v)
	}

	if _, err := os.Stat("/etc/ssl/certs"); err == nil {
		args = append(args, "--ro-bind", "/etc/ssl/certs", "/etc/ssl/certs")
	}
	if _, err := os.Stat("/etc/resolv.conf"); err == nil {
		args = append(args, "--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf")
	}
	if _, err := os.Stat("/etc/nsswitch.conf"); err == nil {
		args = append(args, "--ro-bind", "/etc/nsswitch.conf", "/etc/nsswitch.conf")
	}
	if _, err := os.Stat("/etc/hosts"); err == nil {
		args = append(args, "--ro-bind", "/etc/hosts", "/etc/hosts")
	}
	args = append(args, "--chdir", config.WorkDir, "--die-with-parent")
	return args
}

// AgentDirs holds per-agent runtime directories.
type AgentDirs struct {
	Root string
	Work string
	Tmp  string
	Home string
}

// ensureAgentDirs creates per-agent directories.
func (e *Executor) ensureAgentDirs(agentID string) *AgentDirs {
	if agentID == "" {
		agentID = "default"
	}

	root := filepath.Join(nixenv.GetAgentsDir(), agentID)
	dirs := &AgentDirs{
		Root: root,
		Work: filepath.Join(root, "work"),
		Tmp:  filepath.Join(root, "tmp"),
		Home: filepath.Join(root, "home"),
	}

	_ = os.MkdirAll(dirs.Work, 0755)
	_ = os.MkdirAll(dirs.Tmp, 0755)
	_ = os.MkdirAll(dirs.Home, 0755)

	return dirs
}

// ensureSandboxMountPoint ensures the mount point directory exists in sandboxHome.
func (e *Executor) ensureSandboxMountPoint(sandboxHome string, targetPath string) {
	homeDir, _ := os.UserHomeDir()

	if !strings.HasPrefix(targetPath, homeDir+"/") {
		return
	}

	relativePath, err := filepath.Rel(homeDir, targetPath)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return
	}

	mountPointInSandbox := filepath.Join(sandboxHome, relativePath)
	_ = os.MkdirAll(mountPointInSandbox, 0755)
}

// needsGPU reports whether the run requests a GPU/display device through a bind
// path. The current AgentConfig has no GPU field, so the request is inferred
// from explicitly bound device paths.
func needsGPU(config *types.SandboxConfig) bool {
	for _, p := range config.BindPaths {
		if isGPUDevicePath(p) {
			return true
		}
	}
	for _, p := range config.RoBindPaths {
		if isGPUDevicePath(p) {
			return true
		}
	}
	return false
}

// isGPUDevicePath reports whether a path is a host GPU/display device node.
func isGPUDevicePath(path string) bool {
	return strings.HasPrefix(path, "/dev/dri") || strings.HasPrefix(path, "/dev/nvidia")
}
