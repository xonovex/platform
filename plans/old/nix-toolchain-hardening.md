---
type: plan
has_subplans: true
status: pending-approval
feature: nix-toolchain-hardening
dependencies:
  plans: []
  subplans:
    - plans/nix-toolchain-hardening/01-plugin-typed-config.md
    - plans/nix-toolchain-hardening/02-plugin-fail-closed.md
    - plans/nix-toolchain-hardening/03-plugin-flake-shell-routing.md
    - plans/nix-toolchain-hardening/04-plugin-cache-coherence.md
    - plans/nix-toolchain-hardening/05-sandbox-flake-devshell.md
    - plans/nix-toolchain-hardening/06-sandbox-deny-default.md
    - plans/nix-toolchain-hardening/07-rollout-and-release.md
parallel_groups:
  - group: 1
    plans: [plugin-typed-config]
    note: "Foundation: replace the untyped serde_json::Value config reads with a typed define_toolchain_config schema. Every later plugin change reads the new config struct, so it lands first."
  - group: 2
    plans: [plugin-fail-closed, plugin-flake-shell-routing, plugin-cache-coherence]
    depends_on: [1]
    note: "The three plugin features. Logically independent, but all edit src/lib.rs (resolve_wrap_target / resolve_shell / the hook fns) — so they SERIALIZE on that one file: land them one after another (any order), not in parallel worktrees."
  - group: 3
    plans: [sandbox-flake-devshell, sandbox-deny-default]
    note: "Independent Go track (agent-cli-go + shared-agent-go) — no dependency on the plugin, runs in parallel with groups 1-2. Internally sequential: deny-default builds on the flake-devShell executor."
  - group: 4
    plans: [rollout-and-release]
    depends_on: [1, 2]
    note: "Cross-repo: cut moon_nix_toolchain-v0.6.0 (wasm + GitHub release), bump the @<tag> in BOTH consumers' .moon/toolchains.yml (xonovex + drodan), and enable the new config. Lands after the plugin work; the sandbox track (3) can ship on its own cadence."
proposed_subplans:
  - plugin-typed-config
  - plugin-fail-closed
  - plugin-flake-shell-routing
  - plugin-cache-coherence
  - sandbox-flake-devshell
  - sandbox-deny-default
  - rollout-and-release
skills_to_consult:
  - moon-guide              # toolchains.yml config + the plugin's moon integration surface
  - general-fp-guide        # pure functions / explicit context, applied to both the Rust plugin and Go sandbox
  - code-review-guide       # xonovex ships via PRs (main is protected) — Conventional Comments
  - pull-request-guide      # sizing/splitting + how-tested, per the protected-main workflow
  - git-guide               # conventional commits; cross-repo coordination
  - debugging-guide         # determinism + "did the wrap/guard actually engage" verification
