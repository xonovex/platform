---
type: plan
has_subplans: false
parent_plan: plans/nix-toolchain-hardening.md
parallel_group: 3
status: pending
dependencies:
  plans: []
  files:
    - packages/shared/shared-agent-go/pkg/types/sandbox.go
    - packages/agent/agent-cli-go/internal/sandbox/registry.go
    - packages/agent/agent-cli-go/internal/sandbox/nixflake/nixflake.go
    - packages/agent/agent-cli-go/internal/sandbox/nixflake/nixflake_test.go
    - packages/agent/agent-cli-go/internal/nixenv/types.go
skills_to_consult: [general-fp-guide, code-review-guide, debugging-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Sandbox Flake devShell Mode (`nixflake`)

## Objective

Give the agent sandbox a flake-pinned path that enters a project's own `flake.nix` devShell via `nix develop <flakeRef>#<shell> --command <agentCmd>`, instead of only the synthesized flat `pkgs.buildEnv` produced by `internal/nixenv/render.go`. This reuses the existing bwrap bind/namespace machinery from `internal/sandbox/nix/nix.go` but ro-binds `/nix/store` plus the project flake directory (so `flake.lock` pins the closure) rather than nix-building a buildEnv. The mode keeps deny-default semantics (sandbox `PATH` never includes host `/usr/bin`, `/bin`).

## Tasks

1. **Register a new `nixflake` sandbox method and executor.** Add the method constant in `packages/shared/shared-agent-go/pkg/types/sandbox.go:11` and wire it into the dispatcher at `packages/agent/agent-cli-go/internal/sandbox/registry.go:9-30` and the availability list at `:34-40`.

   ```go
   // packages/shared/shared-agent-go/pkg/types/sandbox.go
   const (
       SandboxNone     SandboxMethod = "none"
       SandboxBwrap    SandboxMethod = "bwrap"
       SandboxDocker   SandboxMethod = "docker"
       SandboxCompose  SandboxMethod = "compose"
       SandboxNix      SandboxMethod = "nix"
       SandboxNixFlake SandboxMethod = "nixflake"
   )
   ```

   ```go
   // packages/agent/agent-cli-go/internal/sandbox/registry.go
   import (
       // ...existing imports...
       "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/nix"
       "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/nixflake"
       // ...
   )

   func GetExecutor(method types.SandboxMethod) (types.SandboxExecutor, error) {
       switch method {
       // ...existing cases...
       case types.SandboxNix:
           return nix.NewExecutor(), nil
       case types.SandboxNixFlake:
           return nixflake.NewExecutor(), nil
       default:
           return nil, fmt.Errorf("unknown sandbox method: %s", method)
       }
   }
   ```
   Note: also append `types.SandboxNixFlake` to `allMethods` in `GetAvailableMethods` so it is probed.

2. **Add the flake config surface.** Extend `packages/agent/agent-cli-go/internal/nixenv/types.go` (currently ends at `NixSandboxConfig`, lines 38-43) with a `FlakeSandboxConfig` carrying the flake ref + shell selector. Keep it a plain value type (FP: no methods, pure parse downstream).

   ```go
   // packages/agent/agent-cli-go/internal/nixenv/types.go
   // FlakeSandboxConfig configures the nixflake sandbox: which flake devShell to enter.
   type FlakeSandboxConfig struct {
       FlakeRef string // e.g. "/repo" or "git+https://...#rev"; defaults to the bound flake dir
       Shell    string // attribute under devShells.<system>; defaults to "default"
   }

   // DefaultFlakeShell is the devShell attribute used when none is requested.
   const DefaultFlakeShell = "default"
   ```

3. **Create the executor skeleton.** New file `packages/agent/agent-cli-go/internal/sandbox/nixflake/nixflake.go`. `IsAvailable` mirrors `nix.IsAvailable` (`internal/sandbox/nix/nix.go:26-46`) but checks `nix` (not `nix-build`) plus flake support, `bwrap`, and `/nix/store`.

   ```go
   package nixflake

   import (
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

   // Executor enters a project flake devShell via `nix develop ... --command <agent>`.
   type Executor struct{}

   func NewExecutor() *Executor { return &Executor{} }

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
   ```

4. **Parse the flake config with a flake-dir default (deny-default friendly).** In `nixflake.go`, parse `config.Image` of form `nixflake:{json}` or `nixflake:ref=...,shell=...`, mirroring `nix.parseNixConfig` (`internal/sandbox/nix/nix.go:157-211`). When no ref is given, default to the bound project flake dir (`config.RepoDir`, falling back to `config.WorkDir`), so the in-sandbox path is what gets evaluated and `flake.lock` there pins the closure.

   ```go
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
   ```

5. **Build bwrap args (no `/env`, flake dir ro-bound, daemon socket).** Adapt `nix.buildBwrapArgs` (`internal/sandbox/nix/nix.go:246-333`) but drop the `--ro-bind envOutPath /env` mount and instead ro-bind the resolved `nix` binary's store dir, the flake dir, and the daemon socket so the sandboxed `nix develop` can talk to the host store. Deny-default `PATH` = the nix bin dir only (no `/usr/bin`, no `/bin`).

   ```go
   func (e *Executor) buildBwrapArgs(config *types.SandboxConfig, fc nixenv.FlakeSandboxConfig, dirs *AgentDirs, nixBinDir string) []string {
       homeDir, _ := os.UserHomeDir()
       e.ensureSandboxMountPoint(dirs.Home, config.WorkDir)

       args := []string{
           "--ro-bind", "/nix/store", "/nix/store",
           "--ro-bind", nixBinDir, nixBinDir, // the `nix` CLI closure
           "--bind", dirs.Work, "/work",
           "--bind", dirs.Tmp, "/tmp",
           "--bind", dirs.Home, homeDir,
       }
       // Daemon socket + db so `nix develop` resolves the flake against the host store.
       if _, err := os.Stat("/nix/var/nix/daemon-socket/socket"); err == nil {
           args = append(args, "--ro-bind", "/nix/var/nix/daemon-socket/socket", "/nix/var/nix/daemon-socket/socket")
       }
       // Flake dir read-only so flake.lock pins the closure; --no-write-lock-file keeps it ro.
       args = append(args, "--ro-bind", fc.FlakeRef, fc.FlakeRef)

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
           "PATH":              nixBinDir, // deny-default: nix devShell prepends its own bin
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
       args = append(args, "--chdir", config.WorkDir, "--die-with-parent")
       return args
   }
   ```
   Note: copy the small `AgentDirs`, `ensureAgentDirs`, and `ensureSandboxMountPoint` helpers from `internal/sandbox/nix/nix.go:127-154,357-374` into the new package (the platform forbids re-export shims, so duplicate the tiny helpers rather than export them).

6. **Assemble the `nix develop` command and Execute (with GPU/display scope guard).** Resolve the real `nix` store path, build `nix develop --no-write-lock-file <ref>#<shell> --command <agentCmd...>`, and run it under bwrap via `sandboxutil.SpawnSandbox`. A pure flake devShell hides host GPU/display libs (no `/usr/lib` graphics stack is bound), so guard to CPU/tooling agents and refuse when the agent declares a GPU/display need.

   ```go
   func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
       if config.Agent != nil && config.Agent.NeedsGPU {
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
       nixBinDir := filepath.Dir(realNix)

       dirs := e.ensureAgentDirs(config.AgentID)
       bwrapArgs := e.buildBwrapArgs(config, fc, dirs, nixBinDir)

       // Deny-default: agent binary resolved by the devShell PATH (no /env prefix).
       agentCmd := sandboxutil.BuildAgentCommand(config, "")
       fullCmd := sandboxutil.WrapWithInitCommands(agentCmd, config.SandboxInitCommands)

       develop := []string{
           filepath.Join(nixBinDir, "nix"), "develop", "--no-write-lock-file",
           fc.FlakeRef + "#" + fc.Shell, "--command",
       }
       develop = append(develop, fullCmd...)

       bwrapArgs = append(bwrapArgs, "--")
       bwrapArgs = append(bwrapArgs, develop...)

       if config.Verbose {
           scriptlib.LogInfo("Entering flake devShell " + fc.FlakeRef + "#" + fc.Shell)
       }

       agentEnv, _ := sandboxutil.BuildProviderEnv(config)
       merged := sandboxutil.MergeEnvMaps(agentEnv, sandboxutil.ParseCustomEnv(config.CustomEnv))
       env := append(os.Environ(), sandboxutil.EnvMapToSlice(merged)...)
       return sandboxutil.SpawnSandbox("bwrap", bwrapArgs, env, "Nix flake sandbox", config.Verbose)
   }

   func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
       fc := e.parseFlakeConfig(config)
       return []string{
           "bwrap --ro-bind /nix/store /nix/store --ro-bind " + fc.FlakeRef + " " + fc.FlakeRef + " ... -- \\",
           "  nix develop --no-write-lock-file " + fc.FlakeRef + "#" + fc.Shell + " --command <agent>",
       }
   }
   ```
   Note: `config.Agent.NeedsGPU` does not exist yet on `AgentConfig`; if absent, gate on a simpler heuristic (e.g. agent name allowlist) and leave a one-line comment documenting the CPU/tooling scope. Keep it explicit error handling per house style.

7. **Add Go tests.** New file `packages/agent/agent-cli-go/internal/sandbox/nixflake/nixflake_test.go`. Cover: parse defaults, flake-dir fallback, and the deny-default `PATH`/ro-bind invariants on the bwrap args.

   ```go
   package nixflake

   import (
       "strings"
       "testing"

       "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
   )

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

   func TestBwrapArgsDenyDefaultPath(t *testing.T) {
       e := NewExecutor()
       fc := nixenv.FlakeSandboxConfig{FlakeRef: "/repo", Shell: "default"}
       dirs := e.ensureAgentDirs("test")
       args := e.buildBwrapArgs(&types.SandboxConfig{WorkDir: "/work"}, fc, dirs, "/nix/store/x/bin")
       joined := strings.Join(args, " ")
       if strings.Contains(joined, "--setenv PATH /usr/bin") || strings.Contains(joined, ":/usr/bin") {
           t.Fatal("host /usr/bin leaked into sandbox PATH")
       }
       if !strings.Contains(joined, "--ro-bind /repo /repo") {
           t.Fatal("flake dir not ro-bound")
       }
   }
   ```
   Note: also add a registry test asserting `GetExecutor(types.SandboxNixFlake)` returns a non-nil executor with no error.

## Validation Steps

```bash
# from repo root: /home/mvierssen/Projects/xonovex/xonovex-platform
npx moon run agent-cli-go:go-typecheck
npx moon run agent-cli-go:go-lint
npx moon run agent-cli-go:go-build
npx moon run agent-cli-go:go-test
npx moon run agent-cli-go:ci-check   # aggregates go-build/test/lint/typecheck + TS checks

# Integration smoke (requires nix + bwrap + a project flake.nix with devShells.default):
cd /home/mvierssen/Projects/xonovex/xonovex-platform
go run ./packages/agent/agent-cli-go --sandbox nixflake \
  --image 'nixflake:shell=default' -- echo flake-ok   # expect: runs inside `nix develop`, prints flake-ok
```

## Success Criteria

- [ ] `types.SandboxNixFlake` exists and `GetExecutor`/`GetAvailableMethods` handle it.
- [ ] `internal/sandbox/nixflake/nixflake.go` enters `nix develop <ref>#<shell> --command <agent>` under bwrap, with `/nix/store` and the flake dir ro-bound and `--no-write-lock-file` set.
- [ ] Sandbox `PATH` contains only the nix bin dir — no host `/usr/bin` or `/bin` (deny-default verified by test).
- [ ] Flake ref defaults to `config.RepoDir` then `config.WorkDir`; shell defaults to `default`.
- [ ] GPU/display agents are explicitly refused with a clear error; CPU/tooling scope documented in a comment.
- [ ] `go-typecheck`, `go-lint`, `go-build`, `go-test`, and `ci-check` all pass.

## Files Modified/Created

- `packages/shared/shared-agent-go/pkg/types/sandbox.go` — add `SandboxNixFlake` constant.
- `packages/agent/agent-cli-go/internal/sandbox/registry.go` — register the `nixflake` executor + add to `allMethods`.
- `packages/agent/agent-cli-go/internal/nixenv/types.go` — add `FlakeSandboxConfig` + `DefaultFlakeShell`.
- `packages/agent/agent-cli-go/internal/sandbox/nixflake/nixflake.go` — new executor (created).
- `packages/agent/agent-cli-go/internal/sandbox/nixflake/nixflake_test.go` — new tests (created).

## Dependencies

None. This subplan is self-contained on the Go sandbox side and shares no files with the Rust plugin subplans, so it runs fully in parallel within `parallel_group: 3`. It reads (but does not modify) `internal/sandbox/nix/nix.go` and `internal/sandboxutil/utils.go` as the pattern source.

## Estimated Duration

~0.5-1 day: ~2-3h for the executor + config + registry wiring, ~1-2h for tests, ~1h for the nix-develop/daemon-socket integration smoke and deny-default verification.
