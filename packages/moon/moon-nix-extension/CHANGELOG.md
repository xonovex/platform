# @xonovex/moon-nix-extension

## 0.1.0

### Minor Changes

- Add a global Moon extension that maps detected native task toolchains to centrally composed Nix environment components without selecting a `nix` toolchain.
- Support typed project and full-target task overrides, including locked workspace-contained project-flake installables.
- Keep realization lazy and make cache ownership explicit: consumers declare central and project Nix inputs because extensions cannot contribute task hash contents.
- Ship the extension as a standalone WASM plugin that does not require the Nix toolchain plugin artifact or configuration.
