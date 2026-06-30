// Hold vite on the 8.0.x line. Vite 8.1.x tightened the rolldown `vite:oxc`
// transform's tsconfig resolution, which fails to resolve package-name
// `extends` (e.g. "@xonovex/ts-config-cli") and breaks vitest with
// "Tsconfig not found" (vitejs/vite#21852, rolldown/rolldown#8097).
// Runbook to complete the upgrade once fixed: plans/vite8.1-upgrade-hold.md
module.exports = {
  target: (name) => (name === "vite" ? "patch" : "latest"),
};
