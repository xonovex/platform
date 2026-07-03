---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - workflows/schemas/step-result.schema.json
    - packages/skill/skill-plan/plan-guide/**
    - packages/skill/skill-pull-request/pull-request-guide/references/create.md
    - packages/skill/skill-git/git-guide/references/commit.md
    - packages/script/script-moon-workflow-validate/**
    - .moon/tasks/**
skills_to_consult:
  - plan-guide
  - skill-guide
  - typescript-guide
  - vitest-guide
  - moon-guide
  - zod-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 01 — Step Result Contract: Machine-Readable Outcomes

## Objective

Define and document the result every workflow step emits — schema,
journal convention, failure synthesis, workflow-context gating — and add
CI validation for journal files. After this subplan, any orchestrator
(human, L1 command, L2 loop, L3 controller) can read what a step did
without parsing session prose. (Parent design decisions 1, 4; critique
resolutions: run-id filenames, context gating, synthesized failures.)

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03). This plan executes AFTER
`plans/agentic/platform-workflow-improvement.md` subplans 01–04 — plan-guide
references and command files WILL have changed (e.g. `[plan-file]`
renames from its subplan 04). Re-read every cited file; treat named
constructs, not line numbers, as the anchor.

1. Steps are thin commands delegating to skill operations; the emission
   rule therefore belongs in the SKILL operation references (plan-guide
   et al), NOT in each command file — commands stay thin.
2. Journal convention (parent decision 4 + critique fixes):
   `plans/<plan>/runs/<run-id>-<step>.yaml`. Run-id format:
   `<branch-slug>-<seq>` — branch-slug is the current git branch with
   `/` replaced by `-`; seq is a zero-padded 3-digit ordinal, one
   higher than the highest existing seq for that branch-slug in the
   runs/ dir. Parallel worktrees are always on different branches, so
   per-branch ordinals cannot collide at merge; a retried step gets
   the next seq. Append-only; plan frontmatter updates remain
   serialized via plan-update/merge steps.
3. Workflow-context gating: results are written ONLY when
   `git config branch.<branch>.plan` is set (established by
   plan-worktree-create after the improvement plan) or an explicit flag
   is passed. Casual manual runs write nothing.
4. Synthesized failure is a RUNNER responsibility (implemented in 04/05),
   but its shape is part of this schema: a `failed` result with
   `synthesized: true`, the session exit reason in `summary`, and the
   session log path (when one exists) in `evidence`.
5. `plans/` is not a moon project; the validation task attaches to the
   new script package with repo-root inputs (same pattern as other
   `script-moon-*` packages; see `.moon/tasks/tag-*.yml` inheritance).

## Tasks

1. **Schema** — `workflows/schemas/step-result.schema.json`: fields
   `step`, `plan`, `run_id`, `status` (`ok | blocked | failed |
   needs-review`), `summary`, `evidence[]` (paths), `artifacts[]`
   (paths), `next` (advisory step name, optional), `gate` (verdict +
   reviewer, optional), `synthesized` (bool, default false), `started` /
   `finished` (ISO timestamps). Versioned via a `schema` field.
2. **plan-guide reference `step-results.md`** — the contract: journal
   path + run-id rule, context gating, append-only discipline, the
   idempotency requirement for steps (safe re-run after a failed
   result), the synthesized-failure shape, and the sweep rule
   (plan-update commits pending journal files). Add to SKILL.md
   progressive disclosure.
3. **Emission sections in operation references** — add a short
   "Result" section to each step operation's reference: plan-guide
   (research, create, continue, validate, update, subplans-create,
   critique, refine, clarify, interrogate, plan-list), pull-request-guide
   (create), git-guide (commit, worktree ops). Each states the status
   mapping for that operation (e.g. plan-validate: criteria met → `ok`,
   unmet → `needs-review` with evidence).
4. **plan-update sweep** — extend plan-guide's plan-update reference:
   collect uncommitted `runs/*.yaml`, include them in the status commit.
5. **New package `packages/script/script-moon-workflow-validate`** —
   validates `plans/*/runs/*.yaml` against the schema (and, extended by
   subplan 02, `workflows/*.yml`); scaffold mirrors
   `script-moon-plugin-validate` from the improvement plan. Unit tests
   with valid/invalid fixtures (bad status, missing run_id, unknown
   fields).
6. **Wire into CI** — task on the script package with inputs
   `/plans/**/runs/*.yaml` + `/workflows/schemas/**`, appended to the
   repo's ci-check path so a malformed journal file fails the PR.

## Validation Steps

- `npx moon run script-moon-workflow-validate:test` green (fixtures).
- Hand-author one valid + one invalid journal file in a scratch plan
  dir: ci-check accepts / rejects respectively.
- `npx moon run '#skill:skill-validate'` green for skill-plan,
  skill-pull-request, skill-git.
- Manual L0 check: run a plan command WITHOUT a workflow context — no
  journal file appears; set the branch config — result file appears and
  validates.

## Success Criteria

- [ ] Schema exists, versioned; validator enforces it in ci-check.
- [ ] step-results.md documents path/run-id/gating/idempotency/
      synthesized-failure/sweep; every step operation reference has a
      Result section.
- [ ] L0 behavior unchanged without workflow context; with context,
      results appear and validate (parent success criterion 1).

## Files Modified/Created

- Created: `workflows/schemas/step-result.schema.json`,
  `plan-guide/references/step-results.md`,
  `packages/script/script-moon-workflow-validate/**`
- Modified: plan-guide SKILL.md + ~11 operation references,
  pull-request-guide create.md, git-guide commit/worktree references,
  moon task wiring

## Dependencies

External: improvement plan groups 1–3 merged (branch-plan association,
`[plan-file]` naming, script package conventions). Within this plan:
none (group 1); 02 extends the validator.

## Estimated Duration

3–4 days.
