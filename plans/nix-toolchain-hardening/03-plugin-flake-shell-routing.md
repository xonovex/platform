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
    - packages/moon/moon-nix-toolchain/README.md
    - packages/moon/moon-nix-toolchain/CHANGELOG.md
skills_to_consult: [moon-guide, general-fp-guide, code-review-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 03 — Plugin: Route Project-Flake Tasks to Named devShells

## Objective

Today a project that ships its own `flake.nix` always runs in that flake's **bare default** devShell — `resolve_wrap_target` sets `project_flake: true` and the two hook fns hard-code `shell = None` for it (`src/lib.rs:185-194` command, `224-233` script). This subplan makes a project flake **also** run the existing `shellBy*` selectors, so a matched selector wraps the task as `nix develop {projectRoot}#{shell}`. No new config key is added — the named devShell must simply exist in the project's own flake (documented requirement).

## Tasks

### 1. Always resolve the devShell in `extend_task_command`

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs:183-194`

Drop the `if target.project_flake { None } else { … }` branch; call `resolve_shell` unconditionally. `flake_ref` (lib.rs:161-166) already emits `{root}#{shell}` or bare `{root}`, so a project flake (`target.root` = project root) automatically routes to `{projectRoot}#{shell}` when a selector matches.

```rust
    // Resolve the devShell selector for every flake, project-local or workspace.
    // For a project flake the selected name must be a devShell in THAT flake;
    // no match (or `default`) keeps the flake's bare default devShell.
    let shell = resolve_shell(
        input.task.target.get_task_id().ok(),
        &input.task.toolchains,
        input.project.id.as_str(),
        &input.toolchain_config,
    )?;
```

Note: the `resolve_shell` argument list is unchanged from today; only the `&input.toolchain_config` argument's *type* changes once `plugin-typed-config` lands (untyped `serde_json::Value` -> the typed config struct). This task stays correct either way.

### 2. Always resolve the devShell in `extend_task_script`

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs:224-233`

Identical change in the script hook so scripted tasks (`nix develop <ref> --command bash -c …`) route project flakes the same way.

```rust
    let shell = resolve_shell(
        input.task.target.get_task_id().ok(),
        &input.task.toolchains,
        input.project.id.as_str(),
        &input.toolchain_config,
    )?;
```

Note: keep the existing `flake_ref(&target.root, shell.as_deref())` call at line 239 verbatim — it already handles both forms.

### 3. Remove the now-dead `project_flake` field from `WrapTarget`

**File:** `packages/moon/moon-nix-toolchain/src/lib.rs:28-34` (struct), `59-82` (constructors)

After Tasks 1-2 nothing reads `target.project_flake`, so `cargo clippy -- -D warnings` (the `lint` task) will fail on `field is never read`. Reduce `WrapTarget` to its single read field and drop the field from both construction sites.

```rust
/// The nix flake a task is wrapped with: the real path passed to `nix develop`. A
/// project-local `flake.nix` wins over the workspace flake; either way the resolved
/// devShell selector (`resolve_shell`) is applied, so a project flake routes a task to
/// `{root}#<shell>` when a selector matches — that named devShell must exist in the
/// project flake.
struct WrapTarget {
    root: String,
}
```

Then at the project-flake branch (was lib.rs:70-73) drop `project_flake: true`:

```rust
                return Ok(Some(WrapTarget {
                    root: project_root.to_string_lossy().into_owned(),
                }));
```

And the workspace fallback (was lib.rs:78-81):

```rust
    Ok(context.workspace_root.real_path().map(|path| WrapTarget {
        root: path.to_string_lossy().into_owned(),
    }))
```

Note: keep `resolve_wrap_target`'s name and `Option<WrapTarget>` return type stable — siblings `plugin-fail-closed` and `plugin-cache-coherence` also edit this fn (group 2 serializes on `src/lib.rs`), so avoid a rename that would churn their diffs.

### 4. Update the doc comments + README to document the project-flake requirement

**Files:** `packages/moon/moon-nix-toolchain/src/lib.rs:183-184` (inline comment, superseded by Task 1's new comment) and `packages/moon/moon-nix-toolchain/README.md:70-74` (the "Per-project flakes" section)

The README still claims selectors "do not apply to a project flake" — now they do. Rewrite section "Per-project flakes":

```markdown
## Per-project flakes

When a project ships its own `flake.nix` (i.e. `<projectRoot>/flake.nix` exists), the plugin wraps that project's tasks with the project flake — `nix develop <projectRoot> --command …` — taking precedence over the workspace flake. The shell selectors above still apply: a matching selector routes the task to `nix develop <projectRoot>#<shell>`, so the **named devShell must be exposed by the project flake**. With no match (or a `default` value) the task uses the project flake's default devShell.

The project flake is detected from the project source over the host, so it auto-applies to every project that ships one — no per-project config. Projects without their own `flake.nix` are unchanged: the workspace flake plus the resolved devShell. This lets a package pin its own toolchain independently of the workspace flake.
```

Note: `resolve_wrap_target`'s doc (lib.rs:36-39) — "that project flake wins over the workspace flake" — stays accurate; no edit needed there.

### 5. Rewrite the project-flake test to assert named-shell routing

**File:** `packages/moon/moon-nix-toolchain/tests/wrap_test.rs:356-387` (`wraps_in_project_flake_when_project_has_one`)

That test currently asserts `!flake_ref.contains('#')` with `toolchain_config = { "shell": "go" }` — the exact behavior this subplan reverses. Replace it with a routing assertion, and add a companion no-selector test that still expects the bare default.

```rust
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn routes_project_flake_to_named_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    // A project that ships its own flake.nix.
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();
    // A selector now routes a project flake to one of ITS OWN named devShells.
    input.toolchain_config = serde_json::json!({ "shell": "go" });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("/packages/proj#go"),
        "a project flake with a selector should route to its named devShell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn project_flake_without_selector_uses_default_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();

    let output = plugin.extend_task_command(input).await;

    let flake_ref = flake_ref(&output);
    assert!(
        flake_ref.ends_with("/packages/proj"),
        "project flake root should be used, got: {flake_ref}"
    );
    assert!(
        !flake_ref.contains('#'),
        "no selector should keep the project flake's default devShell, got: {flake_ref}"
    );
}
```

Note: if `plugin-typed-config` changed `toolchain_config` from a `serde_json::Value` to a typed deserialize, keep these `serde_json::json!({...})` literals — `moon_pdk_test_utils` deserializes the input JSON into the typed config, so the test inputs stay JSON.

### 6. Add a per-task routing test for a project flake

**File:** `packages/moon/moon-nix-toolchain/tests/wrap_test.rs` (new test, alongside Task 5)

Lock in that the most-specific selector (`shellByTask`) routes a flake-owning project per task — the headline use case (one project flake, different shells per task).

```rust
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn routes_project_flake_per_task_from_shell_by_task() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();
    input.task.target = serde_json::from_value(serde_json::json!("proj:go-lint")).unwrap();
    // shellByTask outranks the project-wide `shell`, even for a project flake.
    input.toolchain_config = serde_json::json!({
        "shellByTask": { "go-lint": "go" },
        "shell": "default"
    });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("/packages/proj#go"),
        "shellByTask should route a project flake to its named devShell, got: {}",
        flake_ref(&output)
    );
}
```

### 7. Add the CHANGELOG bullet

**File:** `packages/moon/moon-nix-toolchain/CHANGELOG.md`

Append a bullet under the `## 0.6.0` / `### Minor Changes` section that `plugin-typed-config` (group 1) creates when it bumps `Cargo.toml` `version = "0.6.0"`. This subplan does **not** edit `Cargo.toml` — the version bump is owned by `plugin-typed-config` / `rollout-and-release` to keep three group-2 subplans off the same line.

