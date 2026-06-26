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
    - packages/moon/moon-nix-toolchain/tests/wrap_test.rs
    - packages/moon/moon-nix-toolchain/CHANGELOG.md
    - packages/moon/moon-nix-toolchain/README.md
skills_to_consult: [moon-guide, general-fp-guide, debugging-guide, code-review-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Plugin: Fail-Closed for Opted-In Tasks When `nix` Is Absent

## Objective

Today `resolve_wrap_target` silently returns `Ok(None)` when `nix` is not on the host PATH (`src/lib.rs:55-57`), so an opted-in task quietly falls back to host tools. Make that path **fail closed**: when the task's project opts in (a per-tag/per-language allowlist, `failClosedByTag` / `failClosedByLanguage`) and `nix` is absent, return an `Err` with a clear `<project>:<task>` message instead of `None`. The `IN_NIX_SHELL` and `MOON_NIX_WRAPPED` double-entry no-ops (checked above the `nix` probe at `lib.rs:44-53`) stay unconditional, and non-opted tasks keep the silent no-op.

## Tasks

1. **Add the two fail-closed allowlists to the typed `NixToolchainConfig`.**
   File: `packages/moon/moon-nix-toolchain/src/lib.rs` — the `NixToolchainConfig` struct introduced by plan `01-plugin-typed-config` (the one `resolve_shell` reads instead of `serde_json::Value`). Match its derive/attr style; the snippet below assumes the struct already sets `#[serde(default, rename_all = "camelCase")]`, so these fields map to `failClosedByTag` / `failClosedByLanguage` in `toolchains.yml`.

   ```rust
   /// Project tags whose tasks MUST run inside nix. When `nix` is unavailable for a
   /// task in a project carrying one of these tags, the plugin errors instead of
   /// silently falling back to host tools. Empty (the default) = no enforcement.
   pub fail_closed_by_tag: Vec<String>,

   /// Project languages whose tasks MUST run inside nix — same fail-closed contract
   /// as `fail_closed_by_tag`, keyed on the project's moon `language`.
   pub fail_closed_by_language: Vec<String>,
   ```

   Note: additive and default-empty, so every existing consumer (`xonovex`, `drodan`) is unchanged until they opt in. If plan 01 derives `schematic::Config` with `#[setting]` per field, mirror that here instead of bare fields.

2. **Extract the shared project-load + tag/language readers so the fail-closed check and `resolve_shell` share one host-load path.**
   File: `packages/moon/moon-nix-toolchain/src/lib.rs` — `resolve_shell` currently inlines `unsafe { load_project_by_id(...) }` and the tag/language digging at `lib.rs:122-150`. Lift them into pure helpers next to the `load_project_by_id` host import (`lib.rs:9-12`) and have `resolve_shell` call them (reuse whatever plan 01 already factored; do not duplicate).

   ```rust
   /// Load a project's fragment over the host. The plugin sandbox cannot read the
   /// workspace directly, so tags/language come from moon via `load_project_by_id`.
   fn load_project(project_id: &str) -> AnyResult<serde_json::Value> {
       Ok(unsafe { load_project_by_id(project_id.to_owned())? }.0)
   }

   /// The project's moon tags (`config.tags`), empty when unset.
   fn project_tags(project: &serde_json::Value) -> impl Iterator<Item = &str> {
       project
           .get("config")
           .and_then(|config| config.get("tags"))
           .and_then(|tags| tags.as_array())
           .into_iter()
           .flatten()
           .filter_map(|tag| tag.as_str())
   }

   /// The project's moon `language`, when present.
   fn project_language(project: &serde_json::Value) -> Option<&str> {
       project.get("language").and_then(|language| language.as_str())
   }
   ```

   Action: single source of truth for reading a project over the host — `resolve_shell` and the new `fail_closed_opted_in` both consume it.

3. **Add `fail_closed_opted_in`, loading the project over the host only when an allowlist is set.**
   File: `packages/moon/moon-nix-toolchain/src/lib.rs` (place after the helpers from Task 2).

   ```rust
   /// Whether the task's project opted into fail-closed nix: one of its tags is in
   /// `failClosedByTag`, or its language is in `failClosedByLanguage`. Returns
   /// `false` without a host round-trip when both allowlists are empty.
   fn fail_closed_opted_in(project_id: &str, config: &NixToolchainConfig) -> AnyResult<bool> {
       if config.fail_closed_by_tag.is_empty() && config.fail_closed_by_language.is_empty() {
           return Ok(false);
       }

       let project = load_project(project_id)?;

       let tag_opt_in = project_tags(&project)
           .any(|tag| config.fail_closed_by_tag.iter().any(|allow| allow == tag));
       let language_opt_in = project_language(&project)
           .is_some_and(|language| config.fail_closed_by_language.iter().any(|allow| allow == language));

       Ok(tag_opt_in || language_opt_in)
   }
   ```

   Action: the empty-allowlist short-circuit keeps the common case (no `failClosed*` configured) free of an extra `load_project_by_id` call.

4. **Make `resolve_wrap_target` fail closed in the `nix`-absent branch; keep the `IN_NIX_SHELL`/`SENTINEL` no-ops above it untouched.**
   File: `packages/moon/moon-nix-toolchain/src/lib.rs:40-57`. Widen the signature to carry the project fragment, the task target (for the message), and the typed config, then replace the bare `Ok(None)` at `55-57`. Leave the guards at `44-49` (`IN_NIX_SHELL`) and `51-53` (`SENTINEL == "1"`) exactly as-is — they return `Ok(None)` *before* the `nix` probe, so a task already inside a dev shell or already wrapped never fails closed.

   ```rust
   fn resolve_wrap_target(
       context: &MoonContext,
       project: &ProjectFragment,
       target: &Target,
       config: &NixToolchainConfig,
   ) -> AnyResult<Option<WrapTarget>> {
       // IN_NIX_SHELL guard (lib.rs:44-49) and SENTINEL guard (lib.rs:51-53) UNCHANGED.

       if !command_exists(&get_host_environment()?, "nix") {
           if fail_closed_opted_in(project.id.as_str(), config)? {
               return Err(anyhow!(
                   "nix is required for `{target}` but `nix` was not found on PATH; \
                    this project opted into fail-closed nix \
                    (failClosedByTag / failClosedByLanguage)"
               ));
           }
           return Ok(None);
       }

       // project-flake detection (lib.rs:59-76) and workspace-flake fallback
       // (lib.rs:78-81) UNCHANGED — read `project.source.as_str()` where the old
       // `project_source` param was used.
       // ...
   }
   ```

   Action: `anyhow!` resolves through the existing `use moon_pdk::*` glob (`warpgate_api` re-exports `anyhow::anyhow`); `AnyResult<T> = anyhow::Result<T>`, so the `Err` propagates through the `?` in each hook. `Target`'s `Display` writes its `id` (`"<project>:<task>"`), giving the exact required message.

5. **Update both hook call sites to pass the fragment, target, and typed config.**
   File: `packages/moon/moon-nix-toolchain/src/lib.rs:179` (`extend_task_command`) and `:220` (`extend_task_script`). Both already deserialize the toolchain config to `NixToolchainConfig` via plan 01 (`let config = ...`); pass it plus `&input.project` and `&input.task.target` into `resolve_wrap_target` instead of `input.project.source.as_str()`.

   ```rust
   let Some(target) = resolve_wrap_target(
       &input.context,
       &input.project,
       &input.task.target,
       &config,
   )? else {
       return Ok(Json(output));
   };
   ```

   Action: drop the old `input.project.source.as_str()` argument; `resolve_wrap_target` now reads `project.source` / `project.id` from the fragment. The `resolve_shell` calls below (`lib.rs:188-193` / `227-232`) keep taking `&config`, so the config is parsed once per hook and shared.

6. **Add fail-closed tests; keep the existing no-op coverage green.**
   File: `packages/moon/moon-nix-toolchain/tests/wrap_test.rs`. Reuse the `nix`-absent simulation from `passthrough_when_nix_absent` (a temp-dir `which` that exits 1). Assert the error via the public `plugin.plugin.call_func_with` (the wrapper's `extend_task_command` `.unwrap()`s the plugin `Err` — calling the container directly returns a `Result`, so PATH is restored *before* asserting and never leaks across `#[serial]` tests).

   ```rust
   #[tokio::test(flavor = "multi_thread", worker_threads = 1)]
   #[serial]
   async fn fails_closed_when_opted_in_and_nix_absent() {
       reset_wrap_env();
       let restore = stub_missing_nix(); // factor the which-exit-1 PATH swap out of passthrough_when_nix_absent

       let mut sandbox = create_empty_moon_sandbox();
       sandbox
           .host_funcs
           .mock_load_project(|_id| serde_json::json!({ "config": { "tags": ["cmake"] } }));
       let plugin = sandbox.create_toolchain("nix").await;

       let mut input = command_input("cmake", &["--build", "."]);
       input.context = plugin.create_context();
       input.project =
           serde_json::from_value(serde_json::json!({ "id": "game-worldgen", "source": "packages/game/game-worldgen" })).unwrap();
       input.task.target = serde_json::from_value(serde_json::json!("game-worldgen:build")).unwrap();
       input.toolchain_config = serde_json::json!({ "failClosedByTag": ["cmake"] });

       let result = plugin
           .plugin
           .call_func_with::<ExtendTaskCommandOutput, _>("extend_task_command", input)
           .await;

       restore(); // restore PATH before asserting
       let err = result.expect_err("opted-in task must fail when nix is absent");
       assert!(format!("{err:?}").contains("nix is required"), "got: {err:?}");
   }

   #[tokio::test(flavor = "multi_thread", worker_threads = 1)]
   #[serial]
   async fn no_op_when_not_opted_in_and_nix_absent() {
       reset_wrap_env();
       let restore = stub_missing_nix();

       let mut sandbox = create_empty_moon_sandbox();
       // Project is tagged `go`, but only `cmake` is fail-closed -> stay a no-op.
       sandbox
           .host_funcs
           .mock_load_project(|_id| serde_json::json!({ "config": { "tags": ["go"] } }));
       let plugin = sandbox.create_toolchain("nix").await;

       let mut input = command_input("go", &["build"]);
       input.toolchain_config = serde_json::json!({ "failClosedByTag": ["cmake"] });
       let output = plugin.extend_task_command(input).await;

       restore();
       assert_eq!(output.command, None, "non-opted task must keep the host-tool no-op");
       assert!(output.args.is_none());
   }
   ```

   Also add `in_nix_shell_outranks_fail_closed`: set `IN_NIX_SHELL=impure`, `failClosedByTag: ["cmake"]`, a `cmake`-tagged project, and `nix` absent — assert `output.command == None` (the `lib.rs:44-49` guard wins, no error). Keep the existing `passthrough_when_nix_absent` (no `failClosed*` config) as the canonical not-opted case.

7. **Document the new config in CHANGELOG and README.**
   Files: `packages/moon/moon-nix-toolchain/CHANGELOG.md` (new `## 0.6.0` minor entry above `## 0.5.0`) and `packages/moon/moon-nix-toolchain/README.md` (config table).

   ```markdown
   ## 0.6.0

   ### Minor Changes

   - Fail closed for opted-in tasks when `nix` is unavailable. New `failClosedByTag`
     and `failClosedByLanguage` allowlists name the project tags/languages whose
     tasks MUST run inside nix; when `nix` is absent for such a task the plugin
     errors (`nix is required for <project>:<task> …`) instead of silently using
     host tools. Tasks outside the allowlists keep the no-op fallback, and the
     `IN_NIX_SHELL` / `MOON_NIX_WRAPPED` double-entry guards still no-op
     unconditionally. Both allowlists default to empty, so existing consumers are
     unaffected until they opt in (e.g. drodan's game/C via `failClosedByTag: [cmake]`).
   ```

   Action: enabling the allowlists in `xonovex` / `drodan` `.moon/toolchains.yml` and cutting the `v0.6.0` tag belongs to plan `07-rollout-and-release`; this plan ships the capability behind a default-off flag.

## Validation Steps

Run from the workspace root `/home/mvierssen/Projects/xonovex/xonovex-platform`. The `moon-nix-toolchain` tasks carry `toolchain: [system, nix]`, so `moon run` enters the workspace `#rust` dev shell (giving the tests a real `nix`).

- **Type check:** `nix develop --command cargo check --manifest-path packages/moon/moon-nix-toolchain/Cargo.toml --all-targets`
- **Lint:** `npx moon run moon-nix-toolchain:lint` (clippy `-D warnings`; fix at root cause, no `#[allow]`)
- **Build (wasm):** `npx moon run moon-nix-toolchain:build` (produces `target/wasm32-wasip1/release/moon_nix_toolchain.wasm`)
- **Tests:** `npx moon run moon-nix-toolchain:test` — the new `fails_closed_when_opted_in_and_nix_absent`, `no_op_when_not_opted_in_and_nix_absent`, `in_nix_shell_outranks_fail_closed`, and all existing `wrap_test.rs` cases pass.
- **Integration:** after `:build`, manually confirm the error surfaces through a hook against the freshly built wasm: in a scratch workspace (or drodan) pin the local wasm in `.moon/toolchains.yml` with `failClosedByTag: [cmake]`, then `PATH=$(echo "$PATH" | sed 's#[^:]*nix[^:]*:##g') npx moon run <game-cmake-task>` and verify moon reports `nix is required for <project>:<task>` and exits non-zero, while the same task with `nix` on PATH still wraps in `nix develop`.

## Success Criteria

- [ ] `failClosedByTag` / `failClosedByLanguage` parse into `NixToolchainConfig` and default to empty (no behavior change for current consumers).
- [ ] An opted-in task (matching tag or language) with `nix` absent returns an `Err` whose message names `<project>:<task>` and is not a host-tool no-op.
- [ ] A non-opted task with `nix` absent still returns `Ok(None)` (silent host-tool fallback preserved).
- [ ] `IN_NIX_SHELL` set and `MOON_NIX_WRAPPED == "1"` still no-op unconditionally, even when `failClosed*` is configured and `nix` is absent.
- [ ] `fail_closed_opted_in` performs no host `load_project_by_id` call when both allowlists are empty.
- [ ] `resolve_shell` and the fail-closed check share one project-load path (no duplicated host-read logic).
- [ ] type_check, lint, build, and the full `wrap_test.rs` suite pass; CHANGELOG `0.6.0` + README updated.

## Files Modified/Created

- `packages/moon/moon-nix-toolchain/src/lib.rs` — add `fail_closed_by_tag` / `fail_closed_by_language` to `NixToolchainConfig`; add `load_project` / `project_tags` / `project_language` / `fail_closed_opted_in` helpers; widen `resolve_wrap_target` signature and replace the `nix`-absent `Ok(None)` (`55-57`) with a fail-closed `Err`; update both hook call sites (`179`, `220`).
- `packages/moon/moon-nix-toolchain/tests/wrap_test.rs` — new fail-closed / not-opted / IN_NIX_SHELL-wins tests; factor the `nix`-absent PATH stub into a `stub_missing_nix` helper.
- `packages/moon/moon-nix-toolchain/CHANGELOG.md` — `## 0.6.0` minor entry.
- `packages/moon/moon-nix-toolchain/README.md` — document the two new config keys.

## Dependencies

- **`01-plugin-typed-config` (group 1) must land first.** This plan adds two fields to, and reads them from, the typed `NixToolchainConfig` struct that plan 01 introduces (replacing the untyped `serde_json::Value` reads in `resolve_shell`). It also threads `&config` (the parsed `NixToolchainConfig`) through `resolve_wrap_target` and the two hooks — wiring plan 01 establishes.
- **Serializes on `src/lib.rs` with siblings `03-plugin-flake-shell-routing` and `04-plugin-cache-coherence`** (parent plan group 2): logically independent but all edit `src/lib.rs` / `resolve_wrap_target` / the hook fns, so land them one after another (any order), not in parallel worktrees.

## Estimated Duration

~0.5-1 day (config fields + helper extraction + the `resolve_wrap_target` branch are small; most of the time is the `moon_pdk_test_utils` error-path tests and the manual nix-absent integration repro).
