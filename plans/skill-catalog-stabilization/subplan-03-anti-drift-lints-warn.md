---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 2
status: complete
dependencies:
  plans: [subplan-01-scenario-freeze-and-contract]
  files:
    - packages/script/script-moon-skill-validate/
    - packages/script/script-moon-command-validate/
    - .moon/tasks/tag-skill.yml
    - .moon/tasks/tag-command.yml
skills_to_consult:
  [typescript-guide, vitest-guide, moon-guide, fp-guide, zod-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 3: anti-drift-lints (warn mode)

## Objective

Implement the three anti-drift lints in the existing validation script packages and
wire them into moon in warn mode, so drift becomes measurable before it becomes
blocking.

## Context carried from parent

- Decision 3: ratchet (per-file ceiling in a checked-in budget manifest; growth
  requires an explicit bump in the same PR) + absolute p90 caps for new files:
  SKILL.md ≤ ~900 words, references ≤ ~650, commands ≤ ~250.
- Lints apply to `contract.md` like any other reference file.

## Tasks

1. In `packages/script/script-moon-skill-validate/src`, implement the ratchet + cap
   rule: read `budgets.json` (repo-root manifest, format
   `{ "<path>": <wordBudget> }`), fail files exceeding their budget, apply p90 caps
   to files absent from the manifest; emit warn-only when
   `XONOVEX_LINT_MODE=warn`.
2. Implement the vocabulary-ownership rule: a checked-in `vocabulary.json` maps each
   term of art to its single defining file; other files may use but not define
   (heading or bold-definition patterns) an owned term; undeclared coined terms
   (backticked novel nouns in normative sentences) warn.
3. Implement the invariant-duplication rule: flag near-duplicate normative sentences
   (n-gram similarity above threshold) appearing across ≥ 3 files.
4. Mirror the ratchet + cap rule into
   `packages/script/script-moon-command-validate/src` for command files.
5. Add a `--seed` mode that writes current file sizes into `budgets.json`
   (consumed by subplan 6).
6. Add Vitest coverage for all three rules (fixture files over/under budget,
   duplicate-definition fixtures, duplication-threshold fixtures).
7. Wire into `.moon/tasks/tag-skill.yml` and `tag-command.yml` as warn-mode task
   inputs to `skill-validate` / `command-validate`.

## Validation Steps

- `npx moon run script-moon-skill-validate:ci-check`
- `npx moon run script-moon-command-validate:ci-check`
- `npx moon run '#skill:skill-validate'` (warn output visible, non-blocking)

## Success Criteria

- [x] Three rules implemented with Vitest coverage, typecheck/lint/build green
- [x] Warn mode active on skill and command tags; enforce mode switchable by env/config
- [x] `--seed` mode produces a valid `budgets.json`

## Files Modified/Created

- Modify: `packages/script/script-moon-skill-validate/`,
  `packages/script/script-moon-command-validate/`, `.moon/tasks/tag-skill.yml`,
  `.moon/tasks/tag-command.yml`
- Create: `budgets.json`, `vocabulary.json` (initial), lint rule modules + tests

## Dependencies

- `subplan-01` (contract defines the owned vocabulary seed).
- Parallel-safe with `subplan-02` (disjoint files).

## Estimated Duration

1–2 sessions.

## Appendix: execution record

### Where each rule lives

| Rule                  | Module                                                | Runs from                       |
| --------------------- | ----------------------------------------------------- | ------------------------------- |
| Ratchet + caps        | `script-moon-common/src/drift-budgets.ts`             | both validators                 |
| Vocabulary ownership  | `script-moon-skill-validate/src/drift-vocabulary.ts`  | `skill-validate`, `drift-check` |
| Invariant duplication | `script-moon-skill-validate/src/drift-duplication.ts` | `drift-check` only              |

The ratchet went into `script-moon-common` and is imported by both validators
rather than mirrored, so the manifest format and word count have one owner;
`script-moon-command-validate` gained a `script-moon-common` dependency for it.

Duplication is a cross-catalog rule — it needs every file at once to find an
invariant restated in three of them — so it runs once from a new
`moon-skill-drift` bin (`drift-check` task) rather than per skill. Budgets and
vocabulary are per-file and stay inside `skill-validate` / `command-validate` as
the subplan intended.

### Warn mode had to bypass `--strict`

`skill-validate` runs with `--strict`, which escalates warnings to failures, so
routing drift findings through the existing warning channel would have made warn
mode blocking on the first run. Findings therefore use a separate `[DRIFT]`
channel that `--strict` ignores; `XONOVEX_LINT_MODE=enforce` moves them to
`[FAIL]`. Verified: warn exits 0 and enforce exits 1, for both the per-skill
validator and the repo-wide bin.

### Seeded baseline and first measurement

`--seed` recorded 554 files (min 6 / median 247 / max 1600 words). Because the
ratchet starts from current sizes, budgets produce zero findings today by
construction — that is the point; growth from here needs an explicit bump.

First drift report: 8 findings over 554 files.

- 5 duplication clusters, all genuine: one invariant restated across 20
  `command-utility` commands, another across all 12 `command-workflow` commands,
  plus three smaller workspace-command clusters.
- 3 vocabulary findings, of which `context-forwarding.md` defining
  `context version` is the true positive (subplan 4 removes that machinery).

### Tuning the coined-term rule

As first written it fired 18 times on API identifiers — `src`, `dst`, `libc`,
`readonly`, `satisfies`, `unknown`. A coined term of art is a phrase or
hyphenated compound; a bare backticked token is an identifier a skill may name
freely. Restricting the pattern to multi-word or hyphenated terms cut findings
from 26 to 8 with no loss of true positives. Two code-snippet false positives
remain (`git remote get-url origin`, `square extends rectangle`) — acceptable at
warn level, and exactly what warn mode exists to expose before subplan 6
enforces.

### Duplication threshold

Trigram Jaccard scored a one-word reword at 0.77 and missed it. Bigrams with a
0.75 threshold catch the reword while leaving unrelated invariants near zero, so
the separation is wide; both are asserted in `drift-duplication.test.ts`.

### Also fixed here

`skill-validate` crashed with "Could not find workspace root" when validating a
skill outside a workspace — budgets and vocabulary are repository-scoped, so the
lints now skip when no root is found.