```markdown
- Route a project-flake task to a named devShell. A project that ships its own `flake.nix` now also runs the shell selectors (`shellByTask` > `shellByToolchain` > `shellByTag` > `shellByLanguage` > `shell`); a match wraps the task in `nix develop <projectRoot>#<shell>` instead of always the project flake's bare default devShell. The selected name must be a devShell exposed by the project flake. Projects whose flake exposes only `default` are unchanged: no selector match (or a `default` value) keeps the bare project flake.
```

Note (drodan): drodan's `game-*` project flakes expose only `default` (a `cc` re-export) today, so per-task routing for them needs added named shells in those flakes — **or** they keep their current path of routing to the workspace `#cc` via `shellByTag` (already working). That makes this the lowest-urgency group-2 plugin item; sequence it last in group 2. No drodan flake edits are part of this subplan.

## Validation Steps

Run from the xonovex repo root; every command must pass. The `moon-nix-toolchain` tasks carry `toolchain: [system, nix]`, so they run inside `nix develop` automatically (the tests need `nix` on `PATH`).

```bash
# Format + type-check + lint (cargo fmt --check, then cargo clippy --all-targets -- -D warnings)
npx moon run moon-nix-toolchain:fmt-check
npx moon run moon-nix-toolchain:lint

# Build the wasm32-wasip1 artifact (also the Rust type-check)
npx moon run moon-nix-toolchain:build

# Unit + wrap tests (cargo test; depends on :build)
npx moon run moon-nix-toolchain:test

# Integration: the aggregate gate (build + test + lint + fmt-check)
npx moon run moon-nix-toolchain:ci-check
```

