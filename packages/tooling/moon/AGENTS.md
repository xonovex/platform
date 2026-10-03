# Moon Plugins

- For a version bump, update `Cargo.toml` and `Cargo.lock`, add the exact `## <version>` CHANGELOG header required by `github-check`, then update each `@<plugin>-v<version>` pin in `.moon/toolchains.yml` or `.moon/extensions.yml` only after the release tag exists.
- Release only through a reviewed PR whose title contains `version packages`; merging to `main` runs `.github/workflows/release.yml`. Never publish directly or update consumer pins before tag assets exist.
- Every workspace uses exactly one Xonovex Nix plugin.
- Neither Nix plugin replaces the other; both are maintained, and a workspace picks by trade-off (the comparison table lives in both READMEs).
- `moon_nix_extension`: typed per-project/task overrides in one validated file, central component composition through `lib.mkMoonShell` or explicit installables (`path:` self-contained copy, `dir:` through the workspace git tree), lazy realization; the consumer declares central and project Nix cache inputs because an extension cannot contribute task hash contents.
- `moon_nix_toolchain`: explicit `nix`/task/toolchain/tag/language selectors, auto-discovered composing project flakes, eager `setup_environment`, and automatic exact cache hashing through `hash_task_contents`.
- Never configure both plugins, even temporarily; switching in either direction is one atomic reviewed PR. The extension does not support arbitrary peer-command replacement.
