# @xonovex/moon-nix-toolchain

A moon toolchain plugin that runs selected tasks inside the workspace's Nix flake dev shell, so flake-pinned tools are used identically across local `moon run`, the git pre-commit hook, and CI — with no per-developer setup.

## What it does

Registers a `Nix` toolchain. Tasks that select it are rewritten to run inside `nix develop <workspaceRoot> --command …`, so binaries resolve from the flake's dev shell instead of the developer's `$PATH`. It is generic — the workspace root is resolved at runtime and it carries no consumer-specific config.

It leaves the task **unchanged** when any guard trips:

- `IN_NIX_SHELL` is set — already in a dev shell (e.g. CI's outer `nix develop`); avoids double-entry.
- `MOON_NIX_WRAPPED=1` — already wrapped (the sentinel it sets on every wrapped task).
- `nix` is not on `PATH` — never hard-fails on a host without nix.

## Usage

Register the plugin in `.moon/toolchains.yml`, pinned to a release tag:

```yaml
nix:
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.4.0'
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

An unset, empty, or `default` value selects the flake's default devShell. `shellByTag` and `shellByLanguage` read the project's tags/language over the host; the project is loaded only when one of them is configured.

```yaml
nix:
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.4.0'
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

## Notes

- The flake must provide every binary a wrapped task runs.
- Locally, each uncached wrapped task enters `nix develop` (sub-second warm; slower on a dirty tree). CI enters once via the outer shell, so the plugin no-ops there.
- The pin is deliberate, like `flake.lock`; bump the `@<tag>` to upgrade. The cargo crate / `.wasm` / release tag use underscores (`moon_nix_toolchain`); the moon project directory uses hyphens.

## License

MIT
