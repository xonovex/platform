---
type: plan
has_subplans: true
status: approved
updated: 2026-07-18
feature: moon-nix-extension
dependencies:
  plans: []
  subplans:
    01-shared-nix-runtime: []
    02-global-extension-adapter: [01-shared-nix-runtime]
    03-cache-contract-and-consumer-fixture: [02-global-extension-adapter]
    04-release-and-pilot-rollout: [03-cache-contract-and-consumer-fixture]
proposed_subplans:
  - 01-shared-nix-runtime
  - 02-global-extension-adapter
  - 03-cache-contract-and-consumer-fixture
  - 04-release-and-pilot-rollout
parallel_groups:
  - group: 1
    plans: [01-shared-nix-runtime]
    note: "Establish the compatibility-preserving shared runtime before either adapter depends on it."
  - group: 2
    plans: [02-global-extension-adapter]
    depends_on: [1]
    note: "Add the extension adapter and central flake contract on top of the extracted runtime."
  - group: 3
    plans: [03-cache-contract-and-consumer-fixture]
    depends_on: [2]
    note: "Exercise the completed adapter contract through black-box fixture and cache assertions."
  - group: 4
    plans: [04-release-and-pilot-rollout]
    depends_on: [3]
    note: "Publish and pin a consumer only after local and integration validation pass."
skills_to_consult:
  - moon-guide
  - testing-guide
  - shell-scripting-guide
  - versioning-guide
  - git-guide
research_sources:
  documentation:
    - https://moonrepo.dev/docs/config/extensions
    - https://moonrepo.dev/docs/config/project
    - https://moonrepo.dev/docs/config/tasks
    - https://moonrepo.dev/docs/concepts/cache
    - https://moonrepo.dev/docs/setup-toolchain
    - https://moonrepo.dev/docs/guides/extensions
    - https://moonrepo.dev/docs/guides/wasm-plugins
    - https://moonrepo.dev/blog/moon-v2.0
    - https://github.com/moonrepo/moon/blob/v2.3.5/crates/pdk-api/src/extension.rs
    - https://github.com/moonrepo/moon/blob/v2.3.5/crates/process-augment/src/augmented_command.rs
    - https://github.com/moonrepo/moon/blob/v2.3.5/crates/task-builder/src/tasks_builder.rs
    - https://moonrepo.dev/docs/commands/extension/info
    - https://docs.rs/moon_pdk_api/2.0.4/moon_pdk_api/struct.ExtendTaskCommandInput.html
    - https://docs.rs/moon_pdk_api/2.0.4/moon_pdk_api/struct.ExtendTaskScriptInput.html
    - https://nix.dev/manual/nix/latest/command-ref/new-cli/nix3-develop
    - https://nix.dev/concepts/flakes
    - https://nix.dev/guides/recipes/sharing-dependencies
  versions:
    moon: "2.3.5 (xonovex-platform workspace pin)"
    moon-pdk: "2.0.4"
    moon-pdk-api: "2.0.4"
    extism-pdk: "1.4.1"
    moon-nix-toolchain: "0.6.1"
    nix: "2.34.7 (locally inspected)"
    online-verification: "2026-07-18; public docs are on Moon 2.4, while implementation remains source-checked against workspace-pinned Moon 2.3.5 and PDK 2.0.4"
---

# Moon Nix extension

## Overview

Add a globally configured Moon extension in xonovex-platform that maps a task's
already detected native toolchains to Nix environment components, deduplicates
the complete component set, and asks one central workspace flake to compose the
task-specific devShell. Projects retain semantic toolchains such as
`javascript`, `node`, `npm`, and `go` without also selecting `nix`. Keep
`moon_nix_toolchain` supported for consumers that need its toolchain-only setup
and hashing hooks, but make the extension independently deployable: the two
WASM plugins may share Rust source code and never require one another at
runtime. Release the extension through the same reviewed, versioned
GitHub-release mechanism.

## Execution context

All code paths and validation commands below resolve from the Xonovex repository
root. The inspected baseline is
`80b3d773dde1e1dee516938096e9022cceccda0a`; re-locate named constructs if that
checkout moves before execution. Existing unrelated local changes in
`packages/command/command-utility/commands/skill-optimize.md` and
`packages/skill/skill-skill/skill-guide/references/optimize.md` are outside this
plan and must be preserved.

