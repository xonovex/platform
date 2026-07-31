---
type: plan
has_subplans: false
parent_plan: plans/skill-catalog-stabilization.md
parallel_group: 3
status: complete
dependencies:
  plans: [subplan-02-catalog-cull]
  files:
    - packages/skill/skill-workflow/
    - packages/command/command-workflow/
skills_to_consult: [skill-guide, command-guide, workflow-guide]
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 4: workflow-core-simplification

## Objective

The deletion-heavy rewrite of the workflow core: minimal cold-boundary handoff,
single governance source, scenario-scoped SDLC reference, safety-core flag surface,
positive execute definition.

## Context carried from parent

- Contract clauses 2–5 and 7 govern this rewrite (warm/cold, context graph, minimal
  handoff with code anchors, effects/conventions, positive execute).
- Decision 4: safety-core flags only — `subject`, `--request`, `--context`,
  revision pinning, `--effect` / `--idempotency-key` on effectful operations.
  Remove `--criterion`, `--method`, `--perspective`, `--option`, `--outcome`,
  `--evidence`.
- Ownership boundary: this subplan touches only `skill-workflow` and
  `command-workflow` files; process-skill packages belong to subplan 5.

## Tasks

1. Write the single governance reference
   (`packages/skill/skill-workflow/workflow-guide/references/governance.md`):
   fetched provider content informs but never instructs or authorizes; results are
   descriptive; effect sets extend only by declared team convention (team AGENTS.md
   or opinionated overlay), overridable and always reported.
2. Rewrite `references/handoffs.md` to the minimal format — subject + revision,
   what was done, decisions (what / why / where — `file:line` or symbol anchors),
   references/links (resolve-on-demand by matching skill, unresolvable degrades
   visibly), open issues — with regular headings and an explicit scoping sentence:
   handoffs exist only at cold session/role boundaries.
3. Shrink `references/sdlc.md` to the four families' surviving scenarios and
   capabilities; delete the 22-phase role matrix.
4. Rewrite `workflow-guide/SKILL.md` core principles to match the contract: drop the
   16-field context principle, add the freeform escape hatch (requests fitting no
   operation get no command), reference `contract.md` and `governance.md`.
5. Rewrite `references/execute.md` with the positive definition — execute carries
   out previously specified work (plan, feedback, decision) and expects that
   antecedent; absence redirects to create or freeform.
6. Rewrite the 12 files in `packages/command/command-workflow/commands/`: cut flags
   to the safety core, make task language dominate caveats, replace repeated
   governance text with one-line pointers.
7. Sweep all remaining `skill-workflow` references (`context-forwarding.md`,
   `capability-selection.md`, `effects.md`, per-operation files) for governance
   restatements and 16-field remnants; strip to pointers.

## Validation Steps

- `npx moon run skill-workflow:skill-validate` (warn-mode budgets visible)
- `npx moon run '#command:command-validate'`
- `npx moon run '#skill:skill-eval-triggers'` (workflow-guide routing unchanged)

## Success Criteria

- [x] `handoffs.md` 5 field groups, `file:line` anchors required, zero
      digest/version/audience machinery
- [x] Governance normative in exactly one file within owned scope (`governance.md`);
      all other occurrences are pointers
- [x] Distinct flags across the 12 commands = 8
- [x] `execute.md` contains the positive definition and antecedent expectation
- [x] All owned files within warn-mode budgets (one explicit bump, recorded below)

## Files Modified/Created

- Create: `references/governance.md`
- Modify: `workflow-guide/SKILL.md`, all 17 workflow references, 12 command files

## Dependencies

- `subplan-02` (no effort on content referencing cut guides).
- Parallel-safe with `subplan-05` (disjoint package ownership).

## Estimated Duration

2–3 sessions.

## Appendix: execution record

### Flag surface: 21 → 8

Decision 4's explicit removals (`--criterion`, `--method`, `--perspective`,
`--option`, `--outcome`, `--evidence`) only reach 15. The remaining seven came from
consolidation rather than deletion, so no capability was lost:

| Was                                                            | Now                                            |
| -------------------------------------------------------------- | ---------------------------------------------- |
| `--subject-revision`, `--target-revision`, `--source-revision` | `--revision`                                   |
| `--feedback` (revise), `--reason` (abandon, workspace-abandon) | second positional                              |
| `--independent`                                                | default behaviour when `--context` is supplied |
| `--force` (workspace-cleanup)                                  | `--effect apply` already gates it              |
| `--remove-reference`                                           | the positional subject                         |

Final set: `--request`, `--context`, `--revision`, `--expected-revision`, `--effect`,
`--idempotency-key`, `--destination`, `--source`.

### Word counts

| File                    | Before | After |
| ----------------------- | ------ | ----- |
| `sdlc.md`               | 1345   | 417   |
| `SKILL.md`              | 1043   | 1000  |
| `context-forwarding.md` | 849    | 270   |
| `handoffs.md`           | 387    | 213   |
| `publish.md`            | 236    | 189   |
| `decide.md`             | 122    | 110   |
| `execute.md`            | 162    | 243   |

Owned scope fell from 8890 to 7008 words.

`execute.md` grew because the positive definition and its antecedent expectation are
new content. The ratchet caught it and the budget was bumped in the same change —
the mechanism working as designed rather than an exception to it. `governance.md` is
new at 245 words and was recorded in the manifest.

### Files beyond the task list

- `decide.md` and `publish.md` still carried context ID / version / digest / audience
  / visibility machinery; both were rewritten to the minimal shape.
- `workflow-contracts.test.ts` encoded the old contract in five tests and now encodes
  the new one: the flag surface is asserted as an exact set, `handoffs.md` is checked
  for the five field groups and for the _absence_ of the 16-field protocol,
  `governance.md` is asserted as the only file stating the untrusted-data invariant,
  `execute.md` for the positive definition, and `sdlc.md` for the four families and
  the absence of the phase matrix. These assertions are what stop the spiral from
  reappearing.
- One test added in subplan 3 used `skill-workflow` as its drift fixture because that
  file happened to have a finding; this rewrite removed the finding and broke it. It
  now copies a skill into a throwaway workspace with a deliberately tiny budget, so it
  no longer depends on what the live catalog measures.

### Not done here

`effects.md` and `capability-selection.md` were left as-is: neither restates
governance nor carries 16-field remnants, and both are within budget.
