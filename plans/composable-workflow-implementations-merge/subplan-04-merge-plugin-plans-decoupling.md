---
type: plan
has_subplans: false
parent_plan: composable-workflow-implementations-merge
parallel_group: 2
status: pending
dependencies:
  plans:
  - merge-baseline-and-layout-normalization
  files:
  - packages/command/command-workflow/scripts/validate-documentation.mjs
  - packages/command/command-workflow/docs/**
  - packages/command/command-workflow/moon.yml
  - packages/script/**
skills_to_consult:
- typescript-guide
- command-guide
- moon-guide
- testing-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 04: Plugin `plans/` Decoupling

## Objective

Make the shipped `command-workflow` plugin self-contained: its docs, moon task, and validator must not reference monorepo `plans/` paths, while the plan-traceability checks move to a repo-level task with no coverage loss.

## Tasks

1. **Split the validator.** Partition `validate-documentation.mjs` check groups into (a) package-scoped: command/skill cross-references, docs consistency, evals/SOURCES presence — stays in the package; (b) repo-scoped: traceability artifact checks against `plans/composable-workflow-phases/` — moves to a new script under `packages/script/` (repo convention for moon action scripts) invoked by a root-level moon task. Record each group's check count before and after.
2. **Fix the docs.** Rewrite `docs/validation-and-traceability.md` to describe the traceability artifacts and where they live in the *repository* without relative links that break in an installed plugin (name them; link only package-relative files).
3. **Clean the moon task.** Remove `/plans/composable-workflow-phases/**` from `packages/command/command-workflow/moon.yml` test inputs; the repo-level task declares those inputs instead.
4. **Prove isolation.** Run the package test task with only the package tree available (e.g., from the package dir with cwd-relative paths); it must pass without touching `plans/`.
5. **Prove no coverage loss.** Package-scoped checks + repo-scoped checks ≥ 8,862 baseline total; both tasks wired into CI (moon).
6. **Validate and commit** as `fix(command): decouple command-workflow plugin from repository plan paths`.

## Validation Steps

- `grep -rn 'plans/composable-workflow-phases' packages/command/command-workflow/` → empty.
- Package test passes standalone; repo-level task passes; combined check count ≥ 8,862 recorded.
- `docs/` contains no `../../../../plans/` links; markdown link check green.
- Full lint/format/test green.

## Success Criteria

- [ ] Zero `plans/` references inside the shipped package
- [ ] Repo-level traceability task exists, wired into CI, and passes
- [ ] Combined check count ≥ 8,862 (no silent coverage loss), recorded in the subplan
- [ ] Package test proven to pass in isolation
- [ ] One conventional commit

## Files Modified/Created

- Modified: `packages/command/command-workflow/{scripts/validate-documentation.mjs,moon.yml,docs/validation-and-traceability.md}`
- Created: repo-level traceability validator under `packages/script/` + root moon task wiring

## Dependencies

Subplan 01 only (it edits the same validator for moved paths — sequential via groups). Disjoint from subplan 02; parallel-safe in group 2. Subplan 05 builds on the repo-level validator created here.

## Estimated Duration

0.5–1 day.
