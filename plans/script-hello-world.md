---
type: plan
has_subplans: true
status: complete
proposed_subplans:
  - scaffold-package
dependencies:
  plans: []
  subplans:
    - plans/script-hello-world/subplan-01-scaffold-package.md
parallel_groups:
  - group: 1
    parallel: false
    subplans:
      - plans/script-hello-world/subplan-01-scaffold-package.md
skills_to_consult:
  - skill-general-fp
research_sources:
  documentation: []
  versions:
    typescript: "workspace ts-config-cli"
    node: ">=20"
---

# Hello World Script

Add a minimal TypeScript script package at `packages/script/script-hello-world/` that prints `Hello, World!` when invoked, wired into the monorepo via Moon's `typescript-script` tag.

## Goals

- New package `@xonovex/script-hello-world` under `packages/script/script-hello-world/`
- Single `bin` entry `hello-world` prints `Hello, World!` to stdout and exits 0
- Inherits standard typecheck, lint, build, test tasks via `typescript-script` tag
- No runtime dependencies; matches conventions of existing `script-moon-*` packages

## Current State

- Monorepo uses Moon + npm workspaces; TS scripts live under `packages/script/`
- Reference pattern: `packages/script/script-moon-npm-publish/`
  - `src/index.ts` with `#!/usr/bin/env node` shebang
  - `package.json` → `"bin": { "<name>": "./dist/src/index.js" }`, `"type": "module"`, `"private": true`
  - `moon.yml` → `language: typescript`, `tags: [typescript-script]`
  - `tsconfig.json` extends `@xonovex/ts-config-cli`, `outDir: ./dist`
  - Co-located `eslint.config.ts`, `prettier.config.ts`, `vitest.config.ts`
- Moon task inheritance: `.moon/tasks/tag-typescript-script.yml` + `tag-typescript.yml` provide build/test/lint/typecheck

## Research Findings

**No external library needed.** Entry is `process.stdout.write("Hello, World!\n")` (or `console.log`). Rationale:
- Trivial scope — a single line of Node builtin output
- Zero dependencies aligns with monorepo preference for direct imports and minimal surface
- Matches shebang + ESM convention of `script-moon-npm-publish`

Alternatives considered:
- CLI arg parsing via `node:util` `parseArgs` — rejected; out of scope ("literally Hello, World!")
- Shared helpers from `@xonovex/script-moon-common` (e.g., `logInfo`) — rejected; introduces dependency for no benefit

## Proposed Approach

1. **Create package directory** `packages/script/script-hello-world/`
2. **`package.json`** — name `@xonovex/script-hello-world`, `"type": "module"`, `"private": true`, `"bin": { "hello-world": "./dist/src/index.js" }`, no deps
3. **`src/index.ts`** — shebang + `console.log("Hello, World!")`
4. **`moon.yml`** — `language: typescript`, `tags: [typescript-script]`
5. **`tsconfig.json`** — extends `@xonovex/ts-config-cli`, `outDir: ./dist`, `include: ["src"]`
6. **Config files** — `eslint.config.ts`, `prettier.config.ts`, `vitest.config.ts` (re-exports of shared configs, matching siblings)
7. **Root `package.json` workspaces** — verify `packages/script/*` glob covers the new package (it already does)
8. **Verify** — `moon run script-hello-world:ci-check` (or equivalent aggregate), then `node packages/script/script-hello-world/dist/src/index.js` prints the expected output

## Risk Assessment

- **Risk**: Shebang line-ending / executable bit may fail on some shells — *Mitigation*: LF endings; `bin` entry via npm link handles exec bit
- **Risk**: TS references / path config copy-paste drift from sibling package — *Mitigation*: strip unused `paths`/`references` from `tsconfig.json` (no dependency on `script-moon-common`)
- **Open question**: No test target exists beyond smoke-level — leave a trivial vitest or skip? Recommend one minimal vitest asserting the script runs and writes the expected string, to satisfy the inherited `test` task.

## Proposed Child Plans

Single subplan — the work is small and tightly coupled (one package, all files land together):

- **Group 1** (sequential, 1 subplan):
  - `scaffold-package` — create all files, verify `moon run` tasks pass

## Success Criteria

- `moon query projects` lists `script-hello-world`
- `moon run script-hello-world:build` succeeds
- `moon run script-hello-world:typecheck`, `:lint`, `:test` succeed
- Running the built binary prints `Hello, World!` followed by a newline and exits 0

## Estimated Effort

~15–30 minutes. One package, ~6 small files, no external deps.
