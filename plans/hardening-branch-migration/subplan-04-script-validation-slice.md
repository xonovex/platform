---
type: plan
has_subplans: false
parent_plan: plans/hardening-branch-migration.md
parallel_group: group-2
status: pending
dependencies:
  plans: [plans/hardening-branch-migration/subplan-02-infra-config-slice.md]
  files: [packages/script, .moon/tasks]
skills_to_consult:
  - git-guide
  - github-guide
  - pull-request-guide
  - typescript-guide
  - vitest-guide
  - moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 04: Script Validation Slice

## Objective

Land `packages/script` — the moon action scripts and validators (243 changed
files) that gate everything downstream — plus any `.moon/tasks` gates deferred
from subplan 02. Every validator must pass against main's CURRENT content
before it lands; gates that require not-yet-migrated content are explicitly
deferred to the slice that carries that content.

## Gates Inherited from Subplan 02

Subplan 02 landed the donor's infrastructure but held back six gates that check
`packages/script` content. Re-enable each one in this slice, once the content
it checks is present, and confirm the workspace `ci-check` stays green:

1. `.moon/tasks/tag-skill.yml` — restore `skill-validate` to
   `npx moon-skill-validate-spec --strict` with the
   `script-moon-skill-validate-routing:routing-check` dependency, the
   `XONOVEX_LINT_MODE: enforce` env and the donor's `inputs` list.
2. `.moon/tasks/tag-skill.yml` — return `skill-audit-sources` to `ci-check`'s
   deps and drop its `runInCI: false`, but only if the rewritten script no
   longer reaches the network; `.moon/AGENTS.md` forbids network in `ci-check`.
3. `.moon/tasks/tag-typescript-script.yml` — return `coverage` to `ci-check`'s
   deps.
4. `packages/config/vitest-config-base/src/index.ts` — remove
   `passWithNoTests: true` once every script package has a test file.
5. `packages/config/eslint-config-base/src/index.ts` — restore
   `noInlineConfig: true`, the `no-warning-comments` rule over
   `eslint-disable` / `@ts-ignore` / `@ts-expect-error` / `@ts-nocheck`, and
   `sonarjs/cognitive-complexity: ["error", 30]`.
6. Rebuild `vitest-config-base` and `vitest-config-node` after any config
   change — the compiled `dist/` is what the script packages resolve, so a
   source-only edit silently keeps the old behaviour.

## Tasks

1. **Create the slice branch**:
   `git checkout -b migrate/script-validation origin/main`
2. **Checkout the donor path**:
   `git checkout composable-workflow-platform-hardening -- packages/script`
   and re-apply any `.moon/tasks` templates deferred by subplan 02's list.
3. **Run every validator against main's content**. For each failure, decide:
   (a) main's content is genuinely wrong → fix main's content in this PR;
   (b) the gate assumes branch-only content (e.g. skill/command shapes migrated
   later) → carry the gate in the relevant later subplan instead, and record it
   in that subplan's PR checklist. No gate lands red.
4. **Verify main-side intents survive** (parent plan intent map):
   - `2ea459ef` — script consumes the shared vitest config, no local redefinition
   - `ef22afec` — vitest discovery scoped to `src`, stale `dist` cannot run
   - `b7d68e8a` — skill handoff / manifest-pair / single-ownership guards intact
   - `15b5a21e` — command thin-delegation validator intact
   - `3f04baaa` — skill eval + sources contracts gated in CI
   Also confirm the branch's newer validators (e.g. `61165aa4` em-dash /
   ellipsis / typographic-quote validation) supersede rather than drop these.
5. **Regenerate the lockfile** if `packages/script/package.json` changed
   dependencies: `npm install`, never copy the branch lockfile.
6. **Run full validation**, commit, open the PR with the deferred-gate list.

## Validation Steps

- `npx moon run script-*:typecheck script-*:lint script-*:build script-*:test`
  (confirm project ids via `moon query projects --tags script`)
- Execute each validator task against the workspace
  (`npx moon run :<validator-task>` per template) — all green on main's content
- CI dry-run: the `.github/workflows` checks pass on the slice branch

## Success Criteria

- [ ] All five mapped main-commit intents verified present in the branch's
      newer validator versions
- [ ] Every landed validator green against main's current content
- [ ] Deferred gates enumerated with their target subplan
- [ ] Lockfile regenerated

## Files Modified/Created

- `packages/script/**`, deferred `.moon/tasks/*.yml` templates,
  `package-lock.json` (regenerated)

## Dependencies

Subplan 02. Runs parallel with subplan 03.

## Estimated Duration

4-6 hours (validator-by-validator triage dominates).
