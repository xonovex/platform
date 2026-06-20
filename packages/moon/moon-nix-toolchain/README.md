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
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.2.0'
```

Opt a project in via its `moon.yml` (moon has no global toolchain default, so this is per project):

```yaml
toolchains:
  default: [system, nix]
```

Optionally select a named flake devShell for a project (default: the flake's `default` devShell):

```yaml
toolchains:
  nix:
    shell: go # wraps tasks in `nix develop <root>#go`
```

Set `GITHUB_TOKEN` in CI so moon's `github://` resolver isn't rate-limited; moon downloads and caches the `.wasm` on first use.

## Notes

- The flake must provide every binary a wrapped task runs.
- Locally, each uncached wrapped task enters `nix develop` (sub-second warm; slower on a dirty tree). CI enters once via the outer shell, so the plugin no-ops there.
- The pin is deliberate, like `flake.lock`; bump the `@<tag>` to upgrade. The cargo crate / `.wasm` / release tag use underscores (`moon_nix_toolchain`); the moon project directory uses hyphens.

## License

MIT
