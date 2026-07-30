// Hold vite on the 8.0.x line. Vite 8.1.x tightened the rolldown `vite:oxc`
// transform's tsconfig resolution, which fails to resolve package-name
// `extends` (e.g. "@xonovex/ts-config-cli") and breaks vitest with
// "Tsconfig not found" (vitejs/vite#21852, rolldown/rolldown#8097).
// Runbook to complete the upgrade once fixed: plans/vite8.1-upgrade-hold.md
// Hold typescript on the 6.0.x line. typescript-eslint declares a
// `typescript >=4.8.4 <6.1.0` peer range, and its ts-api-utils dependency
// throws "Cannot read properties of undefined (reading 'Intrinsic')" against
// TypeScript 7, which fails every ts-lint task.
// Hold npm on the 11.x line. npm 12 resolves the optional `typescript >=5.0.0`
// peer of @vitest/eslint-plugin to the newest release instead of reusing the
// pinned typescript, so it demands a second typescript and every moon task that
// installs dependencies fails with ERESOLVE.
module.exports = {
  target: (name) =>
    name === "vite" || name === "typescript" || name === "npm" ? "patch" : "latest",
};
