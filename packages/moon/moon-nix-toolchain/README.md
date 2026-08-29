# @xonovex/moon-nix-toolchain

Use this Moon toolchain plugin to run selected tasks in a flake-pinned Nix devShell across local `moon run`, the Git pre-commit hook, and continuous integration.

## Choose one Nix plugin

Choose `moon_nix_toolchain` for automatic exact cache inputs, automatic project-flake discovery, selector-based routing, and eager realization. Choose `moon_nix_extension` for a central component registry, explicit project coverage, and lazy realization. Every workspace must use exactly one plugin, and switching plugins must be one atomic reviewed change.

| Aspect         | `moon_nix_extension`                                                                                                                                              | `moon_nix_toolchain`                                                                                              |
| -------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Task selection | Detected native toolchains mapped to components, plus typed per-project and per-task overrides, all in one validated file                                         | Explicit `nix` toolchain selection, routed by task, toolchain, tag, and language selectors                        |
| Environment    | Centrally composed components through `lib.mkMoonShell`, or an explicit installable (`path:` self-contained copy, `dir:` resolved through the workspace git tree) | Workspace devShell, or a project's own `flake.nix` discovered automatically; the devShell picked by selector      |
| Cache contract | Consumer declares central and project Nix inputs by hand; a Moon extension cannot contribute task hash contents, so a missed input is a silently stale cache      | Plugin folds the resolved flake root, devShell, and `flake.lock` into every task hash automatically and precisely |
| Realization    | Lazy, when an active task runs                                                                                                                                    | Eager pre-build through `setup_environment`, so the first wrapped task is warm                                    |
| New projects   | Must be added to the config; an unlisted project silently runs on host tools                                                                                      | Wrap by tag or language; adding the tag is the whole opt-in                                                       |
| Fail-closed    | One global `failClosed` flag                                                                                                                                      | Opt-in per tag and language allowlists                                                                            |

The toolchain plugin models Nix as an explicit task toolchain and hashes the resolved shell automatically.

## What it does

The plugin registers a `Nix` toolchain and rewrites selected tasks to run inside `nix develop <root> --command ...`. Binaries resolve from the flake's devShell instead of the developer's `PATH`. The plugin resolves the root at runtime and carries no consumer-specific configuration. A project that contains `flake.nix` uses that flake. Other projects use the workspace flake.

It leaves the task **unchanged** when any guard trips:

