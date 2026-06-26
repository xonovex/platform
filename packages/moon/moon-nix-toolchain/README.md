# @xonovex/moon-nix-toolchain

A moon toolchain plugin that runs selected tasks inside the workspace's Nix flake dev shell, so flake-pinned tools are used identically across local `moon run`, the git pre-commit hook, and CI — with no per-developer setup.

## What it does

Registers a `Nix` toolchain. Tasks that select it are rewritten to run inside `nix develop <root> --command …`, so binaries resolve from the flake's dev shell instead of the developer's `$PATH`. It is generic — the root is resolved at runtime and it carries no consumer-specific config. A project that ships its own `flake.nix` is wrapped with that flake (see [Per-project flakes](#per-project-flakes)); otherwise the workspace flake is used.

It leaves the task **unchanged** when any guard trips:

- `IN_NIX_SHELL` is set — already in a dev shell (e.g. CI's outer `nix develop`); avoids double-entry.
- `MOON_NIX_WRAPPED=1` — already wrapped (the sentinel it sets on every wrapped task).
- `nix` is not on `PATH` — never hard-fails on a host without nix, **unless** the project opted into [fail-closed enforcement](#fail-closed-enforcement).

## Usage

Register the plugin in `.moon/toolchains.yml`, pinned to a release tag:

```yaml
nix:
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.6.0'
```

Opt a project in via its `moon.yml` (moon has no global toolchain default, so this is per project):

```yaml
toolchains:
  default: [system, nix]
```

## Selecting a devShell

Selecting a named devShell gives a task a lean, exact toolchain (default: the flake's `default` devShell). The plugin resolves the shell from the merged toolchain config using ordered selectors, **most specific first** — the first selector with a matching key wins:

1. `shellByTask` — keyed by task id
2. `shellByToolchain` — keyed by a toolchain id present in the task's `toolchains`
3. `shellByTag` — keyed by a project tag
4. `shellByLanguage` — keyed by the project language
5. `shell` — a project-wide default (set in a project's `moon.yml`)

An unset, empty, or `default` value selects the flake's default devShell. `shellByTag` and `shellByLanguage` read the project's tags/language over the host; the project is loaded only when one of them is configured. Every key above is validated against the toolchain's published schema, so an unknown key or wrong type is rejected when moon loads the config.

```yaml
nix:
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.6.0'
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

When a project ships its own `flake.nix` (i.e. `<projectRoot>/flake.nix` exists), the plugin wraps that project's tasks with the project flake — `nix develop <projectRoot> --command …` — taking precedence over the workspace flake. The shell selectors above still apply: a matching selector routes the task to `nix develop <projectRoot>#<shell>`, so the **named devShell must be exposed by the project flake**. With no match (or a `default` value) the task uses the project flake's default devShell.

The project flake is detected from the project source over the host, so it auto-applies to every project that ships one — no per-project config. Projects without their own `flake.nix` are unchanged: the workspace flake plus the resolved devShell. This lets a package pin its own toolchain independently of the workspace flake.

## Fail-closed enforcement

By default a task on a host without `nix` runs unchanged on host tools — convenient, but it silently drops the flake-pinned toolchain. A project can opt out of that silent fallback so its tasks **must** run inside nix:

```yaml
nix:
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.6.0'
  # Tasks in any project carrying one of these tags MUST run inside nix.
  failClosedByTag: [cmake]
  # Or key the same contract on the project language:
  # failClosedByLanguage: [c, cpp]
```

When `nix` is absent for a task in an opted-in project, the plugin errors with `nix is required for <project>:<task> …` and the task fails, instead of falling back to host tools. Tasks in projects outside both allowlists keep the silent no-op. The `IN_NIX_SHELL` and `MOON_NIX_WRAPPED` guards still take precedence — a task already inside a dev shell (or already wrapped) never fails closed. Both allowlists are validated against the published schema and default to empty, so existing consumers are unaffected until they opt in.

## Cache coherence

Editing the flake a task runs in — `flake.lock`, or switching its resolved devShell — busts that task's moon cache via the `hash_task_contents` hook; an unrelated edit does not. The hook resolves the same flake root (the project `flake.nix` when present, else the workspace flake) and the same devShell selector the wrap hooks use, then folds the resolved flake root, the selected shell, and the `flake.lock` contents into the task's cache key. The key is independent of `IN_NIX_SHELL` / `MOON_NIX_WRAPPED` and of whether `nix` is installed on the hashing host, so it is stable across CI and local runs. The `setup_environment` hook pre-builds the resolved devShell (non-blocking) so the first wrapped task is not a cold `nix develop`.

## Notes

- The flake must provide every binary a wrapped task runs.
- Locally, each uncached wrapped task enters `nix develop` (sub-second warm; slower on a dirty tree). CI enters once via the outer shell, so the plugin no-ops there.
- The pin is deliberate, like `flake.lock`; bump the `@<tag>` to upgrade. The cargo crate / `.wasm` / release tag use underscores (`moon_nix_toolchain`); the moon project directory uses hyphens.

## License

MIT
