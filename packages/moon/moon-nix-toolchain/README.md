# @xonovex/moon-nix-toolchain

A moon toolchain plugin that wraps every selected task in the workspace's Nix flake dev shell, so flake-pinned tools are used identically across local `moon run`, the git pre-commit hook, and CI.

## What it does

Registers a toolchain named `Nix`. For every task that selects it, the plugin rewrites the task to run inside `nix develop <workspaceRoot> --command …`, so the task's binaries resolve from the flake's `devShell` PATH rather than from whatever happens to be on the developer's `$PATH`. This closes the "passes locally, fails CI" skew that arises when `moon`, the commit hook, and CI do not share the flake env — moon is the one layer all three share, and the plugin hooks moon's per-task lifecycle (`extend_task_command` / `extend_task_script`) to enforce the flake everywhere with no per-developer setup.

It is generic and consumer-agnostic: the workspace root is resolved at runtime from `MoonContext`, and it carries no consumer-specific naming. Three guards keep it safe to attach broadly:

- `IN_NIX_SHELL` non-empty — already inside a dev shell (e.g. CI's outer `nix develop`), so the task is emitted unchanged. Prevents double-entry and per-task flake re-evaluation under CI.
- `MOON_NIX_WRAPPED=1` — the sentinel the plugin sets on every wrapped task; a task is wrapped at most once (belt-and-suspenders alongside `IN_NIX_SHELL`).
- `nix` not on `PATH` — passthrough. The task runs unwrapped; the plugin never hard-fails on a host without `nix`.

## Consume it

Two pieces of YAML in the consuming repo. There is no `npm install`; moon downloads and caches the published `.wasm` lazily on the first task.

1. Register the plugin in the workspace `.moon/toolchains.yml` under the key `nix`, pinned to a release tag:

```yaml
nix:
  plugin: 'github://xonovex/platform/moon_nix_toolchain@moon_nix_toolchain-v0.1.0'
```

2. Opt each project in by adding `nix` to its project-level toolchain default in `moon.yml`:

```yaml
toolchains:
  default: [system, nix]
```

The per-project `toolchains.default` is the actual lever: moon has no workspace-global toolchain default, so a registered toolchain only participates in a task when that task (or its project default) selects it. The plugin's hooks fire only for tasks where the `nix` toolchain is attached, so it self-gates with no extra config. Opt-in is per project — in drodan the `game-*` C/cmake projects (`game-advanceddrone`, `game-bin2c`, `game-c2x`, `game-common`, `game-prototype`, `game-worldgen`) set this, while a pure-TypeScript project like `game-worldgen-enrichment` does not and is left unwrapped so it does not pay per-task `nix develop` latency for tools it never uses. Order is irrelevant: `[system, nix]` and `[nix, system]` wrap identically.

Set `GITHUB_TOKEN` in consumer CI so moon's `github://` resolver avoids rate limits.

## How it works

For a command task, the plugin replaces the argv with `nix develop <workspaceRoot> --command <origCmd> <args...>`. For a script task, it wraps the script in a shell layer inside the dev shell: `nix develop <workspaceRoot> --command bash -c '<script>'`. In both cases it sets `MOON_NIX_WRAPPED=1` on the task's env. Before wrapping, `resolve_wrap_root` checks the three guards above and returns the workspace root from `MoonContext.workspace_root`, or `None` to leave the task untouched.

## Caveats

- The flake must provide every binary a wrapped task invokes (clang, cmake, etc.). Wrapping makes those commands resolve from the flake's PATH, so a tool missing from the `devShell` fails the task.
- Per-task `nix develop` latency locally. Outside CI there is no outer dev shell, so each wrapped task enters `nix develop`; first entry realises the whole `devShell` once, and on a dirty working tree the flake eval cache effectively does not engage (`git+file:` flakes warn "Git tree is dirty"), so a local top-level `moon run` can pay multi-second eval per task. Keep `flake.nix`/`flake.lock` committed and GC-root the dev shell to mitigate. This is why adoption is scoped per project.
- On a host without `nix` the plugin degrades to a silent no-op passthrough rather than failing loudly, so non-nix machines quietly get unpinned tools. Wrapping only protects contexts that actually have `nix`.
- The locator pin is deliberate. The full explicit `@moon_nix_toolchain-v0.1.0` tag gives a deterministic pin (moon's `github://` resolver uses the `@tag` verbatim — the project segment is not auto-prepended). Bumping the version is a reviewed change, like `flake.lock`: edit the `@<tag>` in `.moon/toolchains.yml`. A `github://` locator with no `@tag` is cached for about seven days and is not deterministic.
- No config options exist. The plugin accepts no configuration — do not add keys under the `nix:` toolchain entry beyond `plugin`.
- Built against `moon_pdk` / `moon_pdk_api` 2.0.4 and targets moon v2.x. These PDK crates are pre-1.0 and moon documents breaking changes in non-major releases, so a PDK bump may mean re-releasing the `.wasm`.

## Build & release

Built and released from this repo. The package is tagged `[rust, moon-plugin]`, which inherit the task chain: `lint` (`cargo clippy --all-targets -- -D warnings`) and `fmt-check` from the `rust` tag; `build` (`cargo build --release --target wasm32-wasip1`, then `wasm-opt -Os --all-features` and `wasm-strip` per `.wasm`), `test`, `ci-check`, `github-check` (validates `Cargo.toml` name/version, a matching `## <version>` CHANGELOG section, and a valid `.wasm` via `wasm-validate --enable-all`), and `github-publish` from the `moon-plugin` tag. `github-publish` derives the tag `${name}-v${version}` from `Cargo.toml` (here `moon_nix_toolchain-v0.1.0`), writes a `.sha256` per `.wasm`, and runs `gh release create` with release notes from the top CHANGELOG section.

The `v0.1.0` release carries assets `moon_nix_toolchain.wasm` and `moon_nix_toolchain.wasm.sha256`. Note the identity split: the cargo crate, `.wasm`, release-tag prefix, and `github://` project segment all use underscores (`moon_nix_toolchain`), while the moon project id and directory use hyphens (`moon-nix-toolchain`). Use `moon_nix_toolchain` identically for the tag prefix and the locator project segment.

## License

MIT
