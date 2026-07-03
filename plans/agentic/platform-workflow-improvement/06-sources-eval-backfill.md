---
type: plan
has_subplans: false
parent_plan: plans/agentic/platform-workflow-improvement.md
parallel_group: 2
status: pending
dependencies:
  plans:
    - plans/agentic/platform-workflow-improvement/01-quality-gates.md
  files:
    - packages/skill/*/*/SOURCES.md
    - packages/skill/*/*/eval-queries.json
skills_to_consult:
  - skill-guide
  - content-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 06 — Sources & Eval Backfill: Make the Audit Mean Something

## Objective

Bring the catalog to "SOURCES.md present or explicitly exempt" for all 74
skills (parent decision 4) using the exemption marker from subplan 01,
without inventing provenance; add `eval-queries.json` to the priority
set of process and easily-confused skills.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03).

1. 29 skills lack SOURCES.md (audit 2026-07-03): adr, fable, fp, astro,
   c99-game-opinionated, cmake, content, docker, git, plan,
   presentation, python, react, remotion, shell-scripting,
   sql-postgresql, kubernetes, lua-opinionated, lua, moon, motion-react,
   strudel, terraform, threejs, typescript-to-lua, typescript, vitest,
   zod, expressjs, reflect. Re-derive the list at execution time
   (`for d in packages/skill/*/*/; do [ -f "$d/SOURCES.md" ] || echo
   "$d"; done`) — 05 may have changed it.
2. Two classes: source-derived (typescript, react, vitest, zod,
   kubernetes, terraform, python, lua, docker, sql-postgresql, astro,
   remotion, threejs, motion-react, strudel, moon, expressjs, cmake,
   shell-scripting, git, typescript-to-lua, content, presentation —
   judge each) vs house-style/principles (fp, fable, adr,
   c99-game-opinionated, lua-opinionated, plan, reflect — candidates
   for `exempt`).
3. Subplan 01 defined the marker: SOURCES.md with frontmatter
   `exempt: "<reason>"`; the audit distinguishes tracked / exempt /
   MISSING. Parent rule: sources that cannot be confidently
   reconstructed get a low-confidence marker or exempt-with-reason —
   never invented provenance. Timebox: ~30 min per skill.
4. SOURCES.md format: follow existing examples (45 skills have one —
   e.g. `skill-code-review/code-review-guide/SOURCES.md`); the
   audit-sources README defines the required fields.
5. eval-queries.json: 42 exist (e.g. react, code-review) as format
   models; 32 missing. Priority set (parent decision 4): plan, git,
   github, gitlab, testing, tdd, bdd, command, instruction, reflect.
   `moon-skill-eval-triggers` runs them (claude CLI, `--max-budget-usd`).
6. Trigger-confusable pairs the queries must discriminate: github vs
   gitlab vs git vs pull-request; testing vs tdd vs bdd vs vitest;
   command vs skill vs instruction.

## Tasks

1. **Classify** the missing-SOURCES list (re-derived) into
   source-derived vs exempt candidates; record the classification table
   in this subplan's Results section when executing.
2. **Author SOURCES.md** for the source-derived set: identify the actual
   upstream docs each guide distills (official docs, release notes
   already cited in reference files); mark entries reconstructed with
   low confidence as such. Timeboxed — escalate to exempt-with-reason
   rather than fabricate.
3. **Write exempt SOURCES.md** for the house-style set with a one-line
   reason ("house style - no external upstream").
4. **Author eval-queries.json** for the 10 priority skills, modeled on
   the existing react/code-review files: include train + validation
   queries, positive triggers AND confusable negatives (Context 6
   pairs).
5. **Run the audit**: `npx moon run '#skill:skill-audit-sources'` —
   report shows zero MISSING; spot-run
   `moon-skill-eval-triggers` for 2–3 priority skills locally with
   `--max-budget-usd 2` to confirm query files work.
6. **Re-run the weekly workflow manually** (workflow_dispatch on 01's
   skill-audit.yml) and confirm the issue report reflects the new state.

## Validation Steps

- Audit report: 74/74 tracked or exempt, zero MISSING.
- Eval spot-runs complete within budget and produce trigger verdicts.
- `npx moon ci :ci-check` green (skill-validate over touched packages).

## Success Criteria

- [ ] Every skill has SOURCES.md — tracked (with confidence markers
      where reconstructed) or exempt-with-reason; none invented.
- [ ] 10 priority skills have eval-queries.json with confusable
      negatives.
- [ ] Manual weekly-audit run reports clean; eval spot-runs pass within
      budget.

## Files Modified/Created

- Created: ~29 SOURCES.md files (tracked or exempt), 10
  eval-queries.json files.
- Modified: none beyond those.

## Dependencies

Requires 01 (exemption marker + audit three-state reporting + weekly
workflow). Runs parallel with 03 (disjoint files).

## Estimated Duration

2–3 days (SOURCES research dominates).
