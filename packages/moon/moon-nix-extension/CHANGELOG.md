# @xonovex/moon-nix-extension

## 0.3.0

### Minor Changes

- Support Moon 2.5 host paths. The extension resolves the physical workspace through Moon's typed host-path mapping, so composed and explicit Nix environments receive the checkout path instead of Moon's guest `/workspace` path.

## 0.2.0

### Minor Changes

- Accept `dir:./workspace-relative#devShell` installables beside `path:`. A `dir:` installable is handed to `nix develop` as a bare directory reference, so nix resolves it through the enclosing git tree: the project flake may compose from shared workspace files through relative inputs (for example `path:../../../nix`), where a `path:` installable is copied into the store alone and must be self-contained. Both forms stay workspace-contained and require a committed `flake.lock`.

## 0.1.0

### Minor Changes

- Add a global Moon extension that maps detected native task toolchains to centrally composed Nix environment components without selecting a `nix` toolchain.
- Support typed project and full-target task overrides, including locked workspace-contained project-flake installables.
- Keep realization lazy and make cache ownership explicit: consumers declare central and project Nix inputs because extensions cannot contribute task hash contents.
- Ship the extension as a standalone WASM plugin that does not require the Nix toolchain plugin artifact or configuration.