## Goals

- A configured extension sees every task, maps all of its native Moon toolchain
  IDs to Nix environment components, sorts and deduplicates those components,
  and composes the task's devShell without `[system, nix]` in each project or
  task.
- The existing toolchain plugin and the new extension reuse one pure,
  independently tested core for Nix discovery, workspace-flake resolution,
  guards, safe expression/argument serialization, and command/script wrapping.
- A consumer can configure only `moon_nix_extension`; it does not need the
  `moon_nix_toolchain` locator, WASM artifact, registered `nix` toolchain, or
  plugin setup at runtime.
- New Xonovex and Drodan consumers adopt the extension as the default Nix/Moon
  integration. Existing toolchain-plugin consumers remain supported until they
  deliberately migrate after the pilot proves the extension contract.
- Native Moon ecosystem behavior remains enabled; consumers that want Nix to
  replace Proto installation set managed toolchain versions to `null` and make
  the selected devShell provide the binaries.
- Projects and individual targets can append to or replace inferred components
  with schema-validated overrides. A project or target that genuinely needs an
  independent environment can select a named, workspace-local flake devShell
  without injecting raw Nix into Moon configuration.
- Global activation is safe: an unmapped task is unchanged, invalid or missing
  environment components fail with an actionable error, and the extension's
  own nested invocations remain idempotent.
- Flake and extension configuration changes invalidate cached consumer tasks
  through a documented inherited-input contract, compensating for the absence
  of extension-side `hash_task_contents`.
- The extension ships as an optimized, validated WASI artifact with changelog
  notes, a checksum, an independently versioned tag, and a pinned
  `github://` locator, using the existing version-packages release workflow.

## Non-goals

- Replacing Proto inside Moon's `SetupToolchain` action or adding a generic
  toolchain-provider API to Moon core.
- Retiring `moon_nix_toolchain`; it remains the stronger option when a consumer
  values `setup_environment` and `hash_task_contents` over transparent global
  activation.
- Requiring, automatically configuring, or supporting `moon_nix_toolchain` as a
  companion plugin. A workspace uses either the toolchain plugin or the
  extension plugin, never both.
- Accepting arbitrary consumer-provided Nix snippets. The extension emits one
  fixed expression template and serializes only validated flake paths,
  installable attributes, and configured component names into it.
- Reading project overrides from custom `moon.yml` metadata in version 0.1.
  Moon permits such metadata, but extension config schemas cannot validate it
  at workspace-load time; the initial contract keeps all override declarations
  typed in `.moon/extensions.yml` while allowing the referenced flake to remain
  project-owned.
- Unioning toolchains across Moon's complete multi-task action graph. Extension
  hooks are task-scoped, so each task receives the composition derived from its
  own `input.task.toolchains`; Nix store reuse amortizes repeated components.
- Automatically rewriting consumer toolchain versions to `null` or migrating
  consumer repositories in the extension's release PR.

## Current state

- `packages/moon/moon-nix-toolchain` is a standalone Rust `cdylib`/library
  crate at 0.6.1. It implements typed config, project/workspace flake
  selection, task/toolchain/tag/language shell selectors, fail-closed policy,
  command and script wrapping, cache hashing, and devShell pre-building.
- The toolchain's `extend_task_command` and `extend_task_script` already receive
  `input.task.toolchains`, but Moon invokes them only after the task selects the
  `nix` toolchain. This makes native toolchains usable as selectors but leaves
  a per-project `[system, nix]` opt-in.
- Moon 2.3.5 loads configured extensions into a registry and calls supported
  `extend_task_command` / `extend_task_script` hooks across that registry. The
  extension input contains the owning project ID/source and the task target plus
  complete toolchain list. These stable identifiers support typed project and
  target maps in `.moon/extensions.yml`; the task fragment does not carry an
  arbitrary extension-specific settings object.
- Moon 2.3.5 only adds `system` when no enabled toolchain matches a task; it is
  a fallback, not an additive base toolchain. A task invoking Node normally
  resolves `javascript`, its package manager, and `node` without `system`.
  The extension therefore needs a separate `baseComponents` setting for common
  Nix inputs such as `general`.
