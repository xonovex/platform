---
type: plan
has_subplans: false
parent_plan: plans/agent-security-hardening.md
parallel_group: 5
status: pending
dependencies:
  plans:
    - 06-cli-host-reach-and-pinning
  files:
    - packages/agent/agent-cli-go/internal/cmd/**
    - packages/agent/agent-cli-go/internal/config/**
    - packages/agent/agent-cli-go/internal/terminal/**
    - packages/agent/agent-cli-go-*/moon.yml
skills_to_consult:
  - code-quality-guide
  - fp-guide
  - moon-guide
  - testing-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 07 — CLI: Version Injection, Architecture Fence, Dead Code

## Objective

Make released binaries report their real version, close the composition-root
loophole the architecture test misses, remove dead config/flag/terminal
surface, and deduplicate small utilities.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `a6f765e2` (2026-07-02) and
will drift as earlier subplans land — always read the cited file before
editing; if a line reference doesn't match, locate the named construct
instead.

`packages/agent/agent-cli-go` (module
`github.com/xonovex/platform/packages/cli/agent-cli-go`) is published to npm
as platform-specific binary packages: `packages/agent/agent-cli-go-<platform>`
(linux-x64, linux-arm64, darwin-x64, darwin-arm64, win32-x64), each with a
`moon.yml` that builds the Go binary. The release line is currently 0.1.31
(see `agent-cli-go/package.json` and `CHANGELOG.md`).

The sandbox core follows a strict plugin architecture:
`internal/sandbox/plugins/plugins.go` `DefaultRegistry()` is documented
(`packages/agent/AGENTS.md`) as the ONLY place that imports concrete plugin
packages (`internal/isolation/{bwrap,docker,none}`,
`internal/provision/nix`). `internal/sandbox/architecture_test.go` is the
fitness test for this.

Current defects, verified against source:

1. **Version misreport.** `internal/cmd/root.go:7` hardcodes
   `var version = "0.1.0"`. The `go-build` task script in each
   `packages/agent/agent-cli-go-*/moon.yml` passes only
   `-ldflags="-s -w"` — no `-X` injection — so every released binary
   (0.1.31) reports 0.1.0. `test/integration/run_test.go:60` asserts the
   stale `0.1.0`, locking the bug in.
2. **Composition-root drift.** `internal/cmd/run.go:15` imports the concrete
   plugin package `internal/provision/nix` (aliased `provnix`) just for
   `SourceFromFlags`. `architecture_test.go:172` only bans plugin
   CONSTRUCTORS outside `plugins.go`, so concrete-package imports drift in
   unnoticed.
3. **Dead config/flag/terminal surface** (repo rule: remove unused code
   immediately, no compat shims):
   - `internal/config/loader.go:16-22`: `FileConfig.HomeDir`, `BindPaths`,
     `RoBindPaths`, `CustomEnv` are parsed but never consumed — only
     `.Provider` is read (`run.go:223`). User config is silently ignored.
   - `LoadDefaultConfig` (`loader.go:104`) has zero callers.
   - Default config path still named `sandboxed-claude` (`loader.go:27`) —
     stale product name.
   - `--debug` persistent flag (`root.go:24`) is never read.
   - `TerminalConfig.AttachExisting` and `TerminalExecutor.IsInside` are
     unused.
4. **Duplication.** `buildShellCommand` exists twice
   (`internal/isolation/shared/spawn.go:45` and
   `internal/terminal/tmux/tmux.go:270`); the identical 4-map env merge in
   bwrap's `sandboxEnv` and docker's `containerEnv` reimplements the existing,
   unused `envutil.MergeEnvMaps`.
5. **bwrap PATH tail is dead.** The bwrap devtmpfs setup lacks
   `--symlink usr/bin /bin` (and siblings), so the `/usr/local/bin:/usr/bin:
   /bin` PATH tail resolves nothing in deny-default mode unless the nix
   closure provides everything.

## Tasks

1. **Version injection.**
   - Determine how the platform `moon.yml` build tasks obtain the package
     version (read one of `packages/agent/agent-cli-go-*/moon.yml` and any
     scripts it calls; the version lives in each platform package's
     `package.json`, kept in lockstep by the release flow).
   - Add `-X github.com/xonovex/platform/packages/cli/agent-cli-go/internal/cmd.version=<version>`
     to the existing `-ldflags` in ALL five platform `go-build` scripts, sourcing
     `<version>` from the package.json (moon tasks can shell out; follow the
     repo's existing script conventions under `packages/script/` if a helper
     fits).
   - Keep the `var version` default meaningful for dev builds (e.g. `dev`).
   - Fix `test/integration/run_test.go:60` to assert the dev default or the
     injected value — not the stale literal.
