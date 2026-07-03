---
type: plan
has_subplans: false
parent_plan: plans/nix-toolchain-hardening.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/moon/moon-nix-toolchain/Cargo.toml
    - packages/moon/moon-nix-toolchain/Cargo.lock
    - packages/moon/moon-nix-toolchain/src/lib.rs
    - packages/moon/moon-nix-toolchain/tests/wrap_test.rs
    - packages/moon/moon-nix-toolchain/README.md
    - packages/moon/moon-nix-toolchain/CHANGELOG.md
skills_to_consult: [moon-guide, general-fp-guide, code-review-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# 01 — Plugin: Typed `NixToolchainConfig` via `define_toolchain_config`

## Objective

Today `resolve_shell` reads the merged toolchain config as an untyped `serde_json::Value` (`src/lib.rs:100-157`), groping for `config.get("shellByTask")`, `config.get("shellByTag")`, etc. with no schema and no validation. This subplan defines a typed `NixToolchainConfig` struct, registers its JSON schema through the moon_pdk_api 2.0.4 `define_toolchain_config` hook (the plugin implements **none** of the tier1/tier2 hooks today), and refactors `resolve_shell` plus both task hooks to read the typed struct — so moon validates user config at load time and every later subplan consumes a real struct instead of `Value` lookups.

## Tasks

### 1. Add the schematic dependency and enable `moon_pdk`'s schematic feature

**File:** `packages/moon/moon-nix-toolchain/Cargo.toml:12-16` (the `[dependencies]` block)

`define_toolchain_config` returns a `schematic::Schema`, and the idiomatic moon path to both the schema **and** validation is `#[derive(schematic::Config)]` + `moon_pdk::parse_toolchain_config_schema` (the latter is gated behind `moon_pdk`'s `schematic` feature, `moon_pdk/src/toolchain.rs:30-51`). `moon_pdk_api` 2.0.4 already pins `schematic = "0.19.7"`, so match that exact version to keep Cargo's feature unification on one copy.

```toml
[dependencies]
extism-pdk = "1.4.1"
moon_pdk = { version = "2.0.4", features = ["schematic"] }
moon_pdk_api = "2.0.4"
schematic = { version = "0.19.7", default-features = false, features = ["config", "schema"] }
serde_json = "1"
```

Note: bump `version = "0.5.0"` -> `version = "0.6.0"` at `Cargo.toml:3` in the same edit (this is a Minor change; the `github-check` task asserts a matching `## 0.6.0` CHANGELOG section). Run `cargo update -p moon_nix_toolchain` so the new dep resolves into `Cargo.lock`. The `## 0.6.0` version line + CHANGELOG section opened here is what sibling group-2 subplans (`03-plugin-flake-shell-routing`, fail-closed, cache-coherence) append their bullets to.

### 2. Define the typed `NixToolchainConfig` struct

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs` (new, after the `use` block at lines 1-3)

`#[derive(schematic::Config)]` gives a serde-`Deserialize` partial type, the finalized struct, **and** the `Schematic` impl that `define_toolchain_config` renders to a schema — all keyed consistently by `#[config(rename_all = "camelCase")]` so the Rust field `shell_by_task` maps to the YAML/JSON key `shellByTask` (matching today's `config.get("shellByTask")` at `lib.rs:107`). The four selector maps default to empty; `shell` defaults to `None`.

```rust
use schematic::{Config, SchemaBuilder};
use std::collections::HashMap;

/// Typed `nix` toolchain configuration, validated against the schema returned by
/// `define_toolchain_config`. The devShell selectors are resolved most-specific
/// first: `shell_by_task` > `shell_by_toolchain` > `shell_by_tag` >
/// `shell_by_language` > `shell` (a project-wide default).
#[derive(Clone, Config, Debug)]
#[config(rename_all = "camelCase")]
pub struct NixToolchainConfig {
    /// Project-wide default devShell name. Empty or `default` selects the flake's
    /// default devShell.
    pub shell: Option<String>,

    /// devShell name keyed by task id.
    pub shell_by_task: HashMap<String, String>,

    /// devShell name keyed by a task toolchain id.
    pub shell_by_toolchain: HashMap<String, String>,

    /// devShell name keyed by a project tag.
    pub shell_by_tag: HashMap<String, String>,

    /// devShell name keyed by the project language.
    pub shell_by_language: HashMap<String, String>,
}
```

Note: do **not** add a `fail_closed` field here. The fail-closed subplan (group 2) adds it as a `#[setting(default = …)]` on this struct; reserving it now would land dead, defaulted config with no reader. Keep the struct strict (schematic's default `allow_unknown_fields = false`) — moon validates user config against the schema from Task 3 before the plugin runs, so the plugin never sees an unknown key in production.

### 3. Implement the `define_toolchain_config` plugin hook

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs` (new `#[plugin_fn]`, beside `register_toolchain` at lines 14-26)

The hook takes no input and returns the struct's schema via `SchemaBuilder::build_root` (schematic_types 0.11.5 `schema_builder.rs:25`). Registering it is what makes moon surface config errors (unknown key, wrong type) at config-load time, instead of the current silent `Value` lookups.

```rust
#[plugin_fn]
pub fn define_toolchain_config() -> FnResult<Json<DefineToolchainConfigOutput>> {
    Ok(Json(DefineToolchainConfigOutput {
        schema: SchemaBuilder::build_root::<NixToolchainConfig>(),
    }))
}
```

Note: `DefineToolchainConfigOutput` is re-exported through `use moon_pdk_api::*;` already at `lib.rs:3` (defined in `moon_pdk_api/src/toolchain/tier1.rs:79-84`); only `SchemaBuilder` needs the new `use schematic::…` from Task 2.

### 4. Refactor `resolve_shell` to take `&NixToolchainConfig`

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs:100-157`

Swap the `config: &serde_json::Value` parameter for `config: &NixToolchainConfig` and replace every `config.get("…")` probe with a typed field access. Critically, preserve the **lazy project-load** optimization at `lib.rs:121` — only call `load_project_by_id` when a tag- or language-based selector is actually configured — by gating on `!is_empty()` of the typed maps instead of `config.get(...).is_some()`.

```rust
fn resolve_shell(
    task_id: Option<&str>,
    task_toolchains: &[Id],
    project_id: &str,
    config: &NixToolchainConfig,
) -> AnyResult<Option<String>> {
    if let Some(value) = task_id.and_then(|id| config.shell_by_task.get(id)) {
        return Ok(normalize_shell(value));
    }

    for toolchain in task_toolchains {
        if let Some(value) = config.shell_by_toolchain.get(toolchain.as_str()) {
            return Ok(normalize_shell(value));
        }
    }

    if !config.shell_by_tag.is_empty() || !config.shell_by_language.is_empty() {
        let project = unsafe { load_project_by_id(project_id.to_owned())? }.0;

        if !config.shell_by_tag.is_empty() {
            let tags = project
                .get("config")
                .and_then(|config| config.get("tags"))
                .and_then(|tags| tags.as_array());

            if let Some(tags) = tags {
                for tag in tags.iter().filter_map(|tag| tag.as_str()) {
                    if let Some(value) = config.shell_by_tag.get(tag) {
                        return Ok(normalize_shell(value));
                    }
                }
            }
        }

        if let Some(value) = project
            .get("language")
            .and_then(|language| language.as_str())
            .and_then(|language| config.shell_by_language.get(language))
        {
            return Ok(normalize_shell(value));
        }
    }

    Ok(config.shell.as_deref().and_then(normalize_shell))
}
```

Note: `normalize_shell` (`lib.rs:87-90`) keeps its `&str` parameter — the `&String` from a `HashMap::get` deref-coerces, and `shell.as_deref()` yields the `Option<&str>` it expects. Precedence and the `default`/empty short-circuit are byte-for-byte unchanged; only the access mechanism changes.

### 5. Parse `input.toolchain_config` into the struct in both task hooks

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs:185-194` (`extend_task_command`) and `224-233` (`extend_task_script`)

Each hook currently passes `&input.toolchain_config` (a `serde_json::Value`) straight into `resolve_shell`. Parse it once via `parse_toolchain_config_schema`, which runs schematic validation and finalizes the struct (a `null`/non-object config -> defaults; `moon_pdk/src/toolchain.rs:32-51`), then pass `&config`.

```rust
    let shell = if target.project_flake {
        None
    } else {
        let config: NixToolchainConfig =
            parse_toolchain_config_schema(input.toolchain_config.clone())?;
        resolve_shell(
            input.task.target.get_task_id().ok(),
            &input.task.toolchains,
            input.project.id.as_str(),
            &config,
        )?
    };
```

Add `use moon_pdk::parse_toolchain_config_schema;` (or call it fully-qualified) and apply the identical 4-line parse-then-resolve change in `extend_task_script` at `lib.rs:224-233`. The `?` on the returned `AnyResult` flows through `FnResult` exactly like the existing `resolve_wrap_target(...)?` calls (`lib.rs:179`, `220`).

Note: keep the `if target.project_flake { None } else { … }` shape intact — `03-plugin-flake-shell-routing` removes that branch later and depends on this typed `resolve_shell` already being in place. This is the only line where this subplan and group 2 touch the same code; the rename of the `resolve_shell` argument type is the contract group 2 builds on.

### 6. Add `wrap_test.rs` coverage for the schema hook and the typed parse

**File:** `packages/moon/moon-nix-toolchain/tests/wrap_test.rs`

The existing selector tests (`wraps_in_named_dev_shell_from_project_shell`, `selects_dev_shell_*`, `shell_by_*_outranks_*`, `lib.rs` lines ~163-348) already feed `toolchain_config = serde_json::json!({ … })` and now exercise the typed parse end-to-end — they are the precedence regression net and must still pass unchanged. Add two cases: one asserting `define_toolchain_config` emits a schema exposing the camelCase keys, and one guarding the `null`-config -> default-shell path.

```rust
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn define_toolchain_config_exposes_camel_case_schema() {
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output: DefineToolchainConfigOutput =
        plugin.plugin.call_func("define_toolchain_config").await.unwrap();

    let rendered = serde_json::to_value(&output.schema).unwrap().to_string();
    for key in ["shell", "shellByTask", "shellByToolchain", "shellByTag", "shellByLanguage"] {
        assert!(rendered.contains(key), "schema should expose `{key}`: {rendered}");
    }
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn typed_config_defaults_to_flake_default_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    // No toolchain config: parse yields struct defaults, so no `#shell` suffix.
    let output = plugin
        .extend_task_command(command_input("cargo", &["build"]))
        .await;

    assert!(
        !flake_ref(&output).contains('#'),
        "an unset config must select the default devShell, got: {}",
        flake_ref(&output)
    );
}
```

Note: the schema test calls the raw `PluginContainer::call_func` (`warpgate-0.30.5/src/plugin.rs:196`) because `moon_pdk_test_utils` 2.0.4 exposes no `define_toolchain_config` wrapper method; `plugin.plugin` is the `Arc<PluginContainer>` field (`toolchain_wrapper.rs:11`). `DefineToolchainConfigOutput` is in scope via the existing `use moon_pdk_test_utils::*;` (`wrap_test.rs:1`).

### 7. Document the typed config in README and CHANGELOG

**Files:** `packages/moon/moon-nix-toolchain/README.md:31-66`, `packages/moon/moon-nix-toolchain/CHANGELOG.md:1-7`

The README's "Selecting a devShell" selector list (lines 33-41) stays accurate — only note that the keys are now schema-validated. Add a `## 0.6.0` / `### Minor Changes` section at the top of the CHANGELOG (the version `Cargo.toml` bumped to in Task 1):

```markdown
## 0.6.0

### Minor Changes

- Define a typed, schema-validated toolchain config. Implements the moon_pdk_api `define_toolchain_config` hook, which registers a JSON schema for `shell`, `shellByTask`, `shellByToolchain`, `shellByTag`, and `shellByLanguage`, so moon validates these keys (unknown key, wrong type) at config-load time instead of silently ignoring them. The plugin now reads a typed `NixToolchainConfig` struct internally rather than probing an untyped `serde_json::Value`; devShell precedence and the lazy project load are unchanged.
```

Note: append, under one short sentence in the README's selector section, that "every key above is validated against the toolchain's published schema." Keep the README's per-key precedence list and YAML examples (lines 43-66) as-is — they remain correct.

## Validation Steps

Run from the xonovex repo root; every command must pass. The `moon-nix-toolchain` tasks carry `toolchain: [system, nix]`, so they run inside `nix develop` automatically (the tests need `nix` on `PATH`).

```bash
# Format + type-check + lint (cargo fmt --check, then cargo clippy --all-targets -- -D warnings)
npx moon run moon-nix-toolchain:fmt-check
npx moon run moon-nix-toolchain:lint

# Build the wasm32-wasip1 artifact (also the Rust type-check)
npx moon run moon-nix-toolchain:build

# Unit + wrap tests (cargo test; depends on :build) — existing selector tests + the two new cases
npx moon run moon-nix-toolchain:test

# Integration: the aggregate gate (build + test + lint + fmt-check) and the release/changelog check
npx moon run moon-nix-toolchain:ci-check
npx moon run moon-nix-toolchain:github-check
```

Integration note: `github-check` (`.moon/tasks/tag-moon-plugin.yml`) greps `CHANGELOG.md` for `^## 0.6.0` and validates the built `.wasm`, so it confirms the Task 1 version bump and Task 7 changelog section line up with a loadable artifact that now exports `define_toolchain_config`.

## Success Criteria

- [ ] `NixToolchainConfig` is a `#[derive(schematic::Config)]` struct with `#[config(rename_all = "camelCase")]` and the five fields (`shell`, `shell_by_task`, `shell_by_toolchain`, `shell_by_tag`, `shell_by_language`); no `fail_closed` field yet.
- [ ] `define_toolchain_config` is implemented and returns `SchemaBuilder::build_root::<NixToolchainConfig>()`.
- [ ] `resolve_shell` takes `&NixToolchainConfig`; no `serde_json::Value::get("…")` config probe remains in it, and the lazy `load_project_by_id` is gated on the typed maps' `!is_empty()`.
- [ ] Both `extend_task_command` and `extend_task_script` parse `input.toolchain_config` via `parse_toolchain_config_schema` and pass `&config`.
- [ ] `Cargo.toml` declares `schematic` + `moon_pdk` `schematic` feature and is bumped to `0.6.0`; `Cargo.lock` is updated.
- [ ] `tests/wrap_test.rs` adds the schema-shape test and the default-shell typed-parse test; all pre-existing selector tests pass unchanged.
- [ ] README notes the keys are schema-validated; CHANGELOG has the `## 0.6.0` section.
- [ ] All Validation Steps pass.

## Files Modified/Created

- `packages/moon/moon-nix-toolchain/Cargo.toml` — add `schematic` dep + `moon_pdk` `schematic` feature; bump `version` to `0.6.0`.
- `packages/moon/moon-nix-toolchain/Cargo.lock` — regenerated for the new dependency.
- `packages/moon/moon-nix-toolchain/src/lib.rs` — add `NixToolchainConfig` + `define_toolchain_config`; retype `resolve_shell`'s `config` param and its field accesses; parse the typed config in both task hooks.
- `packages/moon/moon-nix-toolchain/tests/wrap_test.rs` — add `define_toolchain_config_exposes_camel_case_schema` and `typed_config_defaults_to_flake_default_shell`.
- `packages/moon/moon-nix-toolchain/README.md` — note the keys are schema-validated.
- `packages/moon/moon-nix-toolchain/CHANGELOG.md` — add the `## 0.6.0` Minor section.

## Dependencies

- **None.** This is the group-1 foundation: it introduces the typed config + `define_toolchain_config` hook with no upstream prerequisites.
- **Downstream (must land after this):** `03-plugin-flake-shell-routing`, the plugin fail-closed subplan, and the cache-coherence subplan all read the typed `NixToolchainConfig` / `resolve_shell(&NixToolchainConfig)` introduced here, and the fail-closed subplan adds its `fail_closed` field to this struct. They also append to the `## 0.6.0` CHANGELOG section and rely on the `Cargo.toml` `0.6.0` bump this subplan makes. Group 2 serializes on `src/lib.rs`, so this subplan keeps `resolve_wrap_target`, the two hook fns' names, and the `if target.project_flake { … }` shape stable to minimize their cross-diff churn.

## Estimated Duration

~0.5 day — a contained `src/lib.rs` change (one struct, one new hook, one parameter retype, two parse sites), one dependency wiring, two tests, and doc bullets. The main care points are matching the existing `schematic` version for feature unification and preserving the lazy project-load gate exactly.
