# @xonovex/moon-nix-toolchain

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
