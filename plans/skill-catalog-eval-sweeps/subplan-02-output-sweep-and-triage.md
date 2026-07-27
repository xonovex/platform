---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-eval-sweeps.md
parallel_group: 2
status: complete
dependencies:
  plans: [subplan-01-trigger-sweep-and-triage]
  files:
    - packages/skill/
    - packages/script/script-moon-skill-eval-outputs/
    - budgets.json
skills_to_consult: [skill-guide, testing-guide, moon-guide, git-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: partial
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

- [x] 72/72 skills swept with valid runs (no `invalid-run.json`)
- [x] `plan-guide` gate resolved: eight prompts carried a cited bucket-1 fix and
      the re-run gate is recorded (0.205 -> 0.477, still FAIL, residual stated)
- [~] Every keep-tier gate PASS or its failure carries a cited disposition —
      45 PASS, 27 FAIL, all dispositioned; 22 of the 27 need a gate-policy
      decision rather than a catalog edit
- [x] Golden eval still passes after all triage edits (0.916, anchor 0.833)
- [x] All `--max-turns` exceptions recorded per skill with rationale

## Files Modified/Created

- Modify: `evals.json`, `SKILL.md`, and references in failing skills;
  `budgets.json` for explicit bumps;
  `packages/script/script-moon-skill-eval-outputs/` only for bucket-4 findings
- Create: execution-record appendix in this file

## Dependencies

- `subplan-01-trigger-sweep-and-triage` complete (sequencing rule above).

## Estimated Duration

Unattended sweep hours to a day + 1–3 triage sessions.

## Execution Record (2026-07-27)

### Commits

| SHA | Change |
| --- | --- |
| `a3a42779` | supplied the artifacts eight `plan-guide` output-eval prompts referenced; retargeted eval 1 at a path that exists |
| `6d2e0782` | the 10000-character generation ceiling applies to codex only, matching the caps the harness reports |

### Sweep

Full catalog, `ci-skill-eval-output` one moon invocation per skill under `xargs -P3`
(the task itself runs two concurrent generations, so six model calls in flight —
the bound calibrated in subplan 1). 71 skills in 7050s; `plan-guide` ran
separately because its evals were being edited.

| | |
| --- | --- |
| Benchmarks | 72/72 |
| Invalid runs | 0 (6 before the harness fix) |
| Gate PASS | 45 |
| Gate FAIL | 27 |
| Golden eval (`pull-request-guide` eval 5) | 0.916, anchor 0.833 |

319 evals at 2 runs per arm is 1276 generations plus 1276 judge calls. The
harness reports no per-run cost, so spend is bounded by the $0.10 per-call cap
rather than measured — a $255 ceiling that actual spend is well under, since most
generations finish far below the cap. The pilot estimate of roughly 3 minutes per
skill held: 71 skills in just under two hours at six concurrent calls.

### `--max-turns` exceptions

- `skill-guide`: `MAX_TURNS=20`. Eval 3's with-skill arm exhausted the default 12
  turns on every attempt. Recorded as a per-skill exception.
- `plan-guide`: none in the final run. An earlier attempt at 24 turns exhausted
  the `$0.10` generation budget instead, which is what identified eval 1's real
  defect — it named `packages/api`, absent from this workspace, so the search ran
  until some cap stopped it. Retargeted, it completes at the default.

### Findings and triage

**Bucket 4 — the generation ceiling invalidated five skills.** `content-guide`
eval 3, `github-guide` eval 5, `gitlab-guide` eval 5,
`gpu-rendering-vulkan-guide` eval 1 and `workflow-guide` eval 9 each reported
`output-limit` on the with-skill arm on every attempt, discarding the whole
benchmark. A 10000-character ceiling was applied to both harnesses although the
run summary only advertised it for codex; claude already carries a spend cap and
a turn cap that bind sooner, so the ceiling only threw away long answers — and
length is what a thorough skill produces, so it penalised the arm under test.
Fixed in `6d2e0782`; all five re-ran valid.

**Bucket 1 — `plan-guide`'s prompts named artifacts the eval never supplied.**
Eight of eleven referenced an inline plan, attached feedback, supplied evidence
or handoff notes that were absent, so the model asked for them and every
assertion failed. Fixed in `a3a42779`; the gate moved 0.205 -> 0.477 with valid
runs. Catalog-wide this defect is rare — a scan of all 319 prompts found 16
candidates across 11 skills, most of them false positives that carry their own
context.

### Open: the 27 gate failures

None is an invalid run, and none fails the trigger check — every skill fires on
its own evals. They split cleanly, and neither group is fixed by editing a skill.

**22 fail the absolute bar while measurably working.** `moon-guide` scores 0.75
against 0.19 without, `skill-guide` 0.77 against 0.20, `gitlab-guide` 0.71
against 0.16, `github-guide` 0.54 against 0.04, `pi-guide` 0.50 against 0.00.
The delta is large and positive; the tier's absolute floor is what they miss.
Most are the provider- and harness-facing guides, whose evals ask for
product-specific facts a bare model does not have — which is precisely why the
without-skill arm sits near zero. Whether a skill that lifts a model from 0.00 to
0.50 should fail its gate is a policy question about the floor, and the parent
plan forbids relaxing a gate to make a sweep pass, so nothing was changed.

**5 fail the delta while the bare model is already competent.**
`code-quality-guide` 1.00 against 1.00, `docker-guide` 0.94 against 1.00,
`shell-scripting-guide` 0.92 against 1.00, `c99-guide` 0.75 against 0.79,
`oop-guide` 0.75 against 0.75. These evals do not discriminate: they measure
competence the base model already has, so no skill edit can produce a delta.
Re-authoring them around what each skill uniquely adds is the fix, and it is a
design change rather than a repair.

`plan-guide` sits in the first group with a negative delta after its prompt fix,
and its residual is the same shape as the second: nine of eleven evals now score
identically with and without the skill.

### Constraint recorded for later triage

136 of the 319 output-eval prompts are byte-identical to a trigger query in some
`eval-queries.json`. Editing one of those prompts silently re-targets the trigger
eval as well, so any such edit has to be made and re-measured in both gates
together. `plan-guide`'s eleven are not among them, which is why its prompts
could be rewritten freely.

### Validation

- `npx moon run '#skill:skill-validate'` — 176 tasks, green
- `npx moon run script-moon-skill-eval-outputs:ci-check` — green
- `npx moon run script-moon-skill-eval-triggers:ci-check` — green
- 72/72 benchmarks valid, zero `invalid-run.json`
- Golden eval 0.916, above its 0.833 anchor
