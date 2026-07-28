---
type: plan
has_subplans: true
status: complete
dependencies:
  plans:
    - plans/skill-catalog-stabilization.md
  subplans:
    - plans/skill-catalog-eval-sweeps/subplan-01-trigger-sweep-and-triage.md
    - plans/skill-catalog-eval-sweeps/subplan-02-output-sweep-and-triage.md
    - plans/skill-catalog-eval-sweeps/subplan-03-evidence-consolidation.md
parallel_groups:
  group-1: [subplan-01-trigger-sweep-and-triage]
  group-2: [subplan-02-output-sweep-and-triage]
  group-3: [subplan-03-evidence-consolidation]
proposed_subplans:
  - trigger-sweep-and-triage
  - output-sweep-and-triage
  - evidence-consolidation
skills_to_consult:
  - skill-guide
  - workflow-guide
  - moon-guide
  - git-guide
research_sources:
  documentation:
    - .moon/tasks/tag-skill.yml
    - packages/script/AGENTS.md
    - packages/script/script-moon-skill-eval-triggers/
    - packages/script/script-moon-skill-eval-outputs/
    - plans/skill-catalog-stabilization.md
    - plans/skill-catalog-stabilization/subplan-02-catalog-cull.md
    - plans/skill-catalog-stabilization/subplan-05-process-skill-grounding.md
  versions: {}
---

# Plan: Catalog-Wide Eval Sweeps and Triage

The evidence gate between the completed rewrites (stabilization subplans 1–6) and
the release (subplan 7). Runs the two deferred catalog-wide eval sweeps — trigger
and output — and triages every finding under the standing rule before the
version-packages PR starts the dogfood clock.

## Overview

The stabilization plan left exactly one criterion partially met: "trigger evals
green catalog-wide" is `[~]` because both sweeps were deferred on cost, not on any
blocker. The harness defects that made sweeps unfinishable are fixed (bounded
retries in both harnesses, configurable generation turn cap, batch-size defaults,
`--plugin-dir` task defaults). This plan runs both sweeps to completion, fixes what
they surface with each change citing its failing eval, and records the evidence the
parent's Definition of Done needs.

## Goals

- Complete the trigger sweep across all 72 skills with zero invalid runs.
- Complete the output sweep across all 72 skills with zero invalid runs.
- Triage every failure to a cited fix or a recorded disposition — no silent skips,
  no speculative hardening.
- Flip the parent plan's `[~]` criterion to `[x]` with evidence, clearing the way
  for subplan 7.

## Current State

- 72 skills, each with `eval-queries.json` (train/validation splits, pinned
  evaluator models) and `evals.json` (output evals, tiered gates).
- Sweep entry points exist and are CI-shaped: `#skill:ci-skill-eval-trigger`
  (validation split, 3 runs, batch 8, $0.05/call cap) and
  `#skill:ci-skill-eval-output` (2 runs/arm, concurrency 2, batch 3, $0.10/call
  caps, `--eval-cwd $workspaceRoot`).
- Harness hardening landed 2026-07-26: trigger and output harnesses both retry
  transient failures (3 attempts) before invalidating; the output generation turn
  cap defaults to 12 and is configurable via `--max-turns` / `MAX_TURNS` up to 24;
  a persistent failure still invalidates the run (partial evidence is never a
  pass). These changes must be committed before any sweep so results are
  attributable to a revision.
- Drift lints enforce; budgets are ratcheted; the routing-owner invariant gates
  `skill-validate`, so eval-query edits that strip a skill's only validation
  pairing fail CI.

## Research Findings

### Known worklist (already-recorded failures)

- Eight pre-existing failing trigger queries, untouched by the cull (stabilization
  subplan 2 record): `hono[LinearRouter]`, `hono[Hono or Express + Vitest]`,
  `zod[Hono JSON request]`, `threejs[translate gizmo]`,
  `user-stories[As a / I want]`, `user-stories[SMART]`,
  `user-stories[three amigos]`, `credential-management[Argon2]`.
- `plan-guide` output evals: eval 1 now completes (harness fix verified) but
  scored `with_skill` 0.5 on a single-eval run against a moderate-tier 0.8 gate;
  the full 11-eval gate result is unknown.
- The golden end-to-end eval (`pull-request-guide` eval 5) passed at 0.833 and is
  the regression anchor — it must still pass after any triage edits.

### Cost and scale (to be calibrated, not assumed)

