# Moon plugins

WASM moon toolchain plugins (Rust -> `wasm32-wasip1`), e.g. `moon-nix-toolchain`. Tagged `moon-plugin` (+ `rust`); release tasks are inherited from `.moon/tasks/tag-moon-plugin.yml`.

## Version bump (plugin release)

Each plugin versions independently. To release one:

- Bump `Cargo.toml` `version` and the matching `Cargo.lock` `[[package]]` entry.
- Add a `## <version>` section to `CHANGELOG.md` — `github-check` greps `^## <version>`, so the header must match exactly.
- Validate: `npx moon run <plugin>:{fmt-check,lint,build,test,github-check}` (`github-check` asserts a valid non-empty wasm plus the changelog entry).
- Publish via a `version packages` PR (root `AGENTS.md` Release rule): `release.yml` runs `<plugin>:ci-publish` -> `github-publish` -> the `<plugin>-v<version>` GitHub release (wasm + `.sha256`).
- Consumers pin `github://xonovex/platform/<plugin>@<plugin>-v<version>` in `.moon/toolchains.yml`; bump that pin in a **follow-up** PR once the release tag exists (it cannot resolve before).
