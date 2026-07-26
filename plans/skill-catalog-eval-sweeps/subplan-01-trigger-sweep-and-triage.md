---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-eval-sweeps.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/
    - packages/script/script-moon-skill-eval-triggers/
    - budgets.json
skills_to_consult: [skill-guide, moon-guide, git-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 1: trigger-sweep-and-triage

## Objective

Calibrate cost and reliability on a pilot set, run the trigger-eval sweep across
all 72 skills, and triage every failure to a taxonomy bucket with a cited fix,
ending with affected skills re-run green.

## Context carried from parent

- Standing rule: every catalog edit cites the failing query in its commit; no
  speculative hardening.
- Taxonomy buckets: eval-quality, content gap, boundary dispute, harness defect.
- Known worklist (8 queries, 5 skills): `hono[LinearRouter]`,
  `hono[Hono or Express + Vitest]`, `zod[Hono JSON request]`,
  `threejs[translate gizmo]`, `user-stories[As a / I want]`,
  `user-stories[SMART]`, `user-stories[three amigos]`,
  `credential-management[Argon2]`.
- Routing-owner invariant gates `skill-validate`: an eval-query edit must not
  strip any skill's only validation-split pairing.

## Tasks

1. Commit the pending harness fixes (output-eval retries, `--max-turns`, plan-doc
   corrections) so sweep results attribute to a revision; record the commit SHA in
   this file's execution record.
2. Pilot: run the five known-failure skills —
   `npx moon run skill-hono:ci-skill-eval-trigger skill-zod:ci-skill-eval-trigger
skill-threejs:ci-skill-eval-trigger skill-user-stories:ci-skill-eval-trigger
skill-credential-management:ci-skill-eval-trigger` — and record per-skill wall
   clock, retry/flake count, and any invalid runs.
3. From the pilot, set the parallelism bound for the full run (moon's
   `--concurrency`, or a batched per-skill loop if moon-level bounding proves
   insufficient against rate limits); record the chosen bound and why.
4. Full-catalog sweep, unattended:
   `npx moon run '#skill:ci-skill-eval-trigger'` under the chosen bound, output
   logged; results land per skill under `.skill-eval-results/trigger/claude`.
   Build the findings table: every failing query and every invalid run.
5. Triage each finding into exactly one bucket and land the fix, starting with the
   8 known queries: reword the eval (bucket 1, stating why the eval is wrong),
   edit description/reference within budget (bucket 2), reassign ownership with
   the query as the loser's negative (bucket 3), or fix the script package at
   source (bucket 4). One commit per finding or per coherent cluster, each citing
   its query.
6. Re-run only affected skills to green (same task, 3 runs, validation split);
   re-run `#skill:skill-validate` and `script-moon-skill-eval-triggers:ci-check`
   to confirm routing-owner pairing and drift gates hold.
7. Record the execution appendix: pilot calibration numbers, findings table
   (query → bucket → change → re-run result), and residual dispositions if any
   finding is deliberately left open (must state why).

## Validation Steps

- `npx moon run '#skill:skill-validate'` (routing owners + drift, enforce mode)
- `npx moon run script-moon-skill-eval-triggers:ci-check`
- Affected-skill `ci-skill-eval-trigger` re-runs green, zero invalid runs

## Success Criteria

- [ ] Harness-fix commit recorded; sweep results attributable to it
- [ ] Pilot record: wall clock, flake rate, chosen parallelism bound
- [ ] 72/72 skills swept with valid runs (no `invalid-run.json`)
- [ ] Every failure bucketed with a cited fix or recorded disposition; the 8
      known queries all dispositioned
- [ ] Affected skills re-run green; `skill-validate` green catalog-wide

## Files Modified/Created

- Modify: `eval-queries.json` in failing skills; skill descriptions/references
  where bucket 2 applies; `budgets.json` for explicit bumps;
  `packages/script/script-moon-skill-eval-triggers/` only for bucket-4 findings
- Create: execution-record appendix in this file

## Dependencies

None within this plan; blocks subplan 2 (content edits here would invalidate
output results for the same skills).

## Estimated Duration

1 pilot session + unattended sweep hours + 1–2 triage sessions.
