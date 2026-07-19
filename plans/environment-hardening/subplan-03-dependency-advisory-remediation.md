---
type: plan
has_subplans: false
parent_plan: environment-hardening
parallel_group: 1
status: complete
updated: '2026-07-19'
completed_date: '2026-07-19'
dependencies:
  plans: []
  files:
  - package-lock.json
  - plans/composable-workflow-phases/VALIDATION.txt (shared with siblings — own section only)
skills_to_consult:
- npm-guide
- versioning-guide
validation:
  type_check: complete
  lint: complete
  build: complete
  tests: complete
  integration: complete
---

# Subplan 03: Dependency Advisory Remediation

## Objective

Clear the six npm-audit advisories (1 low, 2 moderate, 3 high: @babel/core, form-data, js-yaml, linkify-it, markdown-it, undici — all `fixAvailable: true` as plain booleans) via lockfile-only transitive updates, in one isolated, easily-revertable commit.

## Tasks

1. **Inventory.** `npm audit --json` snapshot: advisory IDs, current and fixed versions, dependency paths. Record in the subplan on completion.
2. **Apply plain updates.** `npm audit fix` first; for anything it leaves, targeted `npm update <package>` on the transitive offenders. No `overrides` entries, no direct-dependency version changes.
3. **Verify scope.** `git diff --stat` shows only `package-lock.json`; machine version sweep confirms all 93 plugin packages remain at 6.0.1; `npm audit` exits 0. If an advisory survives plain updates, stop and record a one-line waiver with rationale and revisit date instead of escalating tooling (Decision 5 contingency).
4. **Gate.** Full `:ci-check` green (`npm ci` runs in CI, so the lockfile must be self-consistent — the pre-commit lockfile hook also guards this).
5. **Record and commit.** Update VALIDATION.txt's advisory note (own section) to the clean state with date; one isolated conventional commit (`fix(deps): clear npm audit advisories via transitive updates`).

## Validation Steps

- `npm audit` exit 0 (or documented waiver).
- Version sweep: one unique version (6.0.1) across all plugin packages and marketplaces.
- `npm ci` succeeds from the updated lockfile; full `:ci-check` green; nothing pushed.

## Success Criteria

- [x] Zero advisories (or explicit waiver with revisit date)
- [x] Diff confined to `package-lock.json` (+ VALIDATION.txt note)
- [x] Plugin lockstep untouched at 6.0.1, machine-checked
- [x] Full gate green; one isolated conventional commit; remote untouched

## Completion

Completed by commit `052865b1` (`fix(deps): clear npm audit advisories via transitive updates`). The commit changed only `package-lock.json` and this plan set's validation note, cleared all six advisories without overrides or direct-dependency changes, and left the then-current 6.0.1 plugin catalog in lockstep. On 2026-07-19, `npm audit` still reports zero vulnerabilities, `npm ci --dry-run --ignore-scripts` accepts the lockfile, and the later 6.1.0 catalog contains one machine-verified plugin version across 95 package and marketplace entries.

## Files Modified/Created

- Modified: `package-lock.json`, `plans/composable-workflow-phases/VALIDATION.txt`

## Dependencies

None (group 1). Shares only VALIDATION.txt with siblings — own section only.

## Estimated Duration

1–2 hours.
