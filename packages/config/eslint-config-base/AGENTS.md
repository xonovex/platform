# ESLint Config Base

Base ESLint configuration for all Xonovex TypeScript projects.

## Export Condition Ordering

The `"import"` condition **must** appear before `"node"` in `package.json` exports. ESLint uses jiti to load config files, and jiti resolves modules via mlly with conditions `['node', 'import']` checked in **object key order**. Placing `"import"` first ensures jiti resolves to `src/index.ts` (source), allowing consumers to use this package without it being built first.

- **`"import"` first**: jiti resolves to `src/index.ts` (source, no build needed)
- **`"node"` first**: jiti resolves to `dist/src/index.js` (requires build)
- **CJS consumers**: Fall through to `"node"` regardless of order
- **Published packages**: `src` is included in `"files"` so the source path resolves on npm too

## Self-Lint

This package uses its own config to lint itself. The `eslint.config.ts` uses a relative import (`./src/index.ts`) instead of the package name to avoid a project service conflict: when loaded from source, the shared config sets `tsconfigRootDir: import.meta.dirname` to the `src/` subdirectory, which prevents the TypeScript project service from finding files. The `tsconfigRootDir` override in `eslint.config.ts` corrects this.

The `eslint.config.ts` is excluded from `tsconfig.json` `include` because the `.ts` extension import requires `allowImportingTsExtensions`, and the file should not be emitted to `dist/`.

## Guidelines

- See [typescript-guidelines](../../guide/guide-typescript/index.md)
