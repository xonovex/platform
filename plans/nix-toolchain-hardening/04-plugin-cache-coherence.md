---
type: plan
has_subplans: false
parent_plan: plans/nix-toolchain-hardening.md
parallel_group: 2
status: pending
dependencies:
  plans: [plugin-typed-config]
  files:
    - packages/moon/moon-nix-toolchain/src/lib.rs
    - packages/moon/moon-nix-toolchain/tests/hash_test.rs
    - packages/moon/moon-nix-toolchain/CHANGELOG.md
    - packages/moon/moon-nix-toolchain/README.md
skills_to_consult: [moon-guide, general-fp-guide, debugging-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Plugin Cache Coherence — `hash_task_contents` (+ optional `setup_environment`)

## Objective

Implement the `moon_pdk_api` 2.0.4 tier2 `hash_task_contents` hook in `moon_nix_toolchain` so that editing the flake the task runs in (`flake.lock`) or switching the resolved devShell **busts the dependent tasks' moon cache**, while an unrelated edit leaves the cache key untouched. Optionally implement `setup_environment` to pre-build (and GC-root) the resolved devShell for cold-start and GC safety. The hook resolves the same project-flake-vs-workspace-flake root and the same shell selector the wrap hooks already use, so a task's cache identity tracks the exact nix environment it executes in.

## Tasks

1. **Extract a guard-free `resolve_flake_target` shared by the wrap hooks and the hash hook.** `resolve_wrap_target` (`packages/moon/moon-nix-toolchain/src/lib.rs:40-82`) currently fuses three runtime no-op guards (`IN_NIX_SHELL` at 44-49, `SENTINEL` at 51-53, `command_exists nix` at 55-57) with the path resolution (project flake at 59-76, workspace fallback at 78-81). The cache key must **not** depend on transient env (`IN_NIX_SHELL`/`SENTINEL`) or on whether `nix` happens to be installed on the hashing host — otherwise the same task hashes differently under CI's outer shell vs a local run. Split the pure path resolution out:

   ```rust
   /// Resolve the flake that wraps a task purely from paths, with no runtime
   /// guards: the project flake when `<project>/flake.nix` exists, else the
   /// workspace flake. Shared by the wrap hooks (after their guards) and
   /// `hash_task_contents`, whose cache key must not depend on transient env
   /// (`IN_NIX_SHELL`/`SENTINEL`) or on `nix` being installed on the hashing host.
   fn resolve_flake_target(
       context: &MoonContext,
       project_source: &str,
   ) -> AnyResult<Option<WrapTarget>> {
       if !project_source.is_empty() {
           if let Some(project_root) = context.workspace_root.join(project_source).real_path() {
               let flake = project_root.join("flake.nix");
               let flake_path = flake.to_string_lossy();

               if exec_captured("test", ["-f", flake_path.as_ref()])
                   .is_ok_and(|result| result.exit_code == 0)
               {
                   return Ok(Some(WrapTarget {
                       root: project_root.to_string_lossy().into_owned(),
                       project_flake: true,
                   }));
               }
           }
       }

       Ok(context.workspace_root.real_path().map(|path| WrapTarget {
           root: path.to_string_lossy().into_owned(),
           project_flake: false,
       }))
   }

   fn resolve_wrap_target(
       context: &MoonContext,
       project_source: &str,
   ) -> AnyResult<Option<WrapTarget>> {
       if !get_host_env_var("IN_NIX_SHELL")?.unwrap_or_default().is_empty() {
           return Ok(None);
       }
       if get_host_env_var(SENTINEL)?.unwrap_or_default() == "1" {
           return Ok(None);
       }
       if !command_exists(&get_host_environment()?, "nix") {
           return Ok(None);
       }
       resolve_flake_target(context, project_source)
   }
   ```

   Note: pure refactor — `resolve_wrap_target`'s external behavior is unchanged, so the existing `tests/wrap_test.rs` cases (`wraps_command_task_in_nix_develop`, `no_op_when_already_in_nix_shell`, `no_op_when_sentinel_set`, `passthrough_when_nix_absent`, `wraps_in_project_flake_when_project_has_one`) keep passing.

2. **Add a host-side `flake_lock_contents` reader.** New helper near `flake_ref` (`src/lib.rs:161-166`). The plugin sandbox has no direct workspace read access (same constraint that forces `test -f` over the host at `lib.rs:63-67`), so read the lock with `exec_captured("cat", …)`:

   ```rust
   /// Read a flake's `flake.lock` over the host so its pinned inputs fold into the
   /// task hash. Returns an empty string when the lock is absent (a flake with no
   /// lock, or a non-flake workspace root) — an absent lock is a stable value, so
   /// it never thrashes the cache.
   fn flake_lock_contents(root: &str) -> String {
       let lock_path = format!("{root}/flake.lock");
       exec_captured("cat", [lock_path.as_str()])
           .ok()
           .filter(|result| result.exit_code == 0)
           .map(|result| result.stdout)
           .unwrap_or_default()
   }
   ```

   Note: pure function of `root` plus the host file — deterministic for a fixed lock, matching the FP guidance for the hash hook.

3. **Implement the `hash_task_contents` plugin_fn.** Add after `extend_task_script` (`src/lib.rs:214-245`). `HashTaskContentsInput` (`moon_pdk_api` `toolchain/tier2.rs:342-358`) carries `context: MoonContext`, `project: ProjectFragment` (`.id: Id`, `.source: String`), `task: TaskFragment` (`.target: Target`, `.toolchains: Vec<Id>`), and `toolchain_config`. `HashTaskContentsOutput` is `{ contents: Vec<serde_json::Value> }` (`tier2.rs:360-364`) which moon folds into the task's cache key. Mirror the wrap hooks' shell resolution exactly so the hash tracks the same `{root}#{shell}` the task runs in:

   ```rust
   #[plugin_fn]
   pub fn hash_task_contents(
       Json(input): Json<HashTaskContentsInput>,
   ) -> FnResult<Json<HashTaskContentsOutput>> {
       let mut contents = Vec::new();

       if let Some(target) = resolve_flake_target(&input.context, input.project.source.as_str())? {
           // A project-local flake uses its own default devShell; the workspace-flake
           // shell selectors do not apply to it (same rule as the wrap hooks).
           let shell = if target.project_flake {
               None
           } else {
               resolve_shell(
                   input.task.target.get_task_id().ok(),
                   &input.task.toolchains,
                   input.project.id.as_str(),
                   &input.toolchain_config,
               )?
           };

           // Fold the resolved flake root, the selected shell, and the lock's pinned
           // inputs into the cache key: editing flake.lock or switching the shell
           // changes `contents`; an unrelated edit leaves it byte-identical.
           contents.push(serde_json::json!({
               "flakeRoot": target.root,
               "shell": shell,
               "flakeLock": flake_lock_contents(&target.root),
           }));
       }

       Ok(Json(HashTaskContentsOutput { contents }))
   }
   ```

   Note: when `resolve_flake_target` returns `None` (no real workspace path), `contents` is empty — the plugin contributes nothing and moon falls back to its own input hashing; never an error. moon discovers the exported `hash_task_contents` symbol automatically; no change to `register_toolchain` is needed.

4. **(Optional) Implement `setup_environment` to pre-build + GC-root the devShell.** `SetupEnvironmentInput` (`tier2.rs:35-57`) carries `project: Option<ProjectFragment>`, `root: VirtualPath`, and `toolchain_config`; `SetupEnvironmentOutput.commands: Vec<ExecCommand>` (`tier2.rs:59-72`) runs during setup. Warm the closure once so the first wrapped task is not a cold `nix develop`, gated on `nix` actually being present and `allow_failure` so it never blocks setup:

   ```rust
   #[plugin_fn]
   pub fn setup_environment(
       Json(input): Json<SetupEnvironmentInput>,
   ) -> FnResult<Json<SetupEnvironmentOutput>> {
       let mut output = SetupEnvironmentOutput::default();

       if !command_exists(&get_host_environment()?, "nix") {
           return Ok(Json(output));
       }

       let project_source = input.project.as_ref().map_or("", |p| p.source.as_str());
       if let Some(target) = resolve_flake_target(&input.context, project_source)? {
           // Realise (and cache) the devShell closure before the first task runs.
           // `--profile` writes a GC root so a `nix store gc` between setup and the
           // task cannot evict it. `allow_failure` keeps setup non-blocking.
           let reference = flake_ref(&target.root, None);
           output.commands.push(
               ExecCommand::new(ExecCommandInput::new(
                   "nix",
                   ["develop", reference.as_str(), "--command", "true"],
               ))
               .allow_failure(true)
               .label(format!("Pre-building nix devShell {reference}")),
           );
       }

       Ok(Json(output))
   }
   ```

   Note: optional and additive — ship task 3 first; only add this once the hash hook is green. `ExecCommand::new` / `ExecCommandInput::new` come from `moon_pdk_api` (`host.rs:50`, `warpgate_api` `host_funcs.rs`), already in scope via `use moon_pdk_api::*`. Scope a `--profile <gcroot>` upgrade to a follow-up if a writable root path is needed.

5. **Add `moon_pdk_test_utils` cases in a new `tests/hash_test.rs`.** Use a separate file (not `wrap_test.rs`) to minimize the serialize-on-one-file conflict with siblings `02-plugin-fail-closed` / `03-plugin-flake-shell-routing` in group 2. The wrapper exposes `plugin.hash_task_contents(HashTaskContentsInput)` (`moon_pdk_test_utils` `toolchain_wrapper.rs:120-138`) and overrides `input.context` with the sandbox root, so `cat <sandbox.root>/flake.lock` reads the real temp file:

   ```rust
   use moon_pdk_test_utils::*;
   use serial_test::serial;

   fn ws_input(source: &str) -> HashTaskContentsInput {
       let mut input = HashTaskContentsInput::default();
       input.project =
           serde_json::from_value(serde_json::json!({ "id": "proj", "source": source })).unwrap();
       input
   }

   #[tokio::test(flavor = "multi_thread", worker_threads = 1)]
   #[serial]
   async fn flake_lock_edit_busts_the_hash() {
       let sandbox = create_empty_moon_sandbox();
       std::fs::write(sandbox.root.join("flake.lock"), r#"{"version":7,"nodes":{}}"#).unwrap();
       let plugin = sandbox.create_toolchain("nix").await;

       let before = plugin.hash_task_contents(ws_input("")).await;

       std::fs::write(sandbox.root.join("flake.lock"), r#"{"version":7,"nodes":{"a":1}}"#).unwrap();
       let after = plugin.hash_task_contents(ws_input("")).await;

       assert_ne!(
           before.contents, after.contents,
           "editing flake.lock must change the hashed contents"
       );
   }

   #[tokio::test(flavor = "multi_thread", worker_threads = 1)]
   #[serial]
   async fn unrelated_edit_keeps_the_hash() {
       let sandbox = create_empty_moon_sandbox();
       std::fs::write(sandbox.root.join("flake.lock"), r#"{"version":7,"nodes":{}}"#).unwrap();
       let plugin = sandbox.create_toolchain("nix").await;

       let before = plugin.hash_task_contents(ws_input("")).await;

       std::fs::create_dir_all(sandbox.root.join("src")).unwrap();
       std::fs::write(sandbox.root.join("src/main.rs"), "fn main() {}").unwrap();
       let after = plugin.hash_task_contents(ws_input("")).await;

       assert_eq!(
           before.contents, after.contents,
           "an edit that does not touch flake.lock/root/shell must not change the hash"
       );
   }

   #[tokio::test(flavor = "multi_thread", worker_threads = 1)]
   #[serial]
   async fn project_flake_lock_chosen_over_workspace_lock() {
       let sandbox = create_empty_moon_sandbox();
       std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
       std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();
       std::fs::write(sandbox.root.join("packages/proj/flake.lock"), r#"{"marker":"project"}"#)
           .unwrap();
       std::fs::write(sandbox.root.join("flake.lock"), r#"{"marker":"workspace"}"#).unwrap();
       let plugin = sandbox.create_toolchain("nix").await;

       let output = plugin.hash_task_contents(ws_input("packages/proj")).await;
       let blob = serde_json::to_string(&output.contents).unwrap();

       assert!(blob.contains("\\\"marker\\\":\\\"project\\\""), "got: {blob}");
       assert!(!blob.contains("workspace"), "must not embed the workspace lock, got: {blob}");
       assert!(blob.contains("/packages/proj"), "flakeRoot should be the project flake, got: {blob}");
   }
   ```

   Note: the `proj` project source matches the existing `wraps_in_project_flake_when_project_has_one` pattern (`wrap_test.rs:358-387`); the harness keeps a non-empty `project.id`, so `ws_input` controls the source. `serde_json::to_string` escapes the embedded lock JSON, hence the `\"marker\"` assertion form.

6. **Add the 0.6.0 CHANGELOG entry.** `packages/moon/moon-nix-toolchain/CHANGELOG.md` currently tops out at `## 0.5.0` (line 3). Insert a `## 0.6.0` section above it (the group-2 siblings append their bullets to the same section; `07-rollout-and-release` bumps `Cargo.toml` to match):

   ```markdown
   ## 0.6.0

   ### Minor Changes

   - Bust the moon task cache when the flake a task runs in changes. Implements the tier2 `hash_task_contents` hook: it resolves the same flake root (project `flake.nix` when present, else the workspace flake) and devShell selector the wrap hooks use, then folds the resolved flake root, the selected shell, and the `flake.lock` contents into the task's cache key. Editing `flake.lock` or switching the selected devShell invalidates the dependent tasks' cache; an unrelated edit does not. The key is independent of `IN_NIX_SHELL`/`MOON_NIX_WRAPPED` and of whether `nix` is installed on the hashing host, so it is stable across CI and local runs. Optionally pre-builds the resolved devShell during `setup_environment` so the first wrapped task is not a cold `nix develop`.
   ```

   Note: keep the existing `## 0.5.0`…`## 0.1.0` history intact below; `github-check` (`tag-moon-plugin.yml`) still finds the `## 0.5.0` section matching the un-bumped `Cargo.toml` version until group 4 bumps it.

7. **Rebase the `resolve_shell` call onto the typed config + document the hooks.** `01-plugin-typed-config` replaces `resolve_shell`'s untyped `config: &serde_json::Value` (`src/lib.rs:100-157`) with the typed `NixToolchainConfig`. The new `hash_task_contents` call site (task 3) passes `&input.toolchain_config` exactly as the wrap hooks do (`lib.rs:188-193`/`227-232`), so it must adopt whatever signature plugin-typed-config settled on — deserialize `input.toolchain_config` into `NixToolchainConfig` once and pass the typed value. Then add a short paragraph to `packages/moon/moon-nix-toolchain/README.md` documenting that flake/lock/devShell edits invalidate the cache via `hash_task_contents` (and the optional `setup_environment` pre-build):

   ```markdown
   ### Cache coherence

   Editing the flake a task runs in — `flake.lock`, or switching its resolved
   devShell — busts that task's moon cache via the `hash_task_contents` hook; an
   unrelated edit does not. The cache key is independent of `IN_NIX_SHELL` and of
   whether `nix` is installed on the hashing host, so it is stable across CI and
   local runs. `setup_environment` pre-builds the resolved devShell so the first
   wrapped task is not a cold `nix develop`.
   ```

   Note: this task is the cross-subplan seam — confirm the `resolve_shell` signature against the merged `01-plugin-typed-config` before opening the PR; the tests in task 5 pass `toolchain_config` as a `serde_json::Value` on the input regardless of the internal config type.

## Validation Steps

All commands run from the repo root (`packages/game-*`-style tasks run under `nix develop`; the moon-plugin tasks set `toolchain: [system, nix]`, so the cargo invocations execute inside the flake automatically):

```bash
# Format + lint (clippy denies warnings)
npx moon run moon-nix-toolchain:fmt-check
npx moon run moon-nix-toolchain:lint

# Build the wasm artifact (wasm32-wasip1 + wasm-opt/strip)
npx moon run moon-nix-toolchain:build

# Run the plugin tests (depends on build; runs cargo test inside nix develop)
npx moon run moon-nix-toolchain:test

# Full aggregate gate (build + test + lint + fmt-check)
npx moon run moon-nix-toolchain:ci-check
```

Targeted test run while iterating (inside `nix develop`, since `moon_pdk_test_utils` loads the built wasm):

```bash
nix develop --command cargo test --manifest-path packages/moon/moon-nix-toolchain/Cargo.toml --test hash_test
```

Integration: `npx moon run moon-nix-toolchain:github-check` confirms a `## <version>` CHANGELOG section and a non-empty valid `.wasm` exist (the release readiness gate that `07-rollout-and-release` consumes).

## Success Criteria

- [ ] `resolve_flake_target` is extracted (guard-free) and used by both `resolve_wrap_target` and `hash_task_contents`; all existing `wrap_test.rs` cases still pass.
- [ ] `hash_task_contents` is exported and folds `{flakeRoot, shell, flakeLock}` into `HashTaskContentsOutput.contents`.
- [ ] `tests/hash_test.rs`: editing `flake.lock` changes the hashed contents; an unrelated edit does not; a project flake's lock is chosen over the workspace lock.
- [ ] The cache key is independent of `IN_NIX_SHELL`/`MOON_NIX_WRAPPED` and of `nix` presence on the hashing host.
- [ ] (Optional) `setup_environment` emits a non-blocking `nix develop … --command true` pre-build when `nix` is present.
- [ ] `resolve_shell` call site compiles against the typed `NixToolchainConfig` from `01-plugin-typed-config`.
- [ ] `CHANGELOG.md` has a `## 0.6.0` entry; `README.md` documents cache coherence.
- [ ] `moon-nix-toolchain:ci-check` is green (build, test, lint, fmt-check).

## Files Modified/Created

- `packages/moon/moon-nix-toolchain/src/lib.rs` — extract `resolve_flake_target`; add `flake_lock_contents`, `hash_task_contents` (and optional `setup_environment`).
- `packages/moon/moon-nix-toolchain/tests/hash_test.rs` — **new**: three cache-coherence tests.
- `packages/moon/moon-nix-toolchain/CHANGELOG.md` — new `## 0.6.0` section.
- `packages/moon/moon-nix-toolchain/README.md` — cache-coherence paragraph.

## Dependencies

- **`01-plugin-typed-config` (group 1) must land first.** It changes `resolve_shell`'s config parameter from `&serde_json::Value` to the typed `NixToolchainConfig`; the `hash_task_contents` call site (task 3/7) must use that final signature, so this subplan rebases onto it.
- **Serializes with the other group-2 siblings on `src/lib.rs`.** `02-plugin-fail-closed` and `03-plugin-flake-shell-routing` also edit `resolve_wrap_target`/the hooks in `src/lib.rs`; land them one after another (any order), not in parallel worktrees. Putting the new tests in `tests/hash_test.rs` (not `wrap_test.rs`) keeps the test-file edits conflict-free.

## Estimated Duration

~0.5-1 day. The `hash_task_contents` hook + the `resolve_flake_target` extraction + three tests are ~1-2 hours of code; the rest is the optional `setup_environment`, the typed-config rebase, and CHANGELOG/README. Add buffer for the serialize-on-`src/lib.rs` coordination with the two group-2 siblings.
