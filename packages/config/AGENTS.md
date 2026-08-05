# Configuration

- Config is the base of `config -> shared -> agent`: a config package depends only on its siblings here, never on `shared`, `agent`, or `script`.
- Pin exactly every plugin, preset, or parser a config package configures, so each consumer resolves one identical version; `@types/*` and other packages the config does not configure keep a caret range.
- Every package is `layer: configuration` and tagged `npm` and `tenant-shared`. The `ts-config-*` packages ship JSON and take `language: other` with the `tsconfig` tag; the rest are TypeScript and take the `typescript-config` tag.
- Consumers resolve these packages through `package.json` exports, so `"import"` must precede `"node"` in the export map: jiti picks the first matching condition in key order, and `"node"` resolves to `dist/` which may not be built yet. `eslint-config-base` and `eslint-config-cli` document the rest of that resolution order.
- These packages version as one line together with `shared-core`, and the line contains a cycle: `eslint-config-base` devDepends on `prettier-config` while `prettier-config` depends on `eslint-config-base`, so no per-package order works. Move the line with one `npx moon run workspace-config:version-bump-lockstep -- --lockstep <the ten config packages>,shared-core --type <patch|minor|major>`, previewing with `--dry-run` first.