research_sources:
  documentation:
    - packages/moon/moon-nix-toolchain/src/lib.rs                      # resolve_wrap_target no-op guards (44-57), project_flake short-circuit (185-194 / 224-233), resolve_shell precedence (100-157), flake_ref (161-166)
    - packages/moon/moon-nix-toolchain/CHANGELOG.md                    # 0.5.0 = the project-flake wrap (the short-circuit this plan relaxes)
    - packages/moon/moon-nix-toolchain/Cargo.toml                      # moon_pdk_api 2.0.4, edition 2021
    - packages/moon/moon-nix-toolchain/tests/wrap_test.rs              # the wrap assertions (project flake currently asserts no `#shell`)
    - packages/agent/agent-cli-go/internal/sandbox/registry.go         # the SandboxExecutor registry seam (where a flake-devShell tier is added)
    - packages/agent/agent-cli-go/internal/sandbox/nix/nix.go          # nix tier: ro-binds /nix/store + /env, PATH=/env/bin (deny-default)
    - packages/agent/agent-cli-go/internal/sandbox/bwrap/bwrap.go      # bwrap tier LEAKS host /usr,/lib,/bin (the gap deny-default closes)
    - packages/agent/agent-cli-go/internal/nixenv/render.go            # synthesizes a flat pkgs.buildEnv from a package list (not a flake devShell)
    - packages/agent/agent-cli-go/internal/nixenv/build.go             # nix-build of the buildEnv; channel-pinned (fetchTarball), not flake.lock
    - packages/agent/agent-cli-go/internal/sandboxutil/utils.go        # SpawnSandbox launches the agent once (subprocesses inherit the namespace)
    - packages/shared/shared-agent-go/pkg/nix/nix.go                   # shared nix helpers
    - .moon/toolchains.yml                                             # xonovex consumer config (shellByTag go/shell/rust/moon-plugin)
    - ../../drodan/drodan-platform/.moon/toolchains.yml                # drodan consumer (shellByTag cmake:cc; the fail-closed + per-task targets)
  versions:
    moon_pdk_api: "2.0.4"
    moon_pdk: "2.0.4"
    rust_edition: "2021"
    plugin_current: "moon_nix_toolchain-v0.5.0"
    moon: "v2.0 (Phobos) — WASM toolchain plugins stabilized"
  design:
    - "Diagnosed in the drodan nix-boundary investigation (2026-06): the plugin silently no-ops when nix is absent (lib.rs:55-57), flake-owning projects cannot route per-task (project_flake short-circuit), flake/lock edits do not bust moon's task cache, and the agent sandbox's nix method cannot enter a flake.lock-pinned devShell."
    - "Decided defaults from that investigation: full-nix is the default game build (drodan da6c..002f), CC pinned to clang in nix/cc.nix, lean cc shell routed via shellByTag cmake:cc, fail-closed scoped to game/C."
    - "tier2 hooks available in moon_pdk_api 2.0.4: hash_task_contents, setup_environment, define_toolchain_config — none implemented in 0.5.0."
---

# Nix Toolchain Plugin + Agent Sandbox Hardening

## Overview

`moon_nix_toolchain` (the WASM moon plugin both Xonovex and Drodan consume) wraps every opted-in task in `nix develop`, but it enforces nothing, can't route per-task inside flake-owning projects, and never invalidates moon's cache when a flake changes; the agent sandbox can confine an agent to a nix environment but can't consume a project's `flake.lock`-pinned devShell. This plan hardens both — the plugin and the agent sandbox — so the nix toolchain boundary becomes enforced, granular, cache-correct, and usable by the agent sandbox.

## Goals

- Make the plugin **fail-closed** for opted-in tasks (error when nix is absent) instead of silently falling back to host tools.
- Let **flake-owning projects route tasks to named devShells** (per-task granularity inside a project flake).
- Make a `flake.nix`/`flake.lock`/devShell edit **bust moon's task cache** (cache coherence).
- Give the plugin a **typed, validated config** instead of untyped JSON reads.
- Teach the **agent sandbox** to enter a **flake devShell** (`flake.lock`-pinned), not just a synthesized flat package set.
- Make host tools **physically unreachable** for pinned agent runs (deny-default), not merely "not on PATH".
- Roll a single **versioned release** out to both consumer repos with the new config enabled.

## Current State

**Plugin** (`packages/moon/moon-nix-toolchain`, Rust→WASM, v0.5.0 on `moon_pdk_api 2.0.4`):
- Three hooks only: `register_toolchain`, `extend_task_command`, `extend_task_script`; one host import `load_project_by_id`.
- Config is read as an **untyped `serde_json::Value`** — no schema, no validation (`resolve_shell`, `lib.rs:100-157`).
- `resolve_wrap_target` (`lib.rs:40-82`) returns `None` (run unchanged) when `IN_NIX_SHELL` is set, `MOON_NIX_WRAPPED=1`, **or `nix` is absent** (`55-57`) — i.e. it **silently uses host tools** when nix is missing.
- Shell selector precedence works (`shellByTask > shellByToolchain > shellByTag > shellByLanguage > shell`) — but a project that ships its own `flake.nix` hits the **`project_flake` short-circuit** (`185-194` command, `224-233` script): `shell = None`, so it always uses the project flake's bare `default` and **ignores every selector**.
- **No** `hash_task_contents`, `setup_environment`, or `define_toolchain_config` — so a flake edit does not invalidate cached task outputs.

