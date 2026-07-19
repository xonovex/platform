---
type: plan
has_subplans: false
parent_plan: plans/moon-nix-extension.md
parallel_group: 3
status: complete
updated: 2026-07-19
completed_date: "2026-07-19"
dependencies:
  plans:
    - 02-global-extension-adapter
  files:
    - packages/moon/moon-nix-extension/moon.yml
    - packages/moon/moon-nix-extension/README.md
    - packages/moon/moon-nix-extension/tests/**
    - packages/moon/moon-nix-extension/src/lib.rs
    - packages/moon/moon-nix-runtime/src/installable.rs
    - packages/moon/moon-nix-runtime/tests/installable_test.rs
skills_to_consult:
  - moon-guide
  - testing-guide
  - shell-scripting-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  integration: passed
---

# 03 — Prove the cache/adoption contract in a standalone consumer workspace

## Objective

Build a black-box nested Moon workspace that consumes only the local extension
WASM, uses native toolchain detection with Proto installation disabled, and
proves composition, scoped overrides, locked project environments, failures,
and cache invalidation through real Moon 2.3.5 and Nix 2.34 executions.

## Context (read this first — no other context is assumed)

All code paths and commands resolve from the Xonovex repository root. Subplans
01 and 02 must be merged and green. Anchors were verified from the
parent baseline `80b3d773dde1e1dee516938096e9022cceccda0a`; this child mostly
creates new fixture files under the extension package.

This subplan is the boundary between unit confidence and the actual consumer
contract:

- PDK unit tests cannot prove Moon calls every configured extension globally,
  performs native toolchain detection first, merges extension outputs in the
  verified order, or hashes declared task inputs.
- Moon 2.3.5 detects a direct `node` command as `[javascript, npm, node]`; it
  does not add `system`. A script whose executable does not match a toolchain
  falls back to all project toolchains, so a project containing both
  `package.json` and `go.mod` can exercise a real polyglot set. The fixture must
  assert the toolchain list reported by Moon before asserting Nix components.
- Moon's cache hash includes task inputs, command/args/env/dependencies, and
  native ecosystem data, but an extension has no `hash_task_contents` hook.
  The consumer must declare central flake/config inputs and each selected
  project flake/lock/Nix source. Do not claim the extension discovers these
  automatically.
- `.moon/toolchains.yml` enables native integration with `version: null`.
  Official Moon documentation confirms a configured non-null version is what
  triggers automatic Proto download/install. The fixture must contain no `nix`
  toolchain block or `moon_nix_toolchain` locator.
- The root Xonovex workspace currently still uses the toolchain plugin. Keep the
  fixture as a nested, isolated Moon workspace and invoke the repository-pinned
  Moon binary explicitly so outer configuration cannot leak into results.

The fixture should use lightweight `writeShellScriptBin` marker executables
instead of realizing production Node/Go closures. It tests environment
composition and ownership, not language behavior.

## Tasks

1. **Create an isolated nested consumer fixture.** Add
   `packages/moon/moon-nix-extension/tests/fixtures/consumer-workspace/` with
   its own `.moon/workspace.yml`, `.moon/toolchains.yml`,
   `.moon/extensions.yml`, inherited task inputs, root `flake.nix`/`flake.lock`,
   component modules, and projects under `projects/`. Point the extension at
   the just-built local `moon_nix_extension.wasm` through a `file://` locator
   resolved or templated by the runner. Invoke the root repository's pinned
   Moon 2.3.5 binary from the fixture; do not let nested `npx` install another
   Moon release.

2. **Model the central and project-owned Nix environments cheaply.** The root
   flake must expose `lib.mkMoonShell` and components for `general`, `node`,
   `go`, `node20`, `node24`, `postgresql`, and `protobuf`. Each component should
   expose a uniquely named/printing executable or environment marker so tests
   can prove exactly which components were composed. Add a `special-compiler`
   project with its own `flake.nix`, `flake.lock`, and named devShells `moon`
   and `bootstrap`. Keep all inputs pinned and local to the fixture after the
   initial lock fetch.

3. **Prove real native-toolchain detection and Proto opt-out.** Configure
   JavaScript, Node, npm, and Go natively with the versioned tools set to
   `null`; omit `nix`. Add:

   - a Node command task with no explicit `toolchains`, expected by Moon 2.3.5
     to report `[javascript, npm, node]`;
   - a project containing `package.json` and `go.mod` with a script task that
     falls back to its detected project toolchains, expected to include Go and
     the Node/JavaScript/package-manager set;
   - an unmapped system task and scoped-only task to distinguish no-op from
     override activation.

   Assert these lists through `moon task`/query output before running the tasks.
   Capture the action graph or trace output and prove no Proto Node/Go install
   action occurs while Nix provides the marker executables.

4. **Exercise the complete resolution and failure matrix.** Configure
   `baseComponents: [general]`, deduplicating Node mappings, Go, project append
   and replace, task append and replace, project installable, task installable,
   and task supersession of a project installable. Assert exact markers for
   command and opaque script tasks. Add negative cases for inline/remote
   installables, mixed union shapes, parent/symlink path escapes, absent
   devShell attribute, append-over-installable, invalid component name, missing
   component, missing `lib.mkMoonShell`, missing Nix with both fail policies,
   and a deliberate Nix evaluation error. Every failure must be non-zero and
   name the target and rejected value/reason without exposing a generated raw
   expression.

5. **Make cache ownership explicit and test it by mutation.** Add inherited
   implicit inputs for the fixture's central `flake.nix`, `flake.lock`, Nix
   component sources, and `.moon/extensions.yml`. Add the selected project
   flake, lock, and Nix sources to tasks using that installable. Give fixture
   tasks narrow ordinary inputs/outputs so an unrelated file is truly outside
   their hash. In a disposable fixture copy:

   - run a cacheable task twice and prove the second result is cached;
   - change each central lock/shell/config input separately and prove a miss;
   - change each selected project flake/lock/source and prove only its consumers
     miss;
   - change an unrelated file and prove the hash/result remains cached;
   - run an uncached task and prove it always executes.

   Prefer `moon hash` plus run-report evidence over timing as the assertion.
   Restore/mutate only the temporary copy, never tracked fixture files.

6. **Prove mutually exclusive standalone deployment.** Run the fixture with
   isolated Moon/proto state and only the extension WASM staged; assert its
   toolchain config contains no `nix` entry and trace output never loads
   `moon_nix_toolchain`. Add a static fixture/config assertion that fails if a
   Nix toolchain locator or selection is introduced. Exercise nested extension
   re-entry through its own sentinel, and add an environment-only peer extension
   case if the test harness makes it cheap. Do not configure both Nix plugins or
   claim support for an arbitrary peer command replacement.

7. **Wire integration into the extension project's gates and document the
   contract.** Add a deterministic integration runner under `tests/` and an
   `integration` task in the extension's `moon.yml` that depends on the release
   WASM build. Use `set -euo pipefail`, a validated `mktemp -d`, and a cleanup
   trap; print the retained temp path on failure for diagnosis. Include fixture
   files and the runner in task inputs. Update the extension README with the
   exact consumer input pattern, `version: null` rule, mapping-gated base
   semantics, project/task precedence, project-flake requirements, cache-hook
   limitation, and command-replacing-extension limitation. Product release and
   sibling role documentation remain subplan 04.

## Validation Steps

Run from the Xonovex repository root on Linux with Nix available:

```bash
npx moon run moon-nix-extension:moon-build
npx moon run moon-nix-extension:test
npx moon run moon-nix-extension:integration
npx moon run moon-nix-extension:ci-check
npx moon run moon-nix-toolchain:ci-check
nix flake check
```

Run the integration suite once with a cold isolated Moon/proto/plugin cache and
once warm. Review the generated Moon run report/hash evidence rather than only
the runner exit code. macOS/WASI unit coverage remains in subplan 02; the full
Nix black-box gate is Linux because CI publishes the WASI artifact from Linux.

## Success Criteria

- [x] The nested fixture uses the repository-pinned Moon 2.3.5, the local
      extension WASM, native toolchains with null versions, and no `nix`
      toolchain/plugin locator.
- [x] Moon itself reports Node detection without `system`; mapping-gated
      `baseComponents` still resolves `[general, node]`.
- [x] A real polyglot project/task resolves the stable deduplicated
      `[general, go, node]` set without a predeclared combination shell.
- [x] Project/task append, replace, installable, and supersession behavior
      matches parent decision 9 for command and script tasks.
- [x] Every unsafe/malformed path, union, component, flake, and evaluation case
      fails non-zero with a deterministic actionable diagnostic.
- [x] `version: null` produces no Proto-managed Node/Go installation while the
      Nix marker executables are available to the task.
- [x] Central and selected project Nix/config changes invalidate only the
      declared consumers; unrelated edits remain cached; uncached tasks rerun.
- [x] Isolated trace/config evidence proves the toolchain-plugin artifact is
      not a runtime companion.
- [x] Fixture validation rejects any Nix toolchain locator/selection; nested
      extension re-entry remains idempotent, and arbitrary peer command
      replacement remains explicitly unsupported.
- [x] The integration task is part of `moon-nix-extension:ci-check`, and the
      extension/runtime/toolchain gates remain green.

## Files Modified/Created

- Modified: `packages/moon/moon-nix-extension/moon.yml`
- Modified: `packages/moon/moon-nix-extension/README.md`
- Created: `packages/moon/moon-nix-extension/tests/fixtures/consumer-workspace/**`
- Created: `packages/moon/moon-nix-extension/tests/integration/**` (or an
  equivalently focused runner under `tests/`)
- Modified: `packages/moon/moon-nix-extension/src/lib.rs` and focused tests to
  pass validated component values through the task environment instead of
  exposing a generated expression in Nix evaluation diagnostics
- Modified: `packages/moon/moon-nix-runtime/src/installable.rs` and tests for
  the fixed expression/environment serialization contract

## Dependencies

Requires completed subplan `02-global-extension-adapter`. This is execution
group 3; release/pilot work must not begin until the black-box gate passes.

## Validation Results

- The standalone consumer runner passes from a normal shell and from the
  repository's Nix-wrapped Moon task. Repeated runs pass with isolated cold
  Moon/Proto/plugin state, and each run proves warm task cache hits through
  `runReport.json`, `lastRun.json`, and `moon hash` manifests.
- Moon 2.3.5 reports `[javascript, npm, node]` for the direct Node command and
  `[go, javascript, npm, node]` for the opaque polyglot script. Run reports show
  the version-null native toolchain setup actions skipped.
- Central and project flake/config mutation assertions, unrelated-file cache
  stability, uncached reruns, all positive environment combinations, the full
  negative matrix, sentinel re-entry, and standalone plugin/config assertions
  pass.
- Runtime (30), extension (17), and existing toolchain (31) tests pass. The
  extension and toolchain `ci-check` gates, clippy, rustfmt, optimized WASM
  build, shellcheck, shfmt, `nix flake check`, and `git diff --check` pass.

## Estimated Duration

2–3 focused engineering days.