Integration smoke check (manual, optional): in a throwaway dir with a `flake.nix` exposing a named `devShells.go`, opt a task into `[system, nix]` and a `shell: go`, then confirm `moon run <proj>:<task>` invokes `nix develop <projectRoot>#go --command …` (e.g. via `moon run --log trace` or a tool that only the `go` shell provides).

## Success Criteria

- [ ] `extend_task_command` and `extend_task_script` call `resolve_shell` unconditionally — no `target.project_flake` branch remains.
- [ ] `WrapTarget` no longer carries `project_flake`; `cargo clippy -- -D warnings` is clean (no dead-field warning).
- [ ] A project-flake task with a matching `shellBy*`/`shell` selector wraps as `nix develop <projectRoot>#<shell> …`.
- [ ] A project-flake task with no selector (or `default`) still wraps as bare `nix develop <projectRoot> …` (no `#`).
- [ ] `tests/wrap_test.rs` covers all three: named-shell routing, `shellByTask` per-task routing, and the default-shell fallback for a project flake.
- [ ] README "Per-project flakes" and the `WrapTarget` doc comment state that the named devShell must exist in the project flake.
- [ ] CHANGELOG has the `## 0.6.0` bullet; `Cargo.toml` is left to `plugin-typed-config`.
- [ ] All Validation Steps pass.

## Files Modified/Created

- `packages/moon/moon-nix-toolchain/src/lib.rs` — drop the `project_flake` shell short-circuit in both hooks; remove the `project_flake` field from `WrapTarget` and its two constructors; refresh the inline doc comment.
- `packages/moon/moon-nix-toolchain/tests/wrap_test.rs` — rewrite `wraps_in_project_flake_when_project_has_one` into `routes_project_flake_to_named_shell`; add `project_flake_without_selector_uses_default_shell` and `routes_project_flake_per_task_from_shell_by_task`.
- `packages/moon/moon-nix-toolchain/README.md` — rewrite the "Per-project flakes" section.
- `packages/moon/moon-nix-toolchain/CHANGELOG.md` — add the `## 0.6.0` bullet.

## Dependencies

- **`plugin-typed-config` (group 1)** must land first. It (a) replaces the untyped `serde_json::Value` config read with the typed `define_toolchain_config` struct that `resolve_shell` consumes — Tasks 1-2 call `resolve_shell(&input.toolchain_config)` against that type — and (b) bumps `Cargo.toml` to `0.6.0` and opens the `## 0.6.0` CHANGELOG section this subplan appends to.
- **Sibling group-2 subplans** (`plugin-fail-closed`, `plugin-cache-coherence`) also edit `src/lib.rs` (`resolve_wrap_target` / the hook fns). Group 2 serializes on that one file — land these one after another in any order, not in parallel worktrees. This subplan keeps `resolve_wrap_target`'s name and return type stable to minimize cross-diff churn.

## Estimated Duration

~0.5 day (small, well-scoped `src/lib.rs` edit plus three focused tests). The lowest-urgency group-2 item — sequence it last, since drodan's game flakes do not yet expose the named shells that would exercise it in production.