- The extension API does not expose the toolchain-only
  `hash_task_contents` or `setup_environment` hooks. A global wrapper can be
  transparent at execution time, but cache invalidation and eager shell
  realization need different contracts.
- The Xonovex workspace flake already imports reusable per-tool devShells from
  `nix/flake.nix` and composes its static `default`, `go`, `shell`, and `rust`
  shells with `pkgs.mkShell { inputsFrom = ...; }`. It does not yet expose a
  function that composes an arbitrary validated component list.
- Nix 2.34 supports expression installables for `nix develop --expr`. A local
  workspace flake function can therefore return the derivation used by
  `nix develop`, but resolving a mutable local flake and
  `builtins.currentSystem` requires a deliberately fixed `--impure` wrapper.
  A read-only evaluation against the inspected Xonovex flake successfully
  composed its existing `node` and `general` inputs into a shell derivation.
- `.moon/tasks/tag-moon-plugin.yml` already provides the complete plugin
  pipeline to every `moon-plugin` project: release WASI build, `wasm-opt`,
  `wasm-strip`, Rust tests, clippy, rustfmt, changelog/version validation,
  `wasm-validate`, SHA-256 generation, GitHub release creation, and a dry run.
- `.github/workflows/release.yml` runs affected `:ci-publish` tasks after a
  merged PR whose title contains `version packages`. For a crate named
  `moon_nix_extension` at 0.1.0, the inherited publisher naturally creates
  `moon_nix_extension-v0.1.0` with `moon_nix_extension.wasm` and
  `moon_nix_extension.wasm.sha256` assets.

## Research findings and decisions

1. **Use a standalone extension, not a built-in-toolchain override or runtime
   companion.** Moon permits third-party extensions to hook task commands
   globally. Overriding `node`, `go`, or package-manager plugins would require
   reimplementing their ecosystem behavior and would couple this project to
   Moon internals. Sharing an internal Rust crate with `moon_nix_toolchain` is a
   build-time source relationship only; each release is a self-contained WASM
   artifact.
2. **Keep two thin public adapters over a shared runtime.** Copying the current
   flake and wrapping code would let behavior drift. Extracting a private Rust
   library keeps `moon_nix_toolchain` backward compatible while the extension
   switches from `register_toolchain` / `toolchain_config` to
   `register_extension` / `extension_config`.
3. **Make component mapping the activation and composition boundary.** The
   extension reads every ID in `input.task.toolchains`, collects every matching
   `environmentByToolchain` value, and activates when a toolchain mapping or a
   scoped override matches. Once active, it prepends `baseComponents`, applies
   project/task overrides, then sorts and deduplicates the component names.
   Multiple different mapped values are combined, not treated as a conflict.
   `baseComponents: [general]` supplies the common environment because Moon's
   `system` toolchain is a fallback and is not present alongside detected Node,
   Go, or Rust toolchains. An entirely unmapped and unoverridden task remains
   unchanged even when `baseComponents` is non-empty.
4. **Preserve Moon semantics and make Proto replacement explicit.** Detection,
   project syncing, dependency hashing, and dependency installation remain
   owned by native Moon toolchains. A consumer opts out of Proto installation
   with `version: null`; the extension validates execution, but does not mutate
   `.moon/toolchains.yml`.
5. **Compose through a narrow central-flake function.** The workspace flake
   exposes `lib.mkMoonShell = system: names: ...` and resolves names only from
   its curated per-tool `devShells` set. The extension generates a fixed Nix
   expression that loads the workspace flake and calls this function with
   `builtins.currentSystem` plus the sorted component list. The expression
   returns `{ moon = <derivation>; }`, and the task is wrapped with
   `nix develop --impure --expr <expression> moon --command ...`. Unknown
   components fail during evaluation; user-supplied expression text is never
   accepted.
6. **Treat cache coherence as a consumer contract.** Until Moon exposes an
   extension hashing hook, adopters inherit workspace `flake.nix`, `flake.lock`,
   the central Nix shell definitions, `.moon/extensions.*`, and every selected
   project installable's flake, lock, and Nix sources as task inputs. The package
   documents this limitation prominently and tests the recommended fixture.
   Eager devShell realization is deferred to normal `nix develop` execution.
