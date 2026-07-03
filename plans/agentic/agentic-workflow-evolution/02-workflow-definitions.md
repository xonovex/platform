---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 2
status: pending
dependencies:
  plans:
    - plans/agentic/agentic-workflow-evolution/01-step-result-contract.md
  files:
    - workflows/feature.yml
    - workflows/schemas/workflow.schema.json
    - packages/skill/skill-workflow/**
    - packages/script/script-moon-workflow-validate/**
    - .claude-plugin/marketplace.json
    - .github/CODEOWNERS
skills_to_consult:
  - skill-guide
  - plan-guide
  - moon-guide
  - typescript-guide
  - vitest-guide
  - versioning-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 02 — Workflow Definitions: The Declarative Shape + workflow-guide

## Objective

Make the workflow executable data: a definition schema (steps, gates
with policies, foreach/loop constructs, retry caps, budgets),
`workflows/feature.yml` encoding today's diagram, a new
**workflow-guide** skill owning the semantics and gate-authority
mechanics (parent decisions 3, 10, 11), and protected-path enforcement
for `workflows/**`.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); executes after the
improvement plan — re-read cited files.

1. The flow to encode is
   `packages/diagram/diagram-agent-workflow/workflow-diagram.dot`:
   research → gate(research-review) → create → gate(plan-review) →
   subplans-create → foreach parallel_group { worktree-create →
   loop(continue → validate, until ok, max N) → gate(merge-review) →
   worktree-merge } → quality-audit loop → gate(ship) →
   git-commit --push. Plan instance state (parallel groups, statuses)
   comes from plan frontmatter — the definition references those
   dynamic sets, it does not copy them. Gate-extraction rule: only 4 of
   the diagram's 6 decision diamonds become gates (research-review =
   "Good idea?", plan-review = "Review Plan", merge-review = "Merge or
   Abandon?", ship = before the final push); "More Subplans?" and
   "Another coding loop?" are NOT gates — they are the exit conditions
   of the foreach/loop constructs.
2. Gate policies: `manual | assisted | auto` with optional `reviewer`
   naming an existing operation (plan-critique, pr-review-analyze).
   Authority (decision 10): policy changes need human review
   (CODEOWNERS on `workflows/**`); interpreters read policies from the
   trusted base ref (main), never the run's branch; L3 floor comes in
   subplan 05; most-restrictive-wins.
3. Notification (decision 11): gates declare no channel — the runner
   notifies via Discord; the definition only carries the gate identity
   and policy.
4. One semantics, three interpreters (decision 2): this subplan is the
   SINGLE normative description (workflow-guide); 03 implements it as a
   prompt operation, 04 as Go. Ambiguity here becomes interpreter
   drift — write the semantics as testable rules.
5. A new skill package must follow the catalog conventions: scaffold
   like existing `packages/skill/skill-*` (plugin manifests, moon.yml
   tags, lockstep version), register in
   `.claude-plugin/marketplace.json`, pass skill-validate AND the
   improvement plan's plugin-validate.

## Tasks

1. **Definition schema** — `workflows/schemas/workflow.schema.json`:
   `name`, `steps` (command + optional args template), `flow` (sequence
   items: step | gate | foreach over `pending(parallel_group)` | loop
   with `until` result condition + `max`), `gates` (policy, reviewer,
   description), per-step `retries`, `budget` (wall-clock minutes +
   max-steps). Loop `max`, `retries`, and `budget` are REQUIRED —
   validation rejects definitions without them; no silent defaults.
   Token budgets deferred to 04's runner capabilities.
2. **`workflows/feature.yml`** — encode Context 1 with named gates:
   `research-review` (assisted), `plan-review` (manual),
   `merge-review` (assisted, reviewer plan-critique), `ship` (manual).
   Loop caps explicit (e.g. validate loop max 3).
3. **New skill `packages/skill/skill-workflow/workflow-guide`** —
   SKILL.md (description tuned: triggers on workflow definitions,
   gates, workflow-next/run; skip-clause routing to plan-guide for plan
   documents) + references: `definition-format.md` (schema semantics,
   next-step computation rules, foreach/loop evaluation),
   `gates.md` (policy meanings, reviewer wiring, authority per decision
   10, trusted-ref rule, most-restrictive-wins, audit record shape —
   gate verdicts land in the journal per 01's schema).
4. **Extend `script-moon-workflow-validate`** (from 01) to validate
   `workflows/*.yml` against the schema; add fixtures (unknown step,
   gate without policy, loop without max — each rejected).
5. **Protected paths** — `.github/CODEOWNERS` entry for `workflows/**`
   (and the schema dir) requiring owner review; document the GitLab
   equivalent as a comment (GitHub-authoritative per parent decision).
6. **Register + validate** — marketplace.json entry for the new
   package; `plugin-validate` + `skill-validate` green.

## Validation Steps

- `script-moon-workflow-validate` accepts feature.yml, rejects each
  fixture violation.
- `#skill:skill-validate` green for skill-workflow; plugin-validate
  green (registration + lockstep + manifest parity).
- Dry semantic walk: hand-execute feature.yml against this repo's real
  parent plan state (the improvement plan) — the computed next step at
  three sampled states matches expectation; record in Results.

## Success Criteria

- [ ] Schema + feature.yml validate in ci-check; loop caps, retry caps,
      and budgets are mandatory fields (parent risk: runaway loops).
- [ ] workflow-guide ships the single normative semantics incl. gate
      authority mechanics; catalog checks green.
- [ ] `workflows/**` requires human review to change (decision 10).

## Files Modified/Created

- Created: `workflows/schemas/workflow.schema.json`,
  `workflows/feature.yml`, `packages/skill/skill-workflow/**`,
  `.github/CODEOWNERS` (or entry in existing)
- Modified: `script-moon-workflow-validate`,
  `.claude-plugin/marketplace.json`

## Dependencies

Requires 01 (journal/gate-verdict shape, validator package). Blocks 03
and 04.

## Estimated Duration

3–4 days.
