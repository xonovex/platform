# ESLint Config Base

- `"import"` must appear before `"node"` in `package.json` exports — jiti resolves conditions in key order
- `"import"` → `src/index.ts` (no build); `"node"` → `dist/src/index.js` (requires build)
- Published with `src` in `"files"` so source path resolves on npm
- Uses own config via relative import (`./src/index.ts`) to avoid project service conflict
- `eslint.config.ts` excluded from `tsconfig.json` (`allowImportingTsExtensions` required, not emitted)
