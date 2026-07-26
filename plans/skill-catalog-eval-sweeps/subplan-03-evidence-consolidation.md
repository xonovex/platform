---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-eval-sweeps.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - subplan-01-trigger-sweep-and-triage
    - subplan-02-output-sweep-and-triage
  files:
    - plans/skill-catalog-eval-sweeps.md
    - plans/skill-catalog-stabilization.md
skills_to_consult: [plan-guide, moon-guide, git-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 3: evidence-consolidation

## Objective

Turn the two sweeps' results into recorded evidence: verify the citation
discipline held, confirm all gates green, update both parent plans, and state an
explicit go/no-go for stabilization subplan 7.

## Context carried from parent

- The stabilization plan's `[~]` criterion — "trigger evals green catalog-wide" —
  is the one this plan exists to flip; its Progress section also names both
  sweeps as the remaining unrun evidence.
- Zero uncited catalog edits within this plan's window is itself a success
  criterion, checkable from the commit log.

## Tasks

1. Audit the window's commit log (from the harness-fix commit recorded in
   subplan 1 to HEAD): every commit touching `packages/skill/` cites a query,
   eval, or transcript. List any that do not; fix or revert them before
   proceeding.
2. Run the full gate:
   `npx moon run '#skill:ci-check' '#command:ci-check' '#typescript-script:ci-check'`
   — all green in enforce mode.
3. Update `plans/skill-catalog-stabilization.md`: flip the `[~]` success
   criterion to `[x]` with evidence (sweep completion dates, attributable
   revision, findings/disposition counts per bucket); refresh its Progress
   section to say both sweeps ran and what they found.
4. Update this plan's parent document: check off its success criteria, set
   `status: complete`, and record total measured spend and wall clock against the
   pilot estimate.
5. State go/no-go for stabilization subplan 7 in the parent's Progress: either
   "evidence complete, release is unblocked" or the named residual risks that
   should hold it.

## Validation Steps

- Full three-tag `ci-check` green (task 2 output recorded)
- Both parent documents render valid frontmatter and pass
  `npx prettier --check`
- A read-back of the stabilization plan's Success Criteria shows no remaining
  `[~]` other than items owned by subplan 7

## Success Criteria

- [ ] Commit-log audit recorded: zero uncited catalog edits (or each violation
      fixed/reverted)
- [ ] Three-tag `ci-check` green, recorded with task count
- [ ] Stabilization plan updated: criterion flipped with evidence, Progress
      refreshed
- [ ] This plan marked complete with measured actuals
- [ ] Go/no-go for subplan 7 stated in the stabilization plan

## Files Modified/Created

- Modify: `plans/skill-catalog-stabilization.md`,
  `plans/skill-catalog-eval-sweeps.md`, this file's execution record

## Dependencies

- Both sweep subplans complete.

## Estimated Duration

0.5 session.
