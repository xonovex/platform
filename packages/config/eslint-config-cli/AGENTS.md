# ESLint Config CLI

ESLint configuration for CLI tools and scripts, extending `eslint-config-base`.

## Export Condition Ordering

The `"import"` condition **must** appear before `"node"` in `package.json` exports. This is required for the same reason as `eslint-config-base`: jiti (used by ESLint) checks export conditions in object key order, and `"import"` must match first to resolve to source (`src/index.ts`) instead of compiled output (`dist/src/index.js`).

This matters for `packages/script/` packages which use the `typescript-script` moon tag. That tag removes `^:build` dependencies to break circular dependency cycles, so config packages may not be built when script packages run lint. See `eslint-config-base/AGENTS.md` for full details.