7. **Reuse the generic release path unchanged.** The new project carries
   `[rust, moon-plugin]`, `Cargo.toml`, `Cargo.lock`, `CHANGELOG.md`, and README.
   Its initial version-packages PR proves `github-check` and
   `github-publish-dry-run`; merging that PR creates the release. Consumer pins
   move only in a later PR after the tag and assets resolve, matching the
   existing toolchain-plugin precedent.
8. **Give the two plugins distinct product roles.** The extension is the
   recommended default for new Xonovex and Drodan configuration because it
   preserves native toolchains and needs no per-task `nix` opt-in. The
   toolchain plugin remains an optional compatibility and special-purpose
   integration for existing consumers, automatic flake hashing through
   `hash_task_contents`, eager devShell realization through
   `setup_environment`, project-local flake discovery, and explicit
   task/tag/language selection. The two plugins are mutually exclusive in a
   workspace. If Moon later exposes equivalent hashing and setup hooks to
   extensions—and the remaining consumers have migrated—the toolchain plugin
   can be considered for retirement in a separate reviewed plan; this plan does
   not deprecate or remove it.
9. **Support scoped overrides without raw expressions.** Resolution starts with
   `environmentByToolchain`, applies an `environmentByProject` entry keyed by
   project ID, then applies an `environmentByTask` entry keyed by the full
   `project:task` target. Component overrides use `append` or `replace`; task
   scope always outranks project scope. Either scope may instead select a
   workspace-local `path:` flake installable with a named devShell attribute.
   `installable` is mutually exclusive with `components`/`mode`, an `append`
   cannot layer onto a less-specific installable, and a more-specific `replace`
   or installable can supersede it. The extension accepts no inline Nix source.
10. **Keep one version owner per executable.** The extension artifact is pinned
    by its immutable release tag. Nix components select package families and
    their `flake.lock` pins exact nixpkgs/overlay revisions; a project-local
    installable uses that project's own lock. Native Moon integrations stay
    enabled with `version: null`. Setting a non-null Moon version for a binary
    also supplied by a mapped Nix component is an unsupported dual-owner
    configuration, not a task/project version override.

### Plugin roles and lifecycle

| Concern             | `moon_nix_extension`                        | `moon_nix_toolchain`                             |
| ------------------- | ------------------------------------------- | ------------------------------------------------ |
| Recommended use     | Default for new Xonovex/Drodan consumers    | Existing compatibility and specialized consumers |
| Activation          | Global, derived from native task toolchains | Explicit `nix` toolchain selection               |
| Environment         | Dynamic central-flake component composition | Named project/workspace devShell selection       |
| Cache integration   | Consumer-declared inherited inputs          | Automatic `hash_task_contents` contribution      |
| Setup               | Lazy at task execution                      | Eager `setup_environment` realization            |
| Deployment relation | Standalone WASM                             | Standalone WASM; not an extension dependency     |

Migration is an atomic configuration replacement: first prove the extension in
an isolated extension-only fixture, then use one consumer PR to remove the
`nix` toolchain locator/selection and add the extension locator, native
component mappings, and cache inputs. No checked-in or test configuration
registers both plugins.

### Configuration contract

The initial public shape maps exact Moon toolchain IDs to composable environment
names while keeping global activation explicit:

```yaml
# .moon/extensions.yml
nix-environment:
  plugin: "github://xonovex/platform/moon_nix_extension@moon_nix_extension-v0.1.0"
  baseComponents: [general]
  environmentByToolchain:
    javascript: node
    node: node
    npm: node
    go: go
    unstable_python: python
    system: general
  environmentByProject:
    legacy-app:
      mode: replace
      components: [general, node20]
    api:
      mode: append
      components: [postgresql]
    special-compiler:
      installable: "path:./packages/compiler/special-compiler#moon"
  environmentByTask:
    legacy-app:test-modern:
      mode: replace
      components: [general, node24]
    api:generate:
      mode: append
      components: [protobuf]
    special-compiler:bootstrap:
      installable: "path:./packages/compiler/special-compiler#bootstrap"
  failClosed: true
```

Moon's native integrations stay configured, while consumers that want Nix to
provide their executables disable Proto-managed installation explicitly:

```yaml
# .moon/toolchains.yml
node:
  version: null
npm:
  version: null
go:
  version: null
```

