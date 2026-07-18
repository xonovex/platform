---
type: plan
has_subplans: false
parent_plan: plans/moon-nix-extension.md
parallel_group: 2
status: complete
dependencies:
  plans:
    - 01-shared-nix-runtime
  files:
    - packages/moon/moon-nix-runtime/**
    - packages/moon/moon-nix-extension/**
    - flake.nix
    - nix/flake.nix
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

# 02 — Build the global extension adapter and composable flake contract

## Objective

Add the independently deployable `moon_nix_extension` WASM package. Resolve a
task's native Moon toolchains plus typed project/task overrides into one safe,
deterministic Nix environment, expose the workspace flake composition function,
and wrap command and script tasks without selecting the `nix` toolchain.

## Context (read this first — no other context is assumed)

All code paths and commands resolve from the Xonovex repository root. Anchors
were verified at `80b3d773dde1e1dee516938096e9022cceccda0a` on
2026-07-18; locate named constructs after subplan 01 instead of assuming line
numbers still match. Subplan 01 must be merged and `moon-nix-runtime` green.

Online verification deliberately targets the workspace-pinned stack:

- Current public Moon documentation is on v2.4, but Xonovex pins Moon 2.3.5.
  Do not use v2.4-only task checks or project task defaults in this package.
- The Moon v2 extension API is real and supports pipeline
  `extend_task_command`/`extend_task_script` hooks. The exact pinned PDK 2.0.4
  types are authoritative: `register_extension` uses
  `RegisterExtensionInput`/`RegisterExtensionOutput`, not the older
  `ExtensionMetadata*` names still shown in part of the public guide.
- `ExtendTaskCommandInput` and `ExtendTaskScriptInput` carry
  `extension_config`, `ProjectFragment`, and `TaskFragment`.
  `ProjectFragment` contains aliases, dependency scope, ID, source, and
  toolchains; `TaskFragment` contains only target and toolchains. Arbitrary
  `project.*` custom metadata is not present, so v0.1 keeps overrides in typed
  `.moon/extensions.yml` as required by parent decision 9.
- Moon 2.3.5 detects toolchains from the task command and falls back to project
  toolchains when nothing matches. `system` is a fallback, not an additive
  base: a Node command normally produces `[javascript, npm, node]`. Therefore
  `baseComponents: [general]` is a separate, mapping-gated setting.
- Peer extensions are evaluated from the same input, so multiple
  command-replacing extensions for one mapped task are unsupported. The Nix
  toolchain plugin and this extension are mutually exclusive and are never
  configured together.
- Nix 2.34 documents `--expr` as an expression-relative installable and
  `--impure` as required for mutable paths/`builtins.currentSystem`. The fixed
  expression must return an attribute set containing a `moon` derivation, then
  invoke `nix develop --impure --expr <expr> moon --command ...`.

Relevant current Xonovex code:

- `flake.nix:17-60` composes static `default`, `go`, `shell`, and `rust`
  devShells from `nixShells.devShells.${system}`.
- `nix/flake.nix:21-40` exposes the curated component map (`general`, `node`,
  `go`, `shell`, `k8s`, `rust`, `release`, `ci`, `docker`). Reuse this map;
  do not duplicate package lists in the extension.
- `.moon/tasks/tag-moon-plugin.yml` already builds, optimizes, strips, tests,
  validates, checksums, and publishes every `[rust, moon-plugin]` project.
- The extension must start at version 0.1.0 with crate/artifact name
  `moon_nix_extension`; release qualification and publication remain subplan 04.

## Tasks

1. **Scaffold the extension package using the existing plugin template.** Create
   `packages/moon/moon-nix-extension/{Cargo.toml,Cargo.lock,moon.yml,README.md,CHANGELOG.md}`,
   `src/lib.rs`, and focused test modules. Configure
   `crate-type = ["cdylib", "lib"]`, version `0.1.0`, `publish = false`,
   PDK/API/test-utils 2.0.4,
   Extism 1.4.1, Schematic matching the sibling plugin, and a path dependency on
   `../moon-nix-runtime`. Set `language: rust`, `layer: library`, and tags
   `[rust, moon-plugin]` so the existing release pipeline is inherited without
   copying it.

2. **Implement the exact extension registration and typed schema.** Export
   `register_extension` and `define_extension_config`, returning the pinned PDK
   types. Define camelCase configuration equivalent to:

   ```yaml
   baseComponents: [general]
   environmentByToolchain:
     javascript: node
     node: node
     npm: node
     go: go
     system: general
   environmentByProject: {}
   environmentByTask: {}
   failClosed: true
   ```

   Model an override as a mutually exclusive union of either
   `{components, mode: append|replace}` or `{installable}`. Reject unknown
   fields, empty/malformed component names, mixed union shapes, project keys
   that are not Moon IDs, and task keys that are not full `project:task`
   targets. Use PDK project loading for semantic project/task validation when
   scoped maps are configured; do not read arbitrary `moon.yml` text directly.

3. **Implement deterministic environment resolution in the shared runtime.**
   Keep activation and component selection pure:

   ```text
   matching toolchain mapping OR matching scoped override?
       no  -> unchanged (baseComponents alone never activates)
       yes -> baseComponents
              + all mapped toolchain components
              -> project append/replace/installable
              -> task append/replace/installable
              -> validate, sort, deduplicate
   ```

   A project `replace` replaces both base and inferred components; a task scope
   resolves last. A task replace/installable may supersede a project
   installable, while append-over-installable is an explicit error. Produce a
   resolution value suitable for diagnostics and tests, never a raw Nix
   fragment.

4. **Add safe installable and expression construction.** Canonicalize every
   configured `path:./relative#devShell` against `MoonContext.workspace_root`,
   require the canonical flake directory to remain inside the workspace, and
   require an explicit devShell name after `#`. Reject remote schemes, missing
   attributes, parent escapes (including symlink escapes), and inline
   expressions. For component composition, generate only this fixed structure:

   ```nix
   let flake = builtins.getFlake "path:/canonical/workspace";
   in {
     moon = flake.lib.mkMoonShell builtins.currentSystem [ "general" "node" ];
   }
   ```

   Serialize every path/component as a Nix string; never concatenate user text
   as Nix syntax. Direct project installables use the canonical
   `path:/...#name` flake reference and bypass the expression.

5. **Expose `lib.mkMoonShell` from the root flake.** Refactor `flake.nix`'s
   output `let` bindings just enough to export:

   ```nix
   lib.mkMoonShell = system: names:
     let
       pkgs = nixpkgs.legacyPackages.${system};
       shells = nixShells.devShells.${system};
     in
     pkgs.mkShell {
       inputsFrom = map (name: shells.${name}) names;
     };
   ```

   Make the existing static devShells call the same composition helper where it
   improves consistency without changing their package closures. An unknown
   component must fail Nix evaluation with the missing name; an empty component
   list must remain valid when selected deliberately by a `replace` override.
   Run `nix flake check` and a read-only expression evaluation for at least
   `[general, node]` and `[general, go, node]`.

6. **Implement task command/script hooks and re-entry guards.** Parse
   `input.extension_config`, resolve from `input.task.toolchains`,
   `input.project.id`, and `input.task.target`, then map the runtime decision to
   PDK outputs. Command argv must be replaced as:

   ```text
   nix develop --impure --expr <expression> moon --command <original> <args...>
   ```

   Scripts must remain one opaque string executed through `bash -c` inside the
   shell with tested POSIX quoting. Set `MOON_NIX_WRAPPED=1` on wrapped output;
   no-op for `IN_NIX_SHELL` and an inherited sentinel on a nested invocation.
   Missing Nix is an error only for an active mapping when `failClosed` is true.
   Document that `moon_nix_toolchain` and this extension are mutually exclusive,
   and that another command-replacing extension on a mapped task is unsupported;
   environment-only extensions remain valid.

7. **Build a layered test suite before integration.** Add pure runtime tests for
   activation, base components, mapping deduplication, precedence, union errors,
   installable transitions, path containment, and both serializers. Add PDK
   tests for registration/schema generation, config parsing, command/script
   outputs, exact Project/Task fragment use, missing Nix, and
   outer-shell/sentinel guards. Include hostile strings (`"`,
   `\\`, `${`, quotes, spaces), workspace paths with spaces, malformed targets,
   and deterministic error text. Leave full Moon/Nix/cache behavior for
   subplan 03.

## Validation Steps

Run from the Xonovex repository root:

```bash
cargo check --all-targets --manifest-path packages/moon/moon-nix-extension/Cargo.toml
cargo test --manifest-path packages/moon/moon-nix-extension/Cargo.toml
npx moon run moon-nix-runtime:ci-check
npx moon run moon-nix-toolchain:ci-check
npx moon run moon-nix-extension:ci-check
nix flake check
nix eval --impure --expr 'let f = builtins.getFlake "path:'"$(pwd)"'"; in (f.lib.mkMoonShell builtins.currentSystem [ "general" "node" ]).name'
```

Also inspect `moon extension info nix-environment` against a temporary local
`file://` registration and confirm `register_extension`,
`define_extension_config`, `extend_task_command`, and `extend_task_script` are
reported. The final expression command should be copied from a test fixture,
not hand-edited around escaping failures.

## Success Criteria

- [x] `moon_nix_extension.wasm` builds as an independently deployable artifact
      with only a static Rust dependency on `moon-nix-runtime`.
- [x] Moon validates the documented camelCase config and rejects malformed
      override unions/keys before wrapping an affected task.
- [x] A detected Node task without `system` resolves `[general, node]` through
      mapping-gated `baseComponents`; an unmapped/unoverridden task is unchanged.
- [x] Project then full-target task precedence exactly matches parent decision
      9, including deterministic installable transition errors.
- [x] Central components use only `lib.mkMoonShell`; project/task installables
      accept only canonical workspace-contained locked `path:` flakes.
- [x] Command argv and script text survive quoting/escaping tests with the same
      exit behavior and working directory as the original task.
- [x] Outer Nix/nested Moon re-entry is idempotent without relying on or loading
      the Nix toolchain plugin.
- [x] Arbitrary peer command replacement is documented as unsupported instead
      of being falsely claimed safe through the sentinel.
- [x] Runtime, existing toolchain, extension, flake, clippy, rustfmt, and WASM
      validation commands all pass.

## Files Modified/Created

- Created: `packages/moon/moon-nix-extension/**`
- Modified: `packages/moon/moon-nix-runtime/src/**` and tests as required for
  the adapter-neutral composition contract
- Modified: `flake.nix`
- Read-only contract source: `nix/flake.nix`

## Dependencies

Requires completed subplan `01-shared-nix-runtime`. This is execution group 2;
subplan 03 must wait for it to merge.

## Validation Results

- 30 runtime tests, 31 existing toolchain tests, and 17 extension tests pass.
  Runtime, toolchain, and extension `ci-check` tasks pass; the extension
  `github-check` task also passes, including clippy, rustfmt, optimized WASM
  build, `wasm-validate`, and changelog-version checks.
- `nix flake check` passes. Read-only evaluations accept `[general, node]`,
  `[general, go, node]`, and the deliberate empty component set; an unknown
  component fails with an error that names the missing component.
- Local `moon extension info` reports the expected configuration schema and
  only the registration, configuration, command, and script extension hooks.
  The optimized WASM exports those same four hooks plus memory.
- `cargo tree` reports `moon_nix_runtime` as a direct path dependency compiled
  into the extension, with no runtime plugin locator or dynamic dependency.
- Existing static devShell derivation paths are unchanged, and
  `git diff --check` passes.

## Estimated Duration

3–4 focused engineering days.
