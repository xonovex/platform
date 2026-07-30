# Moon Plugins

- For a version bump, update `Cargo.toml` and `Cargo.lock`, add the exact `## <version>` CHANGELOG header required by `github-check`, then update each `@<plugin>-v<version>` pin in `.moon/toolchains.yml` or `.moon/extensions.yml` only after the release tag exists.
- Release only through a reviewed PR whose title contains `version packages`; merging to `main` runs `.github/workflows/release.yml`. Never publish directly or update consumer pins before tag assets exist.
- Every workspace uses exactly one Xonovex Nix plugin.
- Prefer `moon_nix_extension` for new consumers: it detects native toolchains, accepts typed project/task overrides, lazily composes central mapping-gated environments through `lib.mkMoonShell`, and requires consumers to declare central and project Nix cache inputs.
- Keep `moon_nix_toolchain` for compatibility and special-purpose consumers: it uses explicit `nix`/task/toolchain/tag/language selectors, selects a workspace or auto-discovered project devShell, eagerly runs `setup_environment`, and adds resolved flake, shell, and lock data through `hash_task_contents`.
- Migrate atomically in one reviewed PR: remove explicit `nix` selection, add the extension locator and cache inputs, set Nix-owned native toolchain versions to `null`, pass full CI, then remove the old pin only when no workspace needs it.
- Never configure both plugins, even temporarily; the extension does not support arbitrary peer-command replacement.
- Keep `moon_nix_toolchain` supported until a separate retirement plan confirms hook parity and complete consumer migration.