**Agent sandbox** (`packages/agent/agent-cli-go`, `packages/shared/shared-agent-go`, Go):
- Five real tiers behind one `SandboxExecutor` interface: `none`, `bwrap`, `docker`, `compose`, `nix` (`internal/sandbox/{tier}/`, registry at `internal/sandbox/registry.go`).
- The **nix tier** synthesizes a flat `pkgs.buildEnv` from a package list (`internal/nixenv/render.go`), `nix-build`s it (`build.go`), and ro-binds `/nix/store` + the closure with `PATH=/env/bin` (`internal/sandbox/nix/nix.go`) — **deny-default, but channel-pinned (`fetchTarball`), not `flake.lock`-pinned, and cannot consume a project's flake devShell**.
- The **bwrap tier leaks** host `/usr,/lib,/bin` + host PATH (`internal/sandbox/bwrap/bwrap.go`) — so "host tools unreachable" holds only under the nix tier.
- `SpawnSandbox` launches the agent once; subprocesses inherit the namespace (`internal/sandboxutil/utils.go`).

**Consumers**: both repos pin `…/moon_nix_toolchain@moon_nix_toolchain-v0.5.0` in `.moon/toolchains.yml`. Drodan now routes game C tasks via `shellByTag: {cmake: cc}` and pins `CC=clang`; it is the first repo that wants fail-closed + per-task routing.

## Research Findings

