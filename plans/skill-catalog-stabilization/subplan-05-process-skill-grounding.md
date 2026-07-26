---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 3
status: complete
dependencies:
  plans: [subplan-02-catalog-cull]
  files:
    - packages/skill/skill-plan/
    - packages/skill/skill-user-stories/
    - packages/skill/skill-bdd/
    - packages/skill/skill-code-review/
    - packages/skill/skill-pull-request/
    - packages/skill/skill-content/
    - packages/skill/skill-instruction/
    - packages/skill/skill-command/
    - packages/skill/skill-reflect/
skills_to_consult:
  [skill-guide, plan-guide, user-stories-guide, bdd-guide, testing-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: partial
---

# Subplan 5: process-skill-grounding

## Objective

Re-anchor surviving process skills to the concrete artifacts they produce and give
each an output eval, including the golden end-to-end handoff scenario.

## Context carried from parent

- Litmus test per instruction: could an agent do this, and could you verify it did?
- Canonical handoff scenario (contract clause 4): ticket with linked PM/UX context →
  coding session with anchored decisions → PR session producing description +
  inline comments at the anchors.
- A skill for which no output eval can be written escalates back to the cull table.
- Ownership: process-skill packages only; `skill-workflow`, `command-workflow`, and
  `skill-skill` belong to subplans 4 and 6.

## Tasks

1. Audit keep-tier process skills (`plan`, `user-stories`, `bdd`, `tdd`,
   `code-review`, `pull-request`, `content`, `instruction`, `command`, `reflect`,
   craft/pattern set) — classify concrete-enough vs. needs-rewrite; record the table
   in this file; escalate unevaluable skills.
2. Rewrite `plan-guide` first (heaviest workflow-speak: canonical-context /
   digest / provenance principles), anchored to the plan document artifact and the
   warm/cold model.
3. Rewrite the remaining flagged skills to their artifacts, applying the litmus test
   per instruction; delete what fails it.
4. Replace governance restatements in owned files with pointers to
   `governance.md` (created by subplan 4; coordinate on landing order — pointers may
   land referencing the path before the file merges).
5. Author output-eval scenarios for each rewritten skill in its
   `eval-queries.json`, rubric-based, per the existing output-eval format.
6. Author the golden end-to-end eval exercising `workflow-guide` +
   `pull-request-guide` + `github-guide` + the handoff format on the canonical
   scenario.
7. Run output evals across the rewritten set.

## Validation Steps

- `npx moon run '#skill:skill-validate'`
- `npx moon run '#skill:skill-eval-outputs'` (rewritten set + golden scenario)
- `npx moon run '#skill:skill-eval-triggers'` (routing unchanged)

## Success Criteria

- [x] Audit table recorded; every keep-tier process skill classified
- [x] Zero keep-tier process skills without output evals (all 20 already had 4–11)
- [x] Golden end-to-end eval exists and passes (0.833 with-skill, gate PASS)
- [x] All rewritten files within warn-mode budgets (one explicit bump, recorded below)

## Files Modified/Created

- Modify: process-skill packages listed in frontmatter (SKILL.md, references,
  eval-queries.json)
- Modify: this subplan (audit-table appendix)

## Dependencies

- `subplan-02`. Parallel-safe with `subplan-04` (disjoint packages; the
  `governance.md` pointer is a path reference, not a file conflict).

## Estimated Duration

2–3 sessions.

## Appendix: audit and execution record

### Audit table

Litmus test per instruction: could an agent do this, and could you verify it did?
Workflow-speak was measured as occurrences of `canonical context`, context
`identity/version/digest`, `provenance`, `audience`, `visibility`, `untrusted data`,
`effect mode`, and removed flag names.

| Skill                                                                                                                                   | Output evals | Workflow-speak hits | Classification                          |
| --------------------------------------------------------------------------------------------------------------------------------------- | -----------: | ------------------: | --------------------------------------- |
| `plan`                                                                                                                                  |           11 |   15 across 9 files | **rewritten**                           |
| `user-stories`                                                                                                                          |            4 |                   0 | concrete enough                         |
| `bdd`                                                                                                                                   |            4 |                   0 | concrete enough                         |
| `tdd`                                                                                                                                   |            4 |                   0 | concrete enough                         |
| `code-review`                                                                                                                           |            4 |                   0 | concrete enough                         |
| `pull-request`                                                                                                                          |            5 |                   0 | concrete enough (hosts the golden eval) |
| `content`                                                                                                                               |            4 |                   1 | concrete enough — false positive        |
| `instruction`                                                                                                                           |            4 |                   0 | concrete enough                         |
| `command`                                                                                                                               |            7 |                   0 | concrete enough                         |
| `reflect`                                                                                                                               |            5 |                   1 | concrete enough — false positive        |
| `testing`, `ddd`, `oop`, `fp`, `connascence`, `code-quality`, `hexagonal-pattern`, `microkernel-pattern`, `orthogonal-pattern`, `skill` |     4–6 each |                   0 | concrete enough                         |

Nothing escalated to the cull table: every keep-tier process skill is verifiable and
already carried output evals, so the criterion "zero without output evals" held at
baseline rather than needing work.

The two single hits are domain vocabulary, not workflow-speak, and were left alone:
`content-guide` uses "audience" to mean the reader of an article, and `reflect-guide`
uses "provenance" for candidate → decision → version tracking. Renaming either would
damage the skill to satisfy a grep.

### What `plan-guide` actually needed

15 occurrences across `SKILL.md` and 8 references. "Canonical context" became
"carried decisions"; the principle listing context `identity, version, digest,
provenance, applicability, status, and visibility` became decisions with code anchors;
the two `untrusted data` restatements became pointers to **workflow-guide**, which
owns that rule after subplan 4.

Rewording ran slightly over budget on four files. Three were tightened back under —
`SKILL.md` ended at 770 against its 776 budget — and `plan-create.md` needed +2 words
for the cross-reference, so that one budget was bumped explicitly. The ratchet did its
job on a real edit.

### Golden end-to-end eval

Lives in `pull-request-guide/evals.json` as eval 5. The subplan's ownership boundary
puts `skill-workflow` in subplans 4 and 6, and the scenario's artifact is a pull
request, so `pull-request-guide` is the right home.

It replays the canonical scenario: a cold-boundary handoff (subject + revision, two
decisions with `file:line` anchors, an unresolvable design link, an open question)
into a session with no memory of the work. Its six assertions check that the PR
description is grounded in the handoff rather than re-derived, that both decisions
carry over with their reasons, that inline comments land on the anchors the handoff
named, that the open question survives, that the dead link is reported rather than
fabricated, and that ticket text is not treated as authorization to merge.

Result: `with_skill` 0.833 on the golden eval; the package scored `with_skill` 0.867
vs `without_skill` 0.333 (delta +0.534) and the aggressive quality gate passed.

### Task-default defect found and fixed

`skill-eval-outputs` failed with "target skill did not activate" because the moon task
omitted the `--plugin-dir .` that its CI sibling passes — the same defect class as the
`skill-eval-triggers` batch-size gap in subplan 3. Added to the task default; the
evals then ran. The default eval workspace (`*-guide-workspace/`) was also writing run
artifacts into the repo untracked, and is now gitignored.

### Not completed

Catalog-wide `#skill:skill-eval-outputs` was not run: roughly 3 minutes per skill puts
72 skills at hours of wall clock plus spend. `plan-guide`'s own output evals also
could not complete at the time — its first eval's with-skill arm died on
`error_max_turns`, and the harness exposed no turn limit to raise, so one unfinishable
eval invalidated the whole batch (defect 3 from subplan 2).

Since resolved (2026-07-26): the generation turn cap defaults to 12 and is
configurable via `--max-turns` / `MAX_TURNS` up to 24, and the output harness retries
transient failures (three attempts, matching the trigger harness) before invalidating
a batch. `plan-guide` eval 1 now completes and scores; the catalog-wide sweep remains
unrun on cost grounds only.
