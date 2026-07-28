# Configuration

- Config is the base of `config -> shared -> agent`: a config package depends only on its siblings here, never on `shared`, `agent`, or `script`.
- Pin exactly every plugin, preset, or parser a config package configures, so each consumer resolves one identical version; `@types/*` and other packages the config does not configure keep a caret range.
- Every package is `layer: configuration` and tagged `npm` and `tenant-shared`. The `ts-config-*` packages ship JSON and take `language: other` with the `tsconfig` tag; the rest are TypeScript and take the `typescript-config` tag.
- Consumers resolve these packages through `package.json` exports, so `"import"` must precede `"node"` in the export map: jiti picks the first matching condition in key order, and `"node"` resolves to `dist/` which may not be built yet. `eslint-config-base` and `eslint-config-cli` document the rest of that resolution order.