2. **Tighten the architecture fence.**
   - Move `SourceFromFlags` out of `internal/provision/nix` into the flag
     layer (`internal/cmd`) or a neutral shared package
     (`internal/provision/shared`), so `run.go` no longer imports the
     concrete plugin. Check what `SourceFromFlags` returns first — if the
     return type is defined in the nix package, relocate the type to the
     shared/neutral package too (no re-export shims; move it outright).
   - Extend `architecture_test.go` to ban IMPORTS of the concrete plugin
     packages (`isolation/bwrap`, `isolation/docker`, `isolation/none`,
     `provision/nix`) anywhere outside `internal/sandbox/plugins/`, not just
     constructor calls. The test must fail if run.go's import returns.
3. **Delete dead surface.** Each bullet from Context item 3:
   - Wire nothing "for later": DELETE `HomeDir`/`BindPaths`/`RoBindPaths`/
     `CustomEnv` from `FileConfig` unless subplan 06 wired them (check the
     merged state of `internal/config` first; 06's `--mount-user-config` may
     have consumed some fields — do not delete what 06 uses).
   - Delete `LoadDefaultConfig`, `AttachExisting`, `IsInside`, `--debug`.
   - Rename the default config path `sandboxed-claude` → the current product
     name (match the binary/npm name `agent-cli-go` conventions; pick the
     name used elsewhere in docs). No fallback to the old path — repo rule
     forbids compat shims; note the break in the commit message.
4. **Deduplicate.** Single `buildShellCommand` in one owning package (the
   isolation/shared spawn location is the natural owner; tmux imports it);
   replace both hand-rolled env merges with `envutil.MergeEnvMaps` (delete it
   if it turns out unsuitable — one owner either way).
5. **bwrap PATH.** Either add the `--symlink usr/bin /bin`-style links so the
   documented PATH tail resolves, or trim the PATH to what actually exists in
   deny-default mode. Decide by reading how the nix closure provisions PATH
   (`internal/provision/nix` + `nix/agent-env.nix`); keep behavior and PATH
   consistent and tested.

## Validation Steps

Prerequisites: Go toolchain is nix-managed — if `go` is missing, run inside
`nix develop`. No other host tooling needed (integration tests here don't
require docker).

```bash
npx moon run agent-cli-go:go-build
npx moon run agent-cli-go:go-lint
npx moon run agent-cli-go:go-test
cd packages/agent/agent-cli-go && go test -tags=integration ./test/integration/
# version smoke: build one platform artifact and check --version
npx moon run agent-cli-go-linux-x64:go-build
packages/agent/agent-cli-go-linux-x64/bin/agent-cli-go --version
```

## Success Criteria

- [ ] A platform-built binary reports the package.json version; dev builds
      report the dev default; integration test updated accordingly.
- [ ] `run.go` imports no concrete plugin package; `architecture_test.go`
      fails on any such import outside `plugins/` (verified by temporarily
      re-adding one).
- [ ] `rg "LoadDefaultConfig|AttachExisting|IsInside|sandboxed-claude"` in the
      CLI returns nothing (modulo the renamed path constant).
- [ ] One `buildShellCommand`, one env-merge implementation.
- [ ] bwrap PATH entries all resolve (or are removed).
- [ ] All validation commands pass.

## Files Modified/Created

- Modified: `internal/cmd/root.go`, `internal/cmd/run.go`,
  `internal/config/loader.go`, `internal/terminal/**`,
  `internal/isolation/shared/spawn.go`, `internal/terminal/tmux/tmux.go`,
  `internal/isolation/bwrap/bwrap.go`, `internal/sandbox/architecture_test.go`,
  `test/integration/run_test.go`, `packages/agent/agent-cli-go-*/moon.yml`
  (all five).
- Moved: `SourceFromFlags` (+ its types) out of `internal/provision/nix`.

## Dependencies

Depends on `06-cli-host-reach-and-pinning` (same cmd/isolation files; 06 may
consume config fields this subplan would otherwise delete).

## Estimated Duration

1 day.
