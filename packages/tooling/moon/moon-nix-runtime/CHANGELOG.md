# @xonovex/moon-nix-runtime

## 0.2.0

### Minor Changes

- Add the `dir:` installable scheme beside `path:`. `parse_workspace_installable` accepts `dir:./relative#devShell` and `canonical_installable` renders it as a bare directory reference, which nix resolves through the enclosing git tree in a git workspace: the flake may compose from files elsewhere in the workspace through relative inputs, and only tracked files exist to the evaluation. `path:` semantics are unchanged, and both schemes keep the workspace-containment and explicit-devShell validation.

## 0.1.0

### Minor Changes

- Add the shared Nix runtime library behind both Xonovex Moon Nix plugins. It holds the wrap decision and its double-entry sentinel, flake-target resolution, devShell selection, installable parsing, environment composition, and Nix string escaping, so `moon_nix_toolchain` and `moon_nix_extension` share one implementation rather than each carrying its own.
- Ship the library as a path dependency only. It is not a Moon plugin and exposes no WASM hooks; consumers select a plugin, and the plugin pulls this crate in.