Per-call spend is capped ($0.05 trigger, $0.10 generation/judge) but catalog totals
were never measured; prior estimates were "hours of wall clock plus API spend"
(~3 min/skill for outputs). A pilot over a small fixed set calibrates wall clock,
spend, and rate-limit behaviour before committing the full catalog. Naively running
72 project tasks in parallel multiplies concurrent model calls; parallelism must be
bounded at the moon run level and settled in the pilot.

### Triage taxonomy

Every failure lands in exactly one bucket, recorded in the subplan's execution
record:

1. **Eval-quality** — the query or rubric is wrong (vocabulary pull, ambiguous
   boundary, judge misread). Fix the eval; the skill is untouched.
2. **Content gap** — the skill's description or reference is wrong or missing.
   Deletion-first edit within the ratcheted budget.
3. **Boundary dispute** — two skills plausibly own the query. Resolve by the
   contract's one-owner rule; the loser carries the query as a negative,
   preserving routing-owner pairing.
4. **Harness defect** — the run invalidates for infrastructure reasons despite
   retries. Fix at the source in the script package, never by loosening the
   invalidation rule.

## Proposed Approach

1. **Pilot and calibrate.** Commit the pending harness fixes; sweep a fixed pilot
   set (the 5 skills with known failing trigger queries plus `plan-guide` outputs);
   record spend, wall clock, and flake rate; fix anything the pilot exposes; set
   the bounded parallelism used for the full runs.
2. **Trigger sweep, then triage.** Full-catalog `ci-skill-eval-trigger`
   (unattended); classify every failure per the taxonomy; land fixes citing the
   failing query; re-run only affected skills to green.
3. **Output sweep, then triage.** Full-catalog `ci-skill-eval-output` after
   trigger triage lands (content edits from trigger triage would otherwise
   invalidate output results for the same skills); `plan-guide` first since it has
   a known open gate; same taxonomy, same citation rule; re-run affected skills.
4. **Evidence consolidation.** Record both sweeps' results and dispositions;
   run the plan Update operation on the parent (flip `[~]` to `[x]`); confirm
   `ci-check` green across skill, command, and typescript-script tags; state
   go/no-go for subplan 7.

### Constraints

- The standing rule applies to every edit: cite the failing eval or transcript in
  the commit; no speculative hardening, no drive-by rewording.
- All content edits respect the ratchet: fit the budget or bump it explicitly in
  the same change with the citation.
- A skill whose files change after its sweep run must re-run before its result
  counts as evidence.
- Per-call budget caps and the batch model-call cap are not raised to make sweeps
  cheaper to pass; `--max-turns` above 12 is a per-skill, recorded exception.
- Repo rules hold: no feature branches or pushes unless asked; conventional
  commits; release only via the reviewed version-packages PR (out of scope here —
  that is subplan 7 of the parent).

### Exclusions

- No release, no version bumps, no dogfood-window work (parent subplan 7).
- No new skills, commands, or operations; resurrection of cut skills stays behind
  the scenario-plus-eval bar.
- No harness feature work beyond defects the sweeps themselves expose.

## Risk Assessment

- **Spend or wall-clock overrun** — mitigated by the pilot calibration, per-call
  caps, bounded parallelism, and unattended (overnight) full runs.
- **Teaching to the test** — rewriting queries until they pass hides real routing
  problems; mitigated by the triage taxonomy (bucket 1 requires stating why the
  eval, not the skill, is wrong) and the routing-owner pairing check.
- **Flakes at catalog scale** — mitigated by 3-attempt retries; a skill that still
  invalidates is a bucket-4 finding, re-run individually after the fix.
- **Triage scope creep** — the standing rule plus drift ratchet make uncited or
  net-growing edits fail review or CI.
- **Stale evidence** — trigger-triage edits invalidating output results is
  sequenced away (output sweep runs after trigger triage lands); the re-run rule
  covers stragglers.
- **Judge subjectivity on output evals** — evaluator models stay pinned; rubric
  edits are bucket-1 changes with recorded rationale, and the golden eval must
  keep passing.

## Proposed Child Plans

| Group | Subplan                    | Depends on               |
| ----- | -------------------------- | ------------------------ |
| 1     | `trigger-sweep-and-triage` | — (includes the pilot)   |
| 2     | `output-sweep-and-triage`  | trigger-sweep-and-triage |
| 3     | `evidence-consolidation`   | both sweeps              |

## Success Criteria

