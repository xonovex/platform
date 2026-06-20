# @xonovex/moon-nix-toolchain

## 0.2.0

### Minor Changes

- Select the flake devShell per project via the `shell` toolchain config setting: `shell: go` wraps a project's tasks in `nix develop <root>#go` instead of the default devShell. An unset, empty, or `default` shell keeps the flake's default devShell, so existing consumers are unaffected.

## 0.1.0

### Minor Changes

- [`4c7ceca`](https://github.com/xonovex/platform/commit/4c7ceca) - feat(moon-nix-toolchain): add toolchain plugin that wraps tasks in the nix flake
