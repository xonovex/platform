---
type: plan
has_subplans: false
parent_plan: plans/moon-nix-extension.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/moon/moon-nix-runtime/**
    - packages/moon/moon-nix-toolchain/Cargo.toml
    - packages/moon/moon-nix-toolchain/Cargo.lock
    - packages/moon/moon-nix-toolchain/moon.yml
    - packages/moon/moon-nix-toolchain/src/lib.rs
    - packages/moon/moon-nix-toolchain/tests/**
skills_to_consult:
  - moon-guide
  - testing-guide
  - git-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  integration: passed
---

# 01 — Extract the shared Nix runtime without changing the toolchain plugin

## Objective

Create a private, adapter-neutral Rust crate for Nix target resolution, guards,
safe serialization, and command/script wrapping. Move the proven logic out of
`moon_nix_toolchain` while retaining its complete 0.6.1 schema, selectors,
lifecycle hooks, WASM exports, diagnostics, and artifact behavior.

## Context (read this first — no other context is assumed)

All code paths and commands resolve from the Xonovex repository root. Anchors
were verified at `80b3d773dde1e1dee516938096e9022cceccda0a` on
2026-07-18; locate the named construct if they drift. Preserve the unrelated
working-tree changes named in the parent plan.

The existing implementation is one 481-line adapter:

- `packages/moon/moon-nix-toolchain/src/lib.rs:17-42` owns the public typed
  `NixToolchainConfig`. It must stay in the toolchain adapter.
- Lines 44-93 load project tags/language and implement toolchain-specific
  fail-closed selection. These are lifecycle/selector concerns, not generic
  extension behavior.
- Lines 119-203 resolve project/workspace flakes; lines 258-325 build flake
  references, probe named devShells, read locks, and quote scripts.
- Lines 328-415 wrap command argv and opaque scripts; lines 417-480 implement
  the toolchain-only `hash_task_contents` and `setup_environment` hooks.
- `tests/wrap_test.rs` and `tests/hash_test.rs` are the compatibility oracle.
  They cover command/script wrapping, selectors, fail-closed behavior, flake
  hashing, and eager setup. Do not weaken or replace these with only unit tests
  of the extracted crate.
- `Cargo.toml:9-23` builds both `cdylib` and `lib` and pins PDK 2.0.4. The new
  runtime is a normal private Rust library statically linked through a path
  dependency; it is not another WASM plugin or runtime download.
- `.moon/tasks/tag-rust.yml` provides only lint/format aliases. The new runtime
  project therefore needs explicit build/test/`ci-check` tasks in its own
  `moon.yml`; it must not use the `moon-plugin` release tag.

The extraction follows the repository's functional style: host side effects
stay at the adapter boundary, while the shared crate receives explicit facts
and returns immutable decisions. Do not hide PDK calls or mutable caches behind
a global singleton.

## Tasks

1. **Freeze the 0.6.1 compatibility surface before moving code.** Re-read
   `src/lib.rs`, `tests/wrap_test.rs`, `tests/hash_test.rs`, `README.md`, and
   `CHANGELOG.md`. Add focused regression cases only where the current tests do
   not pin an observable behavior needed by the extraction: public schema
   field names, selector precedence, missing-Nix fail-open/fail-closed split,
   project-flake fallback, exact command argv, script quoting, hash contents,
   and setup command. Record the pre-extraction test result.

2. **Scaffold `packages/moon/moon-nix-runtime`.** Create `Cargo.toml`,
   `Cargo.lock`, `moon.yml`, and focused source/test modules. Use
   `publish = false`, a normal `lib` crate, and only dependencies required by
   pure data/serialization logic. Configure the project as `language: rust`,
   `layer: library`, `tags: [rust]`, with explicit `cargo check`, build, and
   test tasks plus a `ci-check` aggregate. Do not give it `moon-plugin`, a
   changelog, GitHub publishing tasks, or a `cdylib` output.

3. **Define the adapter-neutral boundary before extracting implementation.**
   Model host observations and output decisions explicitly. The exact names may
   follow the code, but preserve a shape equivalent to:

   ```rust
   pub struct WrapFacts {
       pub in_nix_shell: bool,
       pub already_wrapped: bool,
       pub nix_available: bool,
   }

   pub enum WrapDecision {
       Unchanged,
       Command { command: String, args: Vec<String> },
       Script(String),
   }
   ```

   Keep PDK types (`MoonContext`, `ProjectFragment`, `Extend*Output`) and host
   functions (`get_host_env_var`, `exec_captured`, project loading) out of the
   shared public API. The adapter gathers facts and filesystem results; pure
   functions validate names/paths, resolve targets, serialize Nix/POSIX
   strings, and create a decision.

4. **Extract the reusable logic in small modules.** Move the sentinel constant,
   Nix availability/re-entry guard decisions, workspace/project flake target
   values, flake/installable formatting, safe Nix-string and POSIX-shell
   serialization, and command/script plan construction into the runtime. Keep
   toolchain-only tag/language/task selectors, `load_project_by_id`, lock
   hashing policy, devShell-existence fallback, and eager setup orchestration in
   `moon-nix-toolchain` unless a lower-level pure helper is genuinely shared.
   Avoid changing diagnostics or fallback behavior while moving code.

5. **Convert `moon_nix_toolchain` into a thin adapter.** Add the path dependency
   in its `Cargo.toml`, regenerate only its `Cargo.lock`, and map its existing
   PDK inputs/config into runtime facts/requests. Map `WrapDecision` back to
   `ExtendTaskCommandOutput`/`ExtendTaskScriptOutput`, preserving
   `MOON_NIX_WRAPPED=1`. Leave `register_toolchain`,
   `define_toolchain_config`, `hash_task_contents`, and `setup_environment`
   exported with their existing signatures and metadata. Do not rename the
   crate, WASM file, tag prefix, config fields, or 0.6.1 version in this
   extraction.

6. **Prove source sharing is build-time only.** Use `cargo tree` to show the
   private runtime is a normal path dependency compiled into the toolchain
   artifact. Inspect the optimized WASM exports and confirm they remain the
   toolchain functions only; there must be no network lookup, second WASM file,
   dynamic loader, or runtime locator for `moon-nix-runtime`. Document this
   invariant in code only where the boundary is otherwise non-obvious; public
   product-role documentation belongs to subplan 04.

## Validation Steps

Run from the Xonovex repository root:

```bash
cargo check --all-targets --manifest-path packages/moon/moon-nix-runtime/Cargo.toml
cargo test --manifest-path packages/moon/moon-nix-runtime/Cargo.toml
npx moon run moon-nix-runtime:ci-check
npx moon run moon-nix-toolchain:ci-check
npx moon run moon-nix-toolchain:github-check
cargo tree --manifest-path packages/moon/moon-nix-toolchain/Cargo.toml
```

Compare the toolchain test output and optimized artifact name with the recorded
pre-extraction baseline. Run `git diff --check` and ensure neither unrelated
dirty Xonovex file was changed.

## Success Criteria

- [x] `moon-nix-runtime` is a private normal Rust library with its own passing
      check/build/test/lint/format aggregate and no plugin/release surface.
- [x] PDK calls and Moon fragments remain in the toolchain adapter; shared
      runtime functions consume explicit data and return deterministic values.
- [x] Every pre-extraction `moon_nix_toolchain` test passes unchanged or with
      strictly additive assertions.
- [x] The toolchain's schema, selector precedence, diagnostics, hook exports,
      artifact name, version, and release tag convention are unchanged.
- [x] `hash_task_contents` still changes only for its documented flake target,
      selector, and lock inputs; `setup_environment` still emits the same eager
      realization command.
- [x] The optimized toolchain WASM is self-contained and has no runtime
      dependency on another plugin artifact or locator.

## Files Modified/Created

- Created: `packages/moon/moon-nix-runtime/{Cargo.toml,Cargo.lock,moon.yml}`
- Created: `packages/moon/moon-nix-runtime/src/**`
- Created: `packages/moon/moon-nix-runtime/tests/**`
- Modified: `packages/moon/moon-nix-toolchain/Cargo.toml`
- Modified: `packages/moon/moon-nix-toolchain/Cargo.lock`
- Modified: `packages/moon/moon-nix-toolchain/moon.yml`
- Modified: `packages/moon/moon-nix-toolchain/src/lib.rs`
- Modified only if required to freeze behavior:
  `packages/moon/moon-nix-toolchain/tests/**`

## Dependencies

None. This is execution group 1 and must merge before subplan 02.

## Validation Results

- Baseline: 30 toolchain tests passed (4 hash, 26 wrap); `ci-check` and
  `github-check` passed; the optimized artifact was
  `moon_nix_toolchain.wasm` with the six toolchain hook exports.
- Final: 14 runtime tests and 31 toolchain tests passed (4 hash, 27 wrap).
  Runtime and toolchain `ci-check` plus toolchain `github-check` passed.
- `cargo tree` reports `moon_nix_runtime v0.1.0` as a direct path dependency
  with no dependencies of its own. The runtime produces no WASM artifact.
- The optimized artifact remains `moon_nix_toolchain.wasm`; its exports remain
  memory plus `define_toolchain_config`, `extend_task_command`,
  `extend_task_script`, `hash_task_contents`, `register_toolchain`, and
  `setup_environment`.

## Estimated Duration

1–2 focused engineering days.
