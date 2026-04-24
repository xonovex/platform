---
type: plan
has_subplans: false
parent_plan: plans/script-hello-world.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files: []
skills_to_consult:
  - skill-general-fp
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Scaffold Package: script-hello-world

Create the `packages/script/script-hello-world/` package from scratch — all config and source files — then verify all Moon tasks pass.

## Objective

Produce a fully working `@xonovex/script-hello-world` npm workspace package that:
- Exposes a `hello-world` bin entry built from TypeScript
- Prints `Hello, World!\n` to stdout and exits 0
- Passes `typecheck`, `lint`, `build`, and `test` via inherited `typescript-script` Moon tasks

## Tasks

### 1. Create `packages/script/script-hello-world/package.json`

```json
{
  "name": "@xonovex/script-hello-world",
  "version": "0.0.0",
  "private": true,
  "type": "module",
  "bin": {
    "hello-world": "./dist/src/index.js"
  },
  "keywords": [
    "xonovex",
    "script"
  ]
}
```

No `dependencies` or `devDependencies` blocks — zero runtime deps; dev tooling is workspace-level.

### 2. Create `packages/script/script-hello-world/moon.yml`

```yaml
$schema: https://moonrepo.dev/schemas/project.json
language: typescript
tags: [typescript-script]
```

No extra tasks needed — `typescript-script` tag inherits build/test/lint/typecheck/ci-check from `.moon/tasks/tag-typescript-script.yml` → `.moon/tasks/tag-typescript.yml`.

### 3. Create `packages/script/script-hello-world/tsconfig.json`

```json
{
  "extends": "@xonovex/ts-config-cli",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "."
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

No `paths` or `references` — unlike `script-moon-npm-publish`, this package has no workspace dependencies.

### 4. Create `packages/script/script-hello-world/src/index.ts`

```ts
#!/usr/bin/env node
console.log("Hello, World!");
```

The shebang must be the very first line (no blank line before it, LF line endings).

### 5. Create `packages/script/script-hello-world/eslint.config.ts`

```ts
export {default} from "@xonovex/eslint-config-cli";
```

### 6. Create `packages/script/script-hello-world/prettier.config.ts`

```ts
export {default} from "@xonovex/prettier-config";
```

### 7. Create `packages/script/script-hello-world/vitest.config.ts`

```ts
import {defineConfig} from "vitest/config";

export default defineConfig({
  test: {
    passWithNoTests: true,
  },
});
```

`passWithNoTests: true` means the `test` task passes with no test files — acceptable for a trivial script. Add a test file only if the inherited task requires one explicitly.

### 8. Register workspace and verify Moon discovers the project

Run from the repo root:

```bash
npm install
npx moon query projects | grep hello-world
```

Expected: `script-hello-world` appears in the project list.

## Validation Steps

Run from the repo root. All must pass before marking complete.

```bash
# Build
npx moon run script-hello-world:build

# Type check
npx moon run script-hello-world:typecheck

# Lint
npx moon run script-hello-world:lint

# Test
npx moon run script-hello-world:test

# Smoke test the output binary
node packages/script/script-hello-world/dist/src/index.js
# Expected output: Hello, World!
```

Alternatively, run all at once:

```bash
npx moon run script-hello-world:ci-check
```

## Success Criteria

- [ ] `npx moon query projects` lists `script-hello-world`
- [ ] `moon run script-hello-world:build` exits 0; `dist/src/index.js` exists
- [ ] `moon run script-hello-world:typecheck` exits 0
- [ ] `moon run script-hello-world:lint` exits 0
- [ ] `moon run script-hello-world:test` exits 0
- [ ] `node dist/src/index.js` (from package dir) prints `Hello, World!` and exits 0

## Files Created

- `packages/script/script-hello-world/package.json`
- `packages/script/script-hello-world/moon.yml`
- `packages/script/script-hello-world/tsconfig.json`
- `packages/script/script-hello-world/src/index.ts`
- `packages/script/script-hello-world/eslint.config.ts`
- `packages/script/script-hello-world/prettier.config.ts`
- `packages/script/script-hello-world/vitest.config.ts`

## Files Modified

None — the workspace `packages/script/*` glob in the root `package.json` already covers the new package.

## Estimated Duration

~15 minutes.
