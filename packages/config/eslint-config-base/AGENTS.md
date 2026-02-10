# ESLint Config Base

Base ESLint configuration for all Xonovex TypeScript projects.

## Export Condition Ordering

`"import"` must appear before `"node"` in `package.json` exports — jiti resolves conditions in object key order.

- `"import"` first → `src/index.ts` (no build needed)
- `"node"` first → `dist/src/index.js` (requires build)
- CJS consumers fall through to `"node"` regardless
- Published: `src` in `"files"` so source path resolves on npm

## Self-Lint

Uses own config via relative import (`./src/index.ts`) to avoid project service conflict (`tsconfigRootDir` would point to `src/`). `eslint.config.ts` excluded from `tsconfig.json` (`allowImportingTsExtensions` required, not emitted).

## Guidelines

- See [typescript-guidelines](../../guide/guide-typescript/index.md)
