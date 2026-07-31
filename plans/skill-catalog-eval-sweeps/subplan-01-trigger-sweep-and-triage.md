---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-eval-sweeps.md
parallel_group: 1
status: complete
dependencies:
  plans: []
  files:
    - packages/skill/
    - packages/script/script-moon-skill-eval-triggers/
    - budgets.json
skills_to_consult: [skill-guide, moon-guide, git-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: partial
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

- [x] Harness-fix commit recorded; sweep results attributable to it
- [x] Pilot record: wall clock, flake rate, chosen parallelism bound
- [x] 72/72 skills swept with valid runs (no `invalid-run.json`)
- [x] Every failure bucketed with a cited fix or recorded disposition; the 8
      known queries all dispositioned
- [~] Affected skills re-run green; `skill-validate` green catalog-wide —
  `skill-validate` is green; after the harness re-baseline 10 of 70 trigger
  failures and 3 of 80 routing scenarios remain open with recorded
  dispositions

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

## Execution Record (2026-07-26/27)

### Harness commits

Sweep results are attributable to these revisions, landed before any sweep ran:

| SHA        | Change                                                                                            |
| ---------- | ------------------------------------------------------------------------------------------------- |
| `0b023be6` | output-eval transient retries (3 attempts) and configurable `--max-turns` (default 12, max 24)    |
| `e4dd519a` | stabilization subplan 5/6 record corrections; this plan and its subplans                          |
| `8013654b` | trigger output-limit stop scores as a non-trigger (bucket 4, found by the pilot)                  |
| `dcf1bca5` | competing skill invocation scores as a non-trigger (bucket 4, found by the first full sweep)      |
| `9a39ae30` | `--catalog-root` defers cross-skill near misses to the routing evaluator; `ci-routing` runs in CI |

### Pilot calibration

Five known-failure skills, run sequentially, validation split, 3 runs, batch 8:

| Skill                         | Wall clock | Queries | Result                                       |
| ----------------------------- | ---------- | ------- | -------------------------------------------- |
| `skill-hono`                  | 130s       | 6       | 5 pass / 1 fail                              |
| `skill-zod`                   | 102s       | 6       | 4 pass / 2 fail                              |
| `skill-threejs`               | 134s       | 6       | 5 pass / 1 fail                              |
| `skill-user-stories`          | 144s       | 7       | invalidated (bucket 4) → 149s, 3/7 after fix |
| `skill-credential-management` | 145s       | 7       | 6 pass / 1 fail                              |

- Mean ~132s per skill at 6–7 validation queries; 508 validation queries
  catalog-wide, 1524 model calls at 3 runs.
- Transient flake rate: **0**. The only two retries in the pilot were the
  deterministic `output-limit` condition that `8013654b` reclassified.
- Spend is bounded by the per-call cap ($0.05) rather than measured; the harness
  reports no per-run cost, so the catalog ceiling is 1524 × $0.05 = $76.20 and
  actual spend is well under it because runs stop at the first skill invocation.

**Parallelism bound: 4, as one `moon run` per skill under `xargs -P4`.** A
concurrency-4 probe over `skill-fp`, `skill-oop`, `skill-git`, and `skill-npm`
finished 4 skills in 105s with zero transient retries, so the bound is not
rate-limited. Per-skill invocation rather than `moon run '#skill:...' --concurrency 4`
because a failing skill is the expected case here and one `moon` invocation
reports `task_runner::run_failed`; separate invocations guarantee all 72 run and
give one log per skill for the findings table.

### Sweeps

| Run                      | Skills | Queries scored     | Invalid | Retries | Failures | Wall       |
| ------------------------ | ------ | ------------------ | ------- | ------- | -------- | ---------- |
| First (aborted at 17/72) | 17     | —                  | 6       | 20      | —        | —          |
| Second (post-`dcf1bca5`) | 72/72  | 508                | 0       | 0       | 70       | 2711s      |
| Final (post-triage)      | 72/72  | 317 + 196 deferred | 0       | 0       | 12       | 2160s      |
| Routing (first ever run) | —      | 77 scenarios       | 0       | 0       | 9        | concurrent |

### Findings and triage

70 failures across 38 skills split into four causes.

**43 — cross-skill near misses (bucket 4, harness).** `ci-skill-eval-trigger`
loads one catalog skill via `--plugin-dir .`, so a negative whose rationale names
another owner was scored with that owner absent and the target as the only
plausible match. Firing is correct behaviour there. Confirmed by running the
routing evaluator, which loads owner and competitors together, over five of them:
`connascence-guide`, `orthogonal-pattern-guide`, `tdd-guide`, `ddd-guide`, and
`memory-management-guide` each won its own query. Fixed at source in `9a39ae30`:
the trigger evaluator now defers any query the catalog assigns elsewhere and
reports the count as `routing-deferred`, and the `ci-routing` task that judges
them runs in CI instead of being defined and never run.

**15 — queries whose referent the eval never supplied (bucket 1, eval quality).**
Every one drew a request for the missing artifact instead of a routing decision;
one `workflow-guide` run answered "I'll use the workflow guide to record your
decision" while invoking nothing. Fixed in `9d57ec0e` and `eb3938d2` by carrying
the story, tasks, snippet, module, options, criteria, handoff, or PR inline. The
skills were not touched; `user-stories-guide` already named both the story
template and SMART as triggers. Skills holding these queries as negatives follow
the new wording so every routing-owner pairing survives — `routing-check`
caught the break when it did not.

**6 — negatives naming an owner that never claimed the query (bucket 3).** No
routing scenario existed, so neither evaluator could judge them. Fixed in
`80d2904d` by adding each as a validation positive to the skill its rationale
named (`microkernel-pattern`, `tdd`, `vitest`, `plan`, `bdd`). `plan-guide`'s
Kafka-versus-NATS negative was simply mislabelled: settling and recording a
decision is the `plan-decide` operation it owns, so it became a positive.

**2 — content gaps (bucket 2).** `credential-management-guide` fired 2/3 on
Argon2 parameters for end-user passwords, a negative no skill claims, so the
over-fire was its own; `36167cd2` scopes the description positively to
machine-to-machine secrets and records the exclusion as a gotcha, per the routing
lint that reserves skip clauses for the body. `c99-opinionated-guide` scored 0/3
on shaping a C API for Lua and C# bindings and answered "I don't have a skill
that covers this" although `references/cross-language-api.md` exists — only the
description reaches the router, and it named neither bindings nor FFI;
`262f1894` adds the trigger. Both budgets were bumped in the same change
(497→540, 914→929).

### The 8 known queries

| Query                            | Disposition                                                                          |
| -------------------------------- | ------------------------------------------------------------------------------------ |
| `hono[LinearRouter]`             | Passes (0.333 as a negative); resolved by the cull, no action                        |
| `hono[Hono or Express + Vitest]` | Bucket 4 — deferred to routing, `vitest-guide` owns it                               |
| `zod[Hono JSON request]`         | Bucket 4 — deferred to routing, `hono-guide` owns it                                 |
| `threejs[translate gizmo]`       | Bucket 4 — deferred to routing; `editor-viewport-guide` loses the scenario 0/3, open |
| `user-stories[As a / I want]`    | Bucket 1 — referent supplied, passes                                                 |
| `user-stories[SMART]`            | Bucket 1 — referent supplied, passes                                                 |
| `user-stories[three amigos]`     | Bucket 4 — deferred to routing, `bdd-guide` owns it                                  |
| `credential-management[Argon2]`  | Bucket 2 — description scoped, scores 0/3, passes                                    |

### Open findings

12 trigger failures remain, none of them a harness or invalid-run problem.

- **Claude Code's bundled skills win 3** — `skill-claude-code` ("Configure Claude
  Code permissions") and `skill-instruction` ("no claude.md — write one") lose to
  the harness's own `update-config` and `init`; `skill-workflow`'s independent
  validate loses to its review skills. The eval environment loads the bundled set
  alongside the catalog, which is faithful to real use, so the catalog cannot win
  these by the one-owner rule. Left open deliberately: the fix is a boundary
  decision about whether these skills should claim work the harness already owns.
- **5 sit at the 0.5 threshold** (`c99-opinionated` ×2, `code-quality`, `github`,
  `skill`), scoring 0.333–0.667 and flipping between sweeps — `c99-opinionated`
  scored 8/9 on one run and 6/9 on the next with a different failing subset.
  These are unstable rather than broken. Not reworded, because tuning validation
  queries against their own results is what the train/validation split exists to
  prevent.
- **`skill-accessibility` and `skill-code-quality` regressed into view** as the
  threshold moved; both are referent-free queries of the same shape as the 15
  already fixed.
- **2 negatives name an owner with no pairing** (`orthogonal-pattern`'s
  dependency-inversion query, `typescript`'s `bodyLimit` query) — same shape as
  the 6 paired in `80d2904d`, surfaced after the reruns changed which queries the
  per-skill evaluator keeps.

The routing gate failed 9 of 77 scenarios on its first run — `debugging-guide`,
`editor-viewport-guide`, `lua-opinionated-guide`, `skill-guide`, and
`typescript-guide` lose outright, and four more sit at 0.333. These are genuine
ownership disputes that no evaluator had ever measured. `ci-routing` is
`runInCI: true` but is not in the `:ci-check` chain that `ci.yml` runs, so the
main CI is unaffected; it must not be added to `skill-evals.yml` until these are
triaged.

### Validation

- `npx moon run '#skill:skill-validate'` — 176 tasks, green
- `npx moon run script-moon-skill-eval-triggers:ci-check` — green
- `npx moon run script-moon-skill-eval-outputs:ci-check` — green
- `npx moon run script-moon-skill-validate:ci-check` — green
- Final sweep: 72/72 skills, zero `invalid-run.json`, zero transient retries

## Re-baseline (2026-07-27)

The open findings were investigated and adversarially verified; the report is
`remediation-open-findings.md` in this directory. Two of its recommendations
landed here because the first was a defect introduced by this subplan.

### Further harness commits

| SHA        | Change                                                                                                                                                |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `0333a195` | an unresolvable skill name no longer counts as a competitor win; turn cap 1 -> 2; `selected_skills` per run; corrected the dependency-loading comment |
| `45594ee1` | three routing-owner pairing adds (hexagonal-pattern, data-oriented-design, hono-opinionated)                                                          |
| `6cff46fa` | the remediation report, with its unreproduced frequency claim corrected                                                                               |

`dcf1bca5` settled a run the moment any non-target name appeared. A captured
stream shows the model invoking the bare plugin name
`xonovex-skill-memory-management`, the harness answering `Unknown skill`, and
nothing launching — the model had chosen the target and mistyped it, and the run
scored the opposite with no error for a retry to catch.

### Results after the re-baseline

| Gate              | Before  | After   |
| ----------------- | ------- | ------- |
| Trigger failures  | 12      | 10      |
| Routing failures  | 9 of 77 | 3 of 80 |
| Invalid runs      | 0       | 0       |
| Transient retries | 0       | 0       |

Six of the nine routing failures were the unresolvable-name defect and cleared
without any catalog edit; `data-oriented-design-guide`'s broad-phase scenario
moved 0.333 -> 1.0. The trigger count moved only 12 -> 10 because its
composition changed: several cleared while `hono-opinionated` and `testing`
crossed the threshold in the other direction.

Run outcomes across 951 trigger runs: `target` 773, `none` 153, `competitor` 20,
`output-limit` 5.

### What `selected_skills` immediately showed

Two failures are a **declared dependency beating its own overlay**, which no
rate could have revealed and which cost probe archaeology to find before:

- `c99-opinionated-guide`'s header-rebuild query lost 2 of 3 runs to
  `c99-guide`.
- `hono-opinionated-guide`'s router query lost 2 of 3 runs to `hono-guide`.

Dropping dependency plugins is not the answer — measured separately, the
overlay alone scores 0/4 on its own physical-design query, because the base
establishes the language context that makes the overlay reachable. This is a
policy question about what a per-skill gate should measure, recorded as U7 in
the remediation report.

`instruction-guide` lost 3 of 3 to the bundled `init` skill, which settles that
finding as a genuine boundary decision rather than a description gap.

### Still open

10 trigger failures and 3 routing scenarios, all dispositioned in the
remediation report: the catalog-versus-bundled boundary (D1, D2), the
`debugging` versus `memory-management` leak-detection dispute (D3), the
degenerate `lua-opinionated` query (D5), the two dependency-competition cases
above, and the near-threshold items that stay red until run escalation (S3)
rather than being tuned against the validation split.