- `IN_NIX_SHELL` is set and a command task's command exists in that shell: avoids double entry for a sufficient outer `nix develop`. A missing command is wrapped with the selected flake. Script tasks are always wrapped because the plugin cannot reliably infer every command that a script needs.
- `MOON_NIX_WRAPPED=1`: already wrapped by this plugin.
- `nix` is not on `PATH`: does not fail on a host without Nix, unless the project enabled [fail-closed enforcement](#fail-closed-enforcement).

## Usage

Register the plugin in `.moon/toolchains.yml` with an exact release tag, then select it in each project that needs Nix.

```yaml
nix:
  plugin: "github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.8.1"
```

Opt a project in via its `moon.yml` (moon has no global toolchain default, so this is per project):

```yaml
toolchains:
  default: [system, nix]
```

## Selecting a devShell

Select a named devShell when a task needs a smaller, exact toolchain. The plugin uses the flake's `default` devShell when no selector matches. It resolves the merged toolchain configuration from most specific to least specific, and the first matching key wins.

1. `shellByTask`: keyed by task identifier
2. `shellByToolchain`: keyed by a toolchain identifier present in the task's `toolchains`
3. `shellByTag`: keyed by a project tag
4. `shellByLanguage`: keyed by the project language
5. `shell`: a project-wide default set in the project's `moon.yml`

An unset, empty, or `default` value selects the flake's default devShell. `shellByTag` and `shellByLanguage` read the project's tags/language over the host; the project is loaded only when one of them is configured. Every key above is validated against the toolchain's published schema, so an unknown key or wrong type is rejected when moon loads the config.

```yaml
nix:
  plugin: "github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.8.1"
  # Tag-based: every project tagged `go` runs its tasks in `nix develop <root>#go`,
  # without enumerating task ids or relying on a real toolchain id.
  shellByTag:
    go: go
    shell: shell
    kubernetes: k8s
  # Language-based alternative (keyed on the project's language):
  # shellByLanguage: { go: go, bash: shell, yaml: k8s }
  # Toolchain-based (keyed on a task toolchain id):
  # shellByToolchain: { go: go }
  # Per-task override (keyed on task id):
  # shellByTask: { go-lint: go }
```

A project-wide default lives in the project's `moon.yml`:

```yaml
toolchains:
  nix:
    shell: go # this project's tasks use `nix develop <root>#go`
```

Set `GITHUB_TOKEN` in CI so moon's `github://` resolver isn't rate-limited; moon downloads and caches the `.wasm` on first use.

## Per-project flakes

When `<projectRoot>/flake.nix` exists, the plugin wraps that project's tasks with `nix develop <projectRoot> --command ...` instead of the workspace flake. Shell selectors still apply. A matching selector uses `nix develop <projectRoot>#<shell>`, so the project flake must expose the named devShell. Without a match, or with a `default` value, the task uses the project flake's default devShell.

The plugin detects the project flake from project source on the host, so it applies automatically without per-project configuration. Projects without `flake.nix` continue to use the workspace flake and resolved devShell. This lets a package pin its toolchain independently of the workspace flake.

## Fail-closed enforcement

Enable fail-closed enforcement when a task must not fall back to host tools. By default, a host without `nix` runs the task unchanged and does not provide the flake-pinned toolchain.

```yaml
nix:
  plugin: "github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.8.1"
  # Tasks in any project carrying one of these tags MUST run inside nix.
  failClosedByTag: [cmake]
  # Or key the same contract on the project language:
  # failClosedByLanguage: [c, cpp]
```

When `nix` is absent for a task in an opted-in project, the plugin reports `nix is required for <project>:<task> ...` and the task fails instead of using host tools. Tasks in projects outside both allowlists keep the silent no-op. `MOON_NIX_WRAPPED` always prevents another wrap. For command tasks, `IN_NIX_SHELL` prevents wrapping when the command is available in that shell. Script tasks are always wrapped because their complete tool requirements cannot be inferred reliably. Both allowlists are validated against the published schema and default to empty, so existing consumers are unaffected until they opt in.

## Cache coherence

The plugin invalidates a task's Moon cache when the resolved flake or devShell changes. It does not invalidate the cache for an unrelated edit. The `hash_task_contents` hook resolves the same flake root and devShell selector as the wrapping hooks. It adds the resolved flake root, selected shell, `flake.lock`, and the flake's `nix/**/*.nix` modules to the task cache key. The key does not depend on `IN_NIX_SHELL`, `MOON_NIX_WRAPPED`, or the availability of `nix` on the hashing host. The `setup_environment` hook starts a non-blocking build of the resolved devShell so the first wrapped task is not a cold `nix develop`.

When the resolved root is a project flake, the workspace's own `nix/**/*.nix` is also part of the cache key. A project flake usually composes the workspace devShells through a relative `path:` input. Since Nix 2.26, these inputs resolve against the parent source tree instead of an independent locked tree, so they have no `narHash`. Editing a shared module such as `nix/cc.nix` can therefore change the resolved devShell while leaving the project's `flake.nix` and `flake.lock` unchanged. Hashing only `<projectRoot>/**` produced cache hits from the previous toolchain. Shared-module paths are relative to the workspace root, so the key remains stable between continuous integration and a developer's machine.

## Notes

Keep these operational constraints with the plugin configuration.

- The flake must provide every binary a wrapped task runs.
- Locally, each uncached wrapped task enters `nix develop` (sub-second warm; slower on a dirty tree). CI enters once via the outer shell, so the plugin no-ops there.
- The pin is deliberate, like `flake.lock`; bump the `@<tag>` to upgrade. The cargo crate / `.wasm` / release tag use underscores (`moon_nix_toolchain`); the moon project directory uses hyphens.

## License

MIT
