---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 5
status: pending
dependencies:
  plans:
    - subplan-02-catalog-cull
    - subplan-04-workflow-core-simplification
    - subplan-05-process-skill-grounding
    - subplan-06-anti-drift-enforce
  files:
    - plans/skill-catalog-stabilization.md
skills_to_consult:
  [versioning-guide, npm-guide, git-guide, github-guide, reflect-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 7: release-and-dogfood

## Objective

Close the repo-vs-installed gap through the reviewed release flow, then run the
dogfood window that proves stability against the parent's Definition of Done.

## Context carried from parent

- Decision 5 (exit criterion): all 12 operations used ≥ once across ≥ 2 project
  families, plus two consecutive weeks with zero transcript-evidenced catalog
  failures.
- Standing rule during the window: every catalog change cites a failing eval or a
  transcript. No speculative hardening.
- Execute smells to track: invocations without an antecedent; execute share far
  above peer operations.
- High-impact gates (release PR review, stability declaration) stay mandatory-human.

## Tasks

1. Prepare the version-packages PR covering all changed and deleted packages;
   hand it to review (never bypass branch protection; merging `main` runs
   `.github/workflows/release.yml`).
2. After release, verify from a fresh session that the installed catalog matches
   the repo: 12 workflow commands, 72 skills, no dangling delegations
   (`/plan-decide`-style references to absent operations).
3. Create the dogfood checklist as `plans/skill-catalog-stabilization/dogfood.md`:
   operation × family coverage matrix, failure log (transcript reference → finding
   → fix), execute-smell tally.
4. Run the window under the standing rule; convert each settled boundary case into
   a boundary eval query in the affected skill's `eval-queries.json`.
5. On meeting the exit criterion, run the plan Validate operation against the
   parent's Definition of Done and Update the parent with the evidence.

## Validation Steps

- Release workflow green; publish verified.
- Fresh-session catalog comparison (task 2) recorded in `dogfood.md`.
- Plan Validate run against the parent at window end.

## Success Criteria

- [ ] Release published through the reviewed version-packages PR
- [ ] Fresh-session installed catalog matches repo exactly
- [ ] Coverage matrix complete: 12/12 operations, ≥ 2 families
- [ ] Two consecutive weeks, zero transcript-evidenced failures
- [ ] Every in-window change cites an eval or transcript
- [ ] Parent plan updated with validation evidence

## Files Modified/Created

- Create: `plans/skill-catalog-stabilization/dogfood.md`
- Modify: package versions repo-wide (via version-packages PR), boundary eval
  queries as cases settle, parent plan status/evidence

## Dependencies

- All prior subplans complete, gates green in enforce mode.

## Estimated Duration

1 session (release) + the dogfood window (calendar weeks 3+, evidence-bounded).