- **`moon_pdk_api 2.0.4` exposes the tier2 hooks** this plan needs — `define_toolchain_config` (typed config + schema), `hash_task_contents` (fold inputs into moon's cache key), `setup_environment` (pre-build/GC-root). The plugin already depends on 2.0.4; no version bump is required to adopt them.
- **`shellByTask` already routes per-task for non-flake projects** — the only gap is the `project_flake` short-circuit. Relaxing it (resolve the selector against the *project* flake and emit `{projectRoot}#{shell}`) is a localized change, gated on the project flake exposing named shells.
- **`extend_task_command` is the right enforcement seam** — it runs for every moon-executed task (humans, `moon run`, CI, agents alike). Returning an error there is how fail-closed surfaces. The `IN_NIX_SHELL`/`MOON_NIX_WRAPPED` no-ops are correct double-entry guards and must stay.
- **The sandbox already proves deny-default works** under the nix tier; the missing capability is *which* environment it enters. A `nix develop <ref>#<shell> --command` executor reuses the existing bind/namespace machinery while swapping the env source to a `flake.lock`-pinned devShell.

## Proposed Approach

1. **Typed config** (`define_toolchain_config`): a `NixToolchainConfig` struct (`shell`, `shellByTask`, `shellByToolchain`, `shellByTag`, `shellByLanguage`, plus new `requireNix`/`failClosed`), replacing every `config.get("…")` read in `resolve_shell`. Foundation for the rest.
2. **Fail-closed** (`lib.rs` `resolve_wrap_target`/hooks): when the task is opted-in (per-tag/per-language `failClosed`, e.g. `cmake`/`c`) and `nix` is absent, return an error instead of `None`. Keep the `IN_NIX_SHELL`/`SENTINEL` no-ops.
3. **Flake-project shell routing** (`lib.rs:185-194`/`224-233`): when `project_flake`, still call `resolve_shell` and emit `{projectRoot}#{shell}` (or a `projectShellBy*` namespace). Requires the project flake to expose the named shells; update `tests/wrap_test.rs`.
4. **Cache coherence** (`hash_task_contents` + optional `setup_environment`): fold the resolved devShell's `flake.lock` hash / store-path into moon's task cache key; optionally pre-build + GC-root the shell.
5. **Sandbox flake-devShell executor** (`internal/sandbox/registry.go` + a new `internal/sandbox/nixflake/` or `nixenv` mode): enter `nix develop <flakeRef>#<shell> --command`, reusing the existing bind/namespace setup; `flake.lock`-pinned.
6. **Sandbox deny-default policy**: a config/flag that mandates the nix(flake) tier and rejects the leaky `bwrap`/`none` tiers when a pinned toolchain is required, so host tools are physically unreachable.
7. **Release + rollout**: cut `moon_nix_toolchain-v0.6.0` (wasm build + GitHub release asset), bump both `.moon/toolchains.yml` consumers, and turn on `failClosed`/per-task routing in drodan.

## Risk Assessment

- **One plugin, two consumers.** A behavior change ships to Xonovex *and* Drodan at once. Mitigate: additive/opt-in config (default off), semver the tag, bump both `toolchains.yml` in lockstep, keep the no-op guards.
- **Fail-closed could break nix-less environments.** Mitigate: opt-in per tag/language only; never global; `IN_NIX_SHELL` still no-ops under CI's outer shell.
- **Flake-project routing needs named shells.** Drodan's game project flakes currently expose only `default` (a `cc` re-export); per-task routing for them needs added named shells, *or* they keep using the workspace `#cc` via `shellByTag` (already working) — so this is the lowest-urgency plugin item. Sequence it last in group 2.
- **Cache key correctness.** A wrong `hash_task_contents` key causes silent stale hits or thrashing misses; cover with `moon_pdk_test_utils` cases for "flake edit busts cache" and "unrelated edit does not".
- **Sandbox flake devShell + GPU/display.** A pure flake devShell can hide host GPU/display libs (the headless-vulkan-loader / nixGL issue from the drodan investigation). Scope the sandbox executor to CPU/tooling agents first; document the GPU caveat.
- **Protected `main` on xonovex.** Land via PRs (GitHub ruleset requires it) — unlike drodan's trunk-based gitlab flow.

## Proposed Child Plans

**Group 1 — foundation:** `01-plugin-typed-config`
**Group 2 — plugin features (serialize on `src/lib.rs`), after group 1:** `02-plugin-fail-closed`, `03-plugin-flake-shell-routing`, `04-plugin-cache-coherence`
**Group 3 — sandbox track (parallel, independent Go):** `05-sandbox-flake-devshell` → `06-sandbox-deny-default`
**Group 4 — release + rollout, after groups 1-2:** `07-rollout-and-release`

## Success Criteria

- An opted-in task with `nix` unavailable **fails with a clear error** (not a host-tool fallback); non-opted tasks and `IN_NIX_SHELL`/CI runs are unchanged.
- A flake-owning project can route a task to a named devShell (`{projectRoot}#<shell>`); `wrap_test.rs` covers it.
- Editing `flake.nix`/`flake.lock`/a shell **invalidates** the dependent tasks' moon cache; an unrelated edit does not.
- Plugin config is typed + validated (bad keys rejected at load).
- The agent sandbox can run an agent inside a project's `flake.lock`-pinned devShell with host tools **unreachable**.
- `moon_nix_toolchain-v0.6.0` released; both consumers green on the new tag; drodan game/C is fail-closed.

## Estimated Effort

- Plugin (groups 1-2): ~3-4 days incl. `moon_pdk_test_utils` tests.
- Sandbox (group 3): ~2-3 days incl. Go tests.
- Release + rollout (group 4): ~0.5 day.
- **Total: ~6-8 days**, parallelizable across the plugin and sandbox tracks (~4-5 days wall-clock with two workstreams).