The same rule applies to any other versioned native toolchain. Mapping keys are
the exact IDs present in `input.task.toolchains`; a custom ID such as
`unstable_python` must therefore match the consumer's registered toolchain ID.
An override changes the Nix component/installable, not the Moon version field.

For the normally detected `input.task.toolchains = [javascript, node, npm]`, a
mapped match activates the extension and base resolution yields the sorted set
`[general, node]`; a polyglot task with `[go, javascript, node, npm]` yields
`[general, go, node]`. Duplicate values are collapsed, while unknown toolchain
IDs are ignored unless explicitly mapped. Project `append`/`replace` applies
next and full-target task `append`/`replace` applies last. A scoped override may
activate an otherwise unmapped task; an empty final result leaves it unchanged.
`IN_NIX_SHELL` and the shared wrapping sentinel prevent re-entry during nested
Moon invocations and execution inside CI or an outer developer shell.

Peer extensions are evaluated from the same pre-extension input and their
replacement outputs are applied afterward; the extension's sentinel cannot
make multiple command-replacing extensions order-independent. A mapped task
must therefore have only one command-replacing extension. The sentinel covers
only the extension's own nested invocation/re-entry behavior.

The override value is a validated union:

- `components` plus `mode: append | replace` (default `append`) continues
  through central composition.
- `installable` selects one named devShell from a workspace-local `path:` flake
  and bypasses `mkMoonShell` for that scope.
- Defining `installable` together with `components` or `mode` is invalid.
- A task `replace` or task `installable` supersedes a project installable; a
  task `append` over a project installable is invalid because arbitrary
  installables are not dynamically recomposed.

Project IDs and task targets must resolve in Moon's graph. Installable paths are
resolved relative to the workspace, canonicalized, constrained to the
workspace root, and must include an explicit devShell attribute after `#`.
Remote/unlocked installables and inline Nix expressions are rejected in 0.1.

The central workspace flake exposes the composition boundary:

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

The extension serializes the resolved flake path and component set into a fixed
expression equivalent to an attribute set containing
`flake.lib.mkMoonShell builtins.currentSystem names`; `nix develop` consumes
that derivation as the `moon` installable. Component names come only from the
schema-validated extension mapping and are never interpolated as raw Nix code.

### Version ownership and scoped pinning

The extension pin and executable pins stay independent:

| Item                            | Owner and pin                                          |
| ------------------------------- | ------------------------------------------------------ |
| Extension WASM                  | Exact `moon_nix_extension-v<version>` locator          |
| Default tool family             | Central Nix component such as `nodejs_24` or `go_1_26` |
| Exact default build             | Central `flake.lock` nixpkgs/overlay revisions         |
| Project/task version exception  | A different named component selected by `replace`      |
| Independent project environment | Named project flake devShell plus its own `flake.lock` |
| Moon/Proto version              | `null` for every executable owned by Nix               |

A project that needs Node 20 selects `node20`; a more-specific target can
replace that with `node24`. Package attributes express the intended family and
the applicable lock file supplies the exact revision and hashes. Reviews update
the Nix expression and lock together. A project or target must not set
`node.version`, `go.version`, or an equivalent Moon version while its resolved
Nix environment supplies the same executable.

Moon supports custom project metadata in `moon.yml`, but `ProjectFragment` does
not pass that arbitrary object to task extension hooks. Loading it separately
would defer validation until execution. Version 0.1 therefore keeps
`environmentByProject` in the typed workspace extension config; a later plan
may add a project-local declaration only if it preserves equivalent schema,
hashing, and diagnostic behavior.

## Proposed approach

1. **Shared runtime boundary** — create a private Rust library under
   `packages/moon/` and move pure Nix discovery, workspace-path resolution,
   guards, safe serialization, and command/script wrapping behind input/output
   types that do not depend on whether Moon invoked a toolchain or extension.
   Keep selector and lifecycle-only behavior in the current toolchain adapter,
   statically link the shared crate into each WASM, and prove no observable
   0.6.1 behavior changes or runtime plugin dependency.
2. **Extension package** — add `packages/moon/moon-nix-extension` with Moon
   extension registration, a schema-validated config, global command/script
   hooks, mapping-gated `baseComponents`, full native-toolchain component
   collection, deterministic
   deduplication, project/task precedence and override unions, safe local
   installable resolution, fixed-expression generation, central-flake
   composition, and extension-specific PDK tests.
