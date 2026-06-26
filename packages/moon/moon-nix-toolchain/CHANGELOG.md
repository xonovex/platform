# @xonovex/moon-nix-toolchain

## 0.6.1

### Patch Changes

- Fall back to a project flake's `default` devShell when a resolved selector names a devShell the flake does not expose. 0.6.0 routed a project-owned `flake.nix` to `nix develop <projectRoot>#<shell>` whenever a `shellBy*` selector matched, even if that named devShell was absent, so a task wrapped outside a nix shell failed with `does not provide attribute 'devShells.<system>.<shell>'`. The wrap hooks now probe the project flake (`nix eval <root>#devShells`, names only — never building or writing a lock file) and drop the `#<shell>` suffix when the name is absent, so the task runs in the flake's `default` devShell exactly as it did under 0.5.0. Workspace-flake selectors and `hash_task_contents` are unchanged (the cache key still tracks the configured shell and stays independent of `nix`).

## 0.6.0

### Minor Changes

- Define a typed, schema-validated toolchain config. Implements the moon_pdk_api `define_toolchain_config` hook, which registers a JSON schema for `shell`, `shellByTask`, `shellByToolchain`, `shellByTag`, and `shellByLanguage`, so moon validates these keys (unknown key, wrong type) at config-load time instead of silently ignoring them. The plugin now reads a typed `NixToolchainConfig` struct internally rather than probing an untyped `serde_json::Value`; devShell precedence and the lazy project load are unchanged.
- Fail closed for opted-in tasks when `nix` is unavailable. New `failClosedByTag` and `failClosedByLanguage` allowlists name the project tags/languages whose tasks MUST run inside nix; when `nix` is absent for such a task the plugin errors (`nix is required for <project>:<task> …`) instead of silently using host tools. Tasks outside the allowlists keep the no-op fallback, and the `IN_NIX_SHELL` / `MOON_NIX_WRAPPED` double-entry guards still no-op unconditionally. Both allowlists default to empty, so existing consumers are unaffected until they opt in (e.g. drodan's game/C via `failClosedByTag: [cmake]`).
- Route a project-flake task to a named devShell. A project that ships its own `flake.nix` now also runs the shell selectors (`shellByTask` > `shellByToolchain` > `shellByTag` > `shellByLanguage` > `shell`); a match wraps the task in `nix develop <projectRoot>#<shell>` instead of always the project flake's bare default devShell. The selected name must be a devShell exposed by the project flake. Projects whose flake exposes only `default` are unchanged: no selector match (or a `default` value) keeps the bare project flake.
- Bust the moon task cache when the flake a task runs in changes. Implements the tier2 `hash_task_contents` hook: it resolves the same flake root (project `flake.nix` when present, else the workspace flake) and devShell selector the wrap hooks use, then folds the resolved flake root, the selected shell, and the `flake.lock` contents into the task's cache key. Editing `flake.lock` or switching the selected devShell invalidates the dependent tasks' cache; an unrelated edit does not. The key is independent of `IN_NIX_SHELL` / `MOON_NIX_WRAPPED` and of whether `nix` is installed on the hashing host, so it is stable across CI and local runs. Also adds the `setup_environment` hook, which pre-builds the resolved devShell (`nix develop … --command true`, non-blocking) so the first wrapped task is not a cold `nix develop`.

## 0.5.0

### Minor Changes

- Wrap a task in its project's own `flake.nix` when present. If `<project>/flake.nix` exists, the task runs in `nix develop <projectRoot> --command …` (the project flake's default devShell), taking precedence over the workspace flake and the shell selectors (which name shells in the workspace flake, so they do not apply to a project flake). Projects without their own flake are unchanged: the workspace flake plus the resolved devShell. Lets a package pin its own toolchain independently of the workspace flake — detected from the project source over the host, so it auto-applies to every project that ships a flake.

## 0.4.0

### Minor Changes

- Select the flake devShell from the project, not only the task. The toolchain config gains four ordered selectors, most specific first: `shellByTask` (task id) > `shellByToolchain` (a task toolchain id) > `shellByTag` (a project tag) > `shellByLanguage` (the project language) — each falling back to the project-wide `shell`, then the default devShell. `shellByTag`/`shellByLanguage` read the project's tags/language over the host, loaded only when one of them is configured. Renames the previous per-task `shells` map to `shellByTask` (no `shells` fallback). `shellByTag` keys a devShell on a project tag (`go`, `shell`, `kubernetes`), so it auto-applies to every tagged project's tasks without enumerating task ids or relying on a real toolchain id.

## 0.3.0

### Minor Changes

- Add a per-task `shells` map (keyed by task id) to the toolchain config: `shells: { go-lint: go, sh-lint: shell }` selects a flake devShell per task. A task's `shells` entry wins over the project-wide `shell`, and both fall back to the default devShell. Lets one project run different tasks in different shells and auto-applies by task id across all projects.

## 0.2.0

### Minor Changes

- Select the flake devShell per project via the `shell` toolchain config setting: `shell: go` wraps a project's tasks in `nix develop <root>#go` instead of the default devShell. An unset, empty, or `default` shell keeps the flake's default devShell, so existing consumers are unaffected.

## 0.1.0

### Minor Changes

- [`4c7ceca`](https://github.com/xonovex/platform/commit/4c7ceca) - feat(moon-nix-toolchain): add toolchain plugin that wraps tasks in the nix flake
