# @xonovex/moon-nix-runtime

## 0.1.0

### Minor Changes

- Add the shared Nix runtime library behind both Xonovex Moon Nix plugins. It holds the wrap decision and its double-entry sentinel, flake-target resolution, devShell selection, installable parsing, environment composition, and Nix string escaping, so `moon_nix_toolchain` and `moon_nix_extension` share one implementation rather than each carrying its own.
- Ship the library as a path dependency only. It is not a Moon plugin and exposes no WASM hooks; consumers select a plugin, and the plugin pulls this crate in.
