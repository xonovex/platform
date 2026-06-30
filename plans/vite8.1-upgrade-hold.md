---
type: runbook
status: blocked-upstream
feature: vite8.1-upgrade
blocked_by: https://github.com/vitejs/vite/issues/21852
---

# Vite 8.1 upgrade — held back

## Summary

Every npm dependency in this repo is upgraded to latest **except** `vite`,
which is held on the **8.0.x** line. Vite 8.1.x breaks vitest in this monorepo.

## Why it is held

Bumping `vite` 8.0.16 -> 8.1.1 makes vitest fail with
`Error: Tsconfig not found .../packages/agent/agent-cli...`. Vite 8.1.x
tightened the rolldown `vite:oxc` transform's tsconfig resolution (via
oxc-resolver), which can no longer resolve the package-name `extends` chain
(`@xonovex/ts-config-cli` -> `@xonovex/ts-config-base`). The error is
misleading — it is a resolution failure, not a missing file. At `8.0.16` the
same test passes (verified: `shared-core` `logging.test.ts`, 14/14).

- Upstream bug: https://github.com/vitejs/vite/issues/21852 (CLOSED)
- Fix merged: https://github.com/vitejs/vite/pull/21932 — **not yet in a
  published release** (latest npm was `vite@8.1.1` at time of writing;
  `8.0.16` is the latest 8.0.x).
- Related: https://github.com/rolldown/rolldown/issues/8097, oxc-resolver
  v11.21.x extends/auto-discovery changes.

## What is currently pinned

- `package.json` devDep `vite`: **`8.0.16`** (latest 8.0.x).
- `.ncurc.cjs`: `target` returns `"patch"` for `vite` (keeps it on 8.0.x),
  `"latest"` for everything else. This file was added by the upgrade — the repo
  had no `.ncurc.cjs` before.
- `vite` is a direct root devDep only; vitest accepts `^8`, so 8.0.16 hoists as
  the single version.

## How to tell the fix has shipped

1. A Vite release **newer than 8.1.1** exists containing PR #21932
   (`npm view vite version`; confirm via changelog / the PR's merged tag).
2. Smoke test before committing the bump:
   ```bash
   cd packages/shared/shared-core
   npx vitest run src/logging.test.ts   # must pass, no "Tsconfig not found"
   ```

## When fixed — steps to complete the upgrade

1. **Lift the guard.** In `.ncurc.cjs`, change the `vite` target from
   `"patch"` to `"latest"` (or delete `.ncurc.cjs` entirely to restore the
   repo's original no-config state).
2. **Bump vite** (and pick up any other new releases):
   ```bash
   npx npm-check-updates -u --workspaces --root --dep prod,dev,optional
   npm install
   ```
3. **Validate** (Go / rust projects are nix-only — exclude them):
   ```bash
   npm run check:all              # lint + typecheck
   npx moon run :build :test
   ```
   The `shared-core` vitest suite passing is the key signal that the oxc
   tsconfig fix is effective.

## Other notes from the upgrade

- `unicorn` 65->69 and `sonarjs` 4.0->4.1 added ~63 new recommended rules;
  they are disabled as a documented block in
  `packages/config/eslint-config-base/src/index.ts` to preserve the baseline.
  Adopt incrementally if desired.
- `@types/node` 25->26, `dependency-cruiser` 17->18, `eslint` 10.6, `vitest`
  4.1.9, `knip` 6.23, etc. all landed and validated green.
