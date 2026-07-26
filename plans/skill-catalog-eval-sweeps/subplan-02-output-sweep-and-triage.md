---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-eval-sweeps.md
parallel_group: 2
status: pending
dependencies:
  plans: [subplan-01-trigger-sweep-and-triage]
  files:
    - packages/skill/
    - packages/script/script-moon-skill-eval-outputs/
    - budgets.json
skills_to_consult: [skill-guide, testing-guide, moon-guide, git-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 2: output-sweep-and-triage

## Objective

Run the output-eval sweep across all 72 skills after trigger triage lands, triage
every gate failure under the taxonomy, and end with every keep-tier gate PASS or a
cited disposition — with the golden end-to-end eval still passing.

## Context carried from parent

- Runs only after subplan 1: content edits from trigger triage would invalidate
  output results for the same skills.
- Gate policy per tier: with-skill pass rate ≥ 0.75/0.8/0.9
  (aggressive/moderate/conservative), delta ≥ 0.05/0.05/0.1, trigger rate = 1.
- `plan-guide` is the known risk: eval 1 completes since the harness fix but
  scored `with_skill` 0.5 on a single-eval run; the full 11-eval gate is unknown.
- `--max-turns` above the default 12 (ceiling 24) is a per-skill recorded
  exception, not a sweep-wide setting.
- The golden eval (`pull-request-guide` eval 5, passed at 0.833) is the
  regression anchor for all triage edits.

## Tasks

1. Pre-flight `plan-guide` alone:
   `npx moon run skill-plan:ci-skill-eval-output`; if any eval still exhausts
   turns, re-run with `-- --max-turns <n>` (≤ 24) and record the exception.
   Triage its gate result first — it is the largest known unknown.
2. Full-catalog sweep, unattended:
   `npx moon run '#skill:ci-skill-eval-output'` under the parallelism bound
   calibrated in subplan 1; per-skill `benchmark.json` lands under
   `.skill-eval-results/output/claude`.
3. Build the findings table from the benchmarks: per skill, gate PASS/FAIL, the
   failing check (pass rate, delta, trigger rate), and any invalid runs (expected
   zero with retries; each one is a bucket-4 finding).
4. Triage each failure into exactly one bucket and land the fix: assertion or
   rubric repair with recorded rationale (bucket 1), deletion-first content edit
   within the ratcheted budget (bucket 2), or script-package fix at source
   (bucket 4). Boundary disputes (bucket 3) are not expected here — the with-skill
   arm invokes the skill explicitly — so any apparent one must be re-examined as
   bucket 1 or 2. One commit per finding or coherent cluster, each citing the
   eval.
5. Re-run affected skills to gate PASS; re-run `skill-pull-request`'s output
   evals if any surface it shares was edited, and confirm the golden eval still
   passes.
6. Record the execution appendix: gate table, dispositions, `--max-turns`
   exceptions, and measured spend/wall-clock actuals against the pilot estimate.

## Validation Steps

- `npx moon run '#skill:skill-validate'` (budgets + vocabulary after edits)
- `npx moon run script-moon-skill-eval-outputs:ci-check`
- Affected-skill `ci-skill-eval-output` re-runs: gate PASS, zero invalid runs
- Golden eval passing in the final `skill-pull-request` benchmark

## Success Criteria

- [ ] 72/72 skills swept with valid runs (no `invalid-run.json`)
- [ ] `plan-guide` gate resolved: PASS, or every failing eval carries a cited
      bucket-1/2 fix and the re-run gate result recorded
- [ ] Every keep-tier gate PASS or its failure carries a cited disposition
- [ ] Golden eval still passes after all triage edits
- [ ] All `--max-turns` exceptions recorded per skill with rationale

## Files Modified/Created

- Modify: `evals.json`, `SKILL.md`, and references in failing skills;
  `budgets.json` for explicit bumps;
  `packages/script/script-moon-skill-eval-outputs/` only for bucket-4 findings
- Create: execution-record appendix in this file

## Dependencies

- `subplan-01-trigger-sweep-and-triage` complete (sequencing rule above).

## Estimated Duration

Unattended sweep hours to a day + 1–3 triage sessions.
