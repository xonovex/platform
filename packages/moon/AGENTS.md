# Moon Plugins

- Version bump: bump `Cargo.toml` + `Cargo.lock` and add a `## <version>` CHANGELOG header (the `github-check` task matches it exactly), then bump each `@<plugin>-v<version>` consumer pin in `.moon/toolchains.yml` or `.moon/extensions.yml` once the release tag exists.
- Release new versions only through a reviewed PR whose title contains `version packages`; merging that PR to `main` drives `.github/workflows/release.yml`. Never publish a Moon plugin directly or update a consumer pin before the tag and assets exist.

## Nix plugin roles

Every workspace uses exactly one Xonovex Nix plugin.

| Capability      | `moon_nix_extension`                                                             | `moon_nix_toolchain`                                                                 |
| --------------- | -------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------ |
| Recommended use | New consumers                                                                    | Compatibility or special-purpose consumers                                           |
| Task selection  | Detected native toolchains plus typed project/task overrides                     | Explicit `nix` toolchain selection plus task, toolchain, tag, and language selectors |
| Environment     | Dynamically composes central, mapping-gated components through `lib.mkMoonShell` | Selects a workspace or automatically discovered project devShell                     |
| Realization     | Lazy, when an active task runs                                                   | Eager, through `setup_environment`                                                   |
| Cache contract  | Consumer declares central and project Nix inputs                                 | Plugin adds resolved flake, shell, and lock data through `hash_task_contents`        |

- Migrate a workspace atomically in one reviewed PR: remove explicit `nix` selection, add the extension locator and cache inputs, set Nix-owned native toolchain versions to `null`, pass full CI, and remove the old pin only when the whole workspace no longer needs it.
- Do not configure both plugins, even temporarily. Arbitrary peer command replacement is unsupported by the extension.
- Keep `moon_nix_toolchain` supported until a separate retirement plan confirms hook parity and all consumers have migrated.
