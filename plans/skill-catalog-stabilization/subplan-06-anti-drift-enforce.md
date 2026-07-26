---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 4
status: complete
dependencies:
  plans:
    - subplan-03-anti-drift-lints-warn
    - subplan-04-workflow-core-simplification
    - subplan-05-process-skill-grounding
  files:
    - budgets.json
    - vocabulary.json
    - packages/command/command-utility/commands/
    - packages/skill/skill-skill/
skills_to_consult: [typescript-guide, moon-guide, skill-guide, command-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 6: anti-drift-enforce

## Objective

Seed the budgets from the corrected catalog, flip the lints from warn to enforce,
and encode deletion-first editing into the meta-skills that perform editing passes.

## Tasks

1. Run the seeding mode from subplan 3 to write post-rewrite file sizes into
   `budgets.json`; commit the manifest.
2. Populate `vocabulary.json` from the contract and governance references (each term
   of art → its single defining file).
3. Flip lint mode to enforce in `.moon/tasks/tag-skill.yml` and `tag-command.yml`;
   verify full catalog passes.
4. Update the editing meta-commands in
   `packages/command/command-utility/commands/` (`skill-optimize`,
   `skill-simplify`, `instructions-consolidate`, and peers): deletion-first rule (a
   pass must cut at least as much as it adds), no-new-nouns rule, budget awareness.
5. Update `packages/skill/skill-skill/skill-guide` authoring rules with the same
   rules plus the artifact-anchoring litmus test.
6. Full-catalog verification run.

## Validation Steps

- `npx moon run '#skill:ci-check'` and `npx moon run '#command:ci-check'` (enforce
  mode)
- Negative test: a deliberately oversized fixture fails; removing it passes.

## Success Criteria

- [x] `budgets.json` (555 entries) and `vocabulary.json` (16 terms) seeded
- [x] Lints enforcing; full ci-check green on both tags; negative test fails and
      passes as expected
- [x] `skill-guide` carries deletion-first, no-new-nouns, and artifact-anchoring;
      `instruction-guide` and `command-guide` cross-reference it (see deviation below)

## Files Modified/Created

- Create/modify: `budgets.json`, `vocabulary.json`
- Modify: `.moon/tasks/tag-skill.yml`, `.moon/tasks/tag-command.yml`,
  `command-utility` editing commands, `skill-skill/skill-guide`

## Dependencies

- Requires final file sizes from subplans 4 and 5; lint machinery from subplan 3.

## Estimated Duration

1 session.

## Appendix: execution record

### Two rules had to be corrected before enforce was honest

Flipping the switch on the warn-mode findings would have failed CI on 8 findings, so
each was diagnosed rather than suppressed.

**Duplication fired on command argument contracts.** Every finding was in command
documents — shared `--request` / `--context` / `--effect` wording across the 12
workflow commands and the 20 `command-utility` commands. That wording is not drift:
`command-validate` requires every flag to appear in both the `argument-hint` and the
Arguments section of each file, so the two gates directly contradicted each other.
Restricted to skill prose, the rule reports 0 — which is the correct answer after
subplan 4 removed the restated invariants. Scoped at the call site so the function
stays general, with the reason recorded next to it.

**Coined-term detection fired only on examples.** `git remote get-url origin` and
`square extends rectangle` are a shell command and a code snippet, each shown once. A
term of art is _used_ after it is defined, so the rule now requires the phrase to
appear at least twice in the file. That is a principled discriminator rather than a
blocklist, and it removed both false positives while keeping the real case.

Both fixes carry tests, including the two real false positives as regression cases.

### Enforce mode blocked its own change

Adding the three authoring rules grew `skill-guide` by 104 words,
`instruction-guide` by 30, and `command-guide` by 30, and enforce mode failed all
three. That is the deletion-first rule applying to the change that introduced it. The
additions are net-new normative content rather than restatements, so the declared
escape valve applied and the three budgets were bumped explicitly in the same change.

Negative test: padding `fp-guide/SKILL.md` to 321 words against its 121-word budget
exits 1 with `[FAIL] budget:`; restoring the file exits 0.

### Deviation: where the editing rules live

Task 4 asked for the rules in the `command-utility` editing commands. All seven are
thin delegators whose own text says "the skill is the source of truth for the
procedure — do not restate them", and adding procedure to them would both contradict
that contract and manufacture the cross-file duplication the lint exists to catch.

The rules therefore went to the owners the commands delegate to: `skill-guide` states
them normatively, and `instruction-guide` and `command-guide` carry a one-line
cross-reference so their simplify / consolidate / distill passes inherit the rule
without copying it. Task 4's intent — editing passes carry the rules — holds; the
placement follows the catalog's own one-owner discipline.

### Seeded state

`budgets.json` reseeded from the corrected catalog: 555 files, 154,350 words, median
245, max 1600. `vocabulary.json` holds 16 terms — 12 owned by `contract.md`, 4 by
`governance.md`.

### Still open for subplan 7

The eval-harness fail-fast retry (defect 3) and the routing-owner invariant living
only in `script-moon-skill-eval-triggers:ts-coverage` rather than the skill-tag gate.
Neither blocks enforcement of the three drift rules.