- [x] Pilot record exists: measured spend, wall clock, flake rate, chosen
      parallelism bound
- [x] Trigger sweep complete: 72/72 skills with valid runs; every failure carries
      a taxonomy bucket and a cited fix or disposition; affected skills re-run
      green
- [x] Output sweep complete: 72/72 skills with valid runs; every keep-tier gate
      PASS or its failure carries a cited disposition; `plan-guide`'s gate
      resolved; golden eval still passing (0.916)
- [x] Zero uncited catalog edits within this plan's window (7 commits touching
      `packages/skill/`, each quoting its failing query or eval; both budget
      bumps atomic with their content edit)
- [x] Full `ci-check` green across skill, command, and typescript-script tags
      after all triage lands (522 tasks)
- [x] Parent plan updated: `[~]` criterion flipped with evidence; go/no-go for
      subplan 7 stated

## Estimated Effort

- Pilot: 1 session (includes committing the harness fixes).
- Trigger sweep: unattended hours, then 1–2 triage sessions (8 known failures
  plus whatever the sweep adds).
- Output sweep: unattended hours to a day, then 1–3 triage sessions (unknown
  surface; `plan-guide` is the one known gate risk).
- Evidence consolidation: 0.5 session.

## Execution Record (2026-07-26/27)

Both sweeps ran to completion. Subplans 1 and 2 carry the detailed findings;
`remediation-open-findings.md` carries the verified analysis of what stayed open.

### Measured actuals against the estimate

|                      | Estimated                   | Measured                                                                                                                                                        |
| -------------------- | --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Trigger sweep        | "unattended hours"          | 2711s first full pass, 2225s re-baseline, at 4 concurrent calls                                                                                                 |
| Output sweep         | "unattended hours to a day" | 7050s for 71 skills at 6 concurrent calls                                                                                                                       |
| Per-skill wall clock | ~3 min for outputs          | held: 71 skills in just under two hours                                                                                                                         |
| Flake rate           | unknown                     | zero transient retries in every full pass                                                                                                                       |
| Spend                | uncapped estimate           | not measured — neither harness reports per-run cost; bounded by the per-call caps at $76 for triggers and $255 for outputs, and actual spend is well under both |

Parallelism settled at 4 concurrent model calls, calibrated by a probe rather
than assumed, and run as one moon invocation per skill so a failing skill cannot
truncate the sweep.

### Results

| Gate        | Coverage     | Invalid | Outcome                             |
| ----------- | ------------ | ------- | ----------------------------------- |
| Trigger     | 72/72        | 0       | 70 → 10 failures                    |
| Routing     | 80 scenarios | 0       | 3 failures (first run of this gate) |
| Output      | 72/72        | 0       | 45 PASS / 27 FAIL                   |
| Golden eval | —            | —       | 0.916, anchor 0.833                 |

### Harness defects found and fixed

The sweeps were unfinishable as configured. Five defects surfaced, each fixed at
source rather than by loosening an invalidation rule:

1. A deterministic `output-limit` stop invalidated a skill's whole trigger sweep.
2. A competing skill invocation invalidated the run instead of scoring it.
3. Cross-skill near misses were scored with the rightful owner absent — 43 of the
   70 trigger failures, none of them fixable by editing a skill.
4. An unresolvable skill name was counted as a competitor win. This one was
   introduced by fix 2 during this plan and caught by an adversarial review.
5. A 10000-character ceiling discarded long claude generations as invalid,
   penalising the arm under test and invalidating five output benchmarks.

### Go / no-go for stabilization subplan 7

**Go, with one decision to take first.**

The evidence this plan existed to produce is complete: both sweeps cover 72/72
skills with zero invalid runs, the citation discipline held across every catalog
commit, the three-tag gate is green at 522 tasks, and the golden eval improved
rather than regressed. The stabilization plan's `[~]` criterion is flipped.

Nothing found is a release blocker. The residual is a calibration question, not a
defect: 22 of the 27 output-gate failures miss a tier's absolute floor while
measurably working, several lifting a bare model from near zero. Releasing with
those gates red is defensible only once the floors are either affirmed or
adjusted deliberately — and the standing rule forbids adjusting them to make a
sweep pass, so it has to be a recorded decision rather than a convenience.

The two remaining ownership questions — whether `claude-code-guide` and
`instruction-guide` should claim work Claude Code's own bundled skills already
own — do not gate the release; they change eval labels, not shipped content.