3. **Cache and adoption contract** — add an integration fixture that registers
   the local WASM extension without selecting `nix`, supplies native toolchains
   from a small central flake, composes single and polyglot environments,
   exercises project/task append and replace precedence plus a project-owned
   locked devShell, exercises cached and uncached command/script tasks, proves
   the toolchain plugin is absent, and demonstrates the required inherited
   inputs plus `version: null` setup.
4. **Release and rollout** — inherit the existing `moon-plugin` build/release
   tasks, document the pinned GitHub locator, validate the 0.1.0 dry run, merge
   a dedicated version-packages PR to publish the tag/assets, then migrate one
   Xonovex consumer to extension-only configuration in a separate PR before
   proposing extension-only Drodan adoption. Document the toolchain plugin's
   retained special-purpose role and do not migrate other existing consumers
   implicitly.

## Risk assessment

- **Global blast radius — high:** every configured extension receives task
  hooks. Mapping-gated activation, no-op defaults, fail-closed Nix/component
  errors, and a local consumer fixture prevent an incomplete environment from
  running a task silently.
- **Cache staleness — high:** extensions cannot contribute arbitrary contents
  to Moon's task hash today. Required inherited Nix/config inputs make the
  limitation testable; scoped installables additionally require their project
  `flake.nix`, `flake.lock`, and Nix sources. The README must not claim parity
  with the toolchain plugin's hash hook.
- **Regression during extraction — medium:** moving proven logic can alter
  quoting, workspace-path resolution, or host guards. Existing toolchain tests
  remain the compatibility oracle and must pass before the extension adapter
  lands.
- **Expression injection or impurity drift — high:** dynamic composition uses a
  local mutable flake through `--impure`. The implementation owns one fixed
  expression template, uses a tested Nix-string serializer for paths and
  component values, rejects invalid config values before wrapping, and never
  accepts raw Nix from configuration or task metadata.
- **Override ambiguity — high:** append/replace and installable precedence can
  otherwise make the active environment surprising. A typed mutually exclusive
  union, full-target keys, canonical resolution traces, and explicit errors for
  append-over-installable make the final environment explainable and testable.
- **Path escape or unlocked environment — high:** a project/task installable
  could bypass the central lock. Version 0.1 accepts only canonical
  workspace-contained `path:` flakes with explicit attributes, rejects remote
  and parent-traversing references, and requires the selected flake plus lock in
  task inputs.
- **Per-task evaluation cost — medium:** Moon invokes extension hooks per task,
  not once for the whole action graph. Stable sorting makes identical component
  sets evaluate identically, while Nix evaluation/store caching and the pilot
  fixture measure whether repeated `nix develop` startup is acceptable.
- **Proto and Nix both active — medium:** a non-null native toolchain version
  still triggers Proto setup before the extension runs. Documentation and the
  fixture explicitly verify `version: null` at workspace and project scope;
  extension task hooks receive no merged native-toolchain config, so they cannot
  detect or prevent redundant setup from their lifecycle position.
- **Multiple extension ordering — medium:** Moon 2.3.5 evaluates peer extension
  hooks from the same input and applies their replacement outputs afterward, so
  an environment sentinel cannot compose arbitrary command replacements.
  Public configuration allows only one command-replacing extension for mapped
  tasks; an environment-only peer extension remains compatible.
- **Unsupported dual-plugin configuration — medium:** running both Nix plugins
  would obscure which lifecycle owns wrapping and cache behavior. Public
  examples, fixtures, and the pilot are extension-only, documentation marks the
  plugins as mutually exclusive, and a later consumer migration replaces the
  old configuration atomically in one reviewed PR.
- **Release discoverability — low:** GitHub locators are cached and rate
  limited. Documentation pins the exact tag and repeats the existing
  `GITHUB_TOKEN` CI guidance; checksums accompany every WASM asset.

## Child plans and execution groups

1. **01-shared-nix-runtime** — extract the private core and retain complete
   `moon_nix_toolchain` 0.6.1 compatibility while proving both WASM artifacts
   are independently deployable.
2. **02-global-extension-adapter** — build and test the extension-facing schema,
   toolchain/project/task precedence, override union, deterministic composition
   expression, safe workspace-local installables, and wrappers. Depends on 01.
3. **03-cache-contract-and-consumer-fixture** — prove transparent activation,
   central-flake composition, scoped version overrides, project-owned locked
   devShells, Proto opt-out, mutually exclusive standalone deployment, cache
   inputs, and failure behavior. Depends on 02.
4. **04-release-and-pilot-rollout** — finish public docs and changelog, validate
   inherited release tasks, publish 0.1.0 through the version-packages PR, and
   consume the pinned release in a separate extension-only Xonovex PR. Record
   the plugin-role policy and retained toolchain use cases in both plugin
   READMEs. Depends on 03.

The four groups are intentionally sequential: the public adapter depends on a
stable shared core, the fixture must test the final adapter contract, and the
release must publish only the artifact proven by that fixture.

## Success criteria

- A fixture task detected as `javascript`, `node`, and `npm`, with no `nix`
  toolchain selected, activates `baseComponents: [general]`, resolves
  `[general, node]`, and runs inside the central flake's dynamically composed
  environment.
- A polyglot fixture detected as `go`, `javascript`, `node`, and `npm` resolves
  the stable, duplicate-free component set `[general, go, node]`, and both Go and Node
  executables are supplied by the resulting environment without a predeclared
  `node-go` devShell.
- A project-level `replace` selects `[general, node20]` for all of that project's
  tasks, a project `append` adds its component to the inferred set, and a
  full-target task override deterministically appends or replaces after the
  project scope.
- A task can replace its project's Node 20 component with Node 24 without
  enabling Proto; the applicable central `flake.lock` pins the exact builds.
- A project/task installable runs the named devShell from its workspace-local
  flake and lock. Inline expressions, remote/unlocked references, path escapes,
  mixed installable/component objects, and append-over-installable fail with
  deterministic diagnostics.
- An unmapped task remains byte-for-byte unchanged; a missing component,
  malformed mapped name, absent `lib.mkMoonShell`, or Nix evaluation failure
  produces a deterministic fail-closed diagnostic.
- Command argv and opaque scripts preserve arguments, quoting, exit codes,
  working directory, and environment behavior across Linux and the supported
  WASI test harness.
- Workspace-flake paths containing special characters, Nix expression/string
  escaping, `builtins.currentSystem`, missing Nix, `IN_NIX_SHELL`, and
  duplicate-wrap sentinels have explicit passing tests.
- Native Moon toolchains retain project/dependency behavior, and the fixture
  verifies that `version: null` avoids Proto-managed installation while Nix
  supplies the executable.
- Editing the fixture's central or selected project flake lock, shell
  definitions, or extension config invalidates the task cache through declared
  inputs; unrelated edits do not.
- Every existing `moon_nix_toolchain` test remains green after shared-core
  extraction, and its public schema, selector precedence, tag, and artifact name
  remain unchanged.
- The integration fixture contains no `nix` entry or plugin locator in
  `.moon/toolchains.yml`; removing the toolchain-plugin WASM/cache does not
  change extension behavior, proving there is no runtime companion dependency.
- Both READMEs state that new Xonovex/Drodan consumers use the extension alone,
  identify the toolchain plugin's remaining hashing/setup/project-flake use
  cases, define the two plugins as mutually exclusive, and state that retirement
  requires a separate plan after lifecycle-hook parity and consumer migration.
- `npx moon run moon-nix-extension:ci-check` passes build, tests, clippy,
  rustfmt, optimization, stripping, and WASM validation.
- `moon-nix-extension:github-check` and
  `moon-nix-extension:github-publish-dry-run` identify
  `moon_nix_extension-v0.1.0` and the expected `.wasm` plus `.sha256` assets.
- A merged version-packages PR publishes those assets; a later consumer PR pins
  the exact `github://xonovex/platform/moon_nix_extension@...` locator and
  passes its full Moon CI before any Drodan migration is proposed.

## Estimated effort

Four subplans over approximately 7–11 focused engineering days: 1–2 days for
the shared-core extraction, 3–4 days for the extension, scoped override schema,
safe installables, composition expression, and tests, 2–3 days for the
integration/cache fixture, and 1–2 days for release qualification and the pilot
consumer PR. Upstream Moon changes are excluded.
