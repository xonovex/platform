---
type: plan
has_subplans: false
parent_plan: ../symmetric-workflow-commands.md
parallel_group: 2
status: complete
updated: 2026-07-20
completed_date: "2026-07-20"
dependencies:
  plans:
    - subplan-01-command-contract-and-inventory.md
  files:
    - packages/command/command-workflow/commands/create.md
    - packages/command/command-workflow/commands/review.md
    - packages/command/command-workflow/commands/revise.md
    - packages/command/command-workflow/commands/execute.md
    - packages/command/command-workflow/commands/validate.md
skills_to_consult:
  - skill-guide
  - plan-guide
  - command-guide
  - testing-guide
skills_to_avoid:
  - workflow-guide
  - workflow-runtime-guide
  - ai-governance-guide
validation:
  type_check: not_applicable
  lint: not_applicable
  build: passed
  tests: passed
  integration: passed
---

# Subplan 02: Plan skill decoupling

## Objective

Reduce `skill-plan` to planning and code-research guidance. Remove its ownership
of early lifecycle, UX, solution design, decisions, approval, authority,
profiles, and provider-neutral persistence so it becomes one optional domain
capability selected by the generic operations.

This subplan changes the plan skill package only. It does not recreate any
deleted workflow/governance skill and does not rewrite the command or role
documentation owned by other children.

## Tasks

### 1. Rewrite the skill boundary and routing table

- Update `packages/skill/skill-plan/plan-guide/SKILL.md:1-105` so its name,
  description, principles, lifecycle, gotchas, and progressive-disclosure list
  cover only:
  - plan research, creation, critique, revision, subplan creation, continuation,
    update, and validation;
  - code-research operations already represented by the `code-*` and `todos`
    references.
- Remove discovery, generic research/formulation, experience design, solution
  design, decision records, accept/reject, mandatory-human gates, authority,
  automation levels, and profiles.
- Describe the plan skill as selectable by generic core commands rather than as
  a workflow owner.

The retained routing shape should be equivalent to:

```text
Planning: research | create | critique | revise | expand | continue | update | validate
Code research: barrels | comments | shared extraction | templates | TODO inventory
```

### 2. Delete references outside the planning boundary

- Delete these reference families:
  - `references/early-lifecycle-contracts.md`
  - `references/discovery-run.md`
  - `references/research-run.md`
  - `references/formulation-run.md`
  - `references/experience-design-*.md`
  - `references/solution-design-*.md`
  - `references/decision-*.md`
  - `references/plan-accept.md`
  - `references/plan-reject.md`
- Do not replace them with redirects or consolidated compatibility references.

### 3. Remove governance and persistence bleed from retained references

- Update `references/plan-research.md`, `plan-create.md`, `plan-critique.md`,
  `plan-revise.md`, `plan-subplans-create.md`, `plan-continue.md`,
  `plan-update.md`, and `plan-validate.md`.
- Remove `--profile`, authority fields, approval preconditions, governed status
  transitions, provider-neutral result wrappers, and mandatory stage ordering.
- Let callers pass an explicit plan and optional provider-native revision.
- Let subplan creation use that explicit revision regardless of an approval
  field.
- Treat status as optional descriptive/provider metadata; do not gate another
  operation with it.
- Preserve validation against explicit success criteria and evidence without
  turning successful validation into approval.

Use a neutral result contract such as:

```text
Input: inline plan or provider-native plan reference + optional revision
Action: one requested planning operation
Output: inline result or provider-native reference returned by the selected provider
```

### 4. Rebuild the evaluation corpus around the reduced skill

- Rewrite `plan-guide/eval-queries.json` so positive cases cover planning and
  code research only; remove discovery/formulation/UX/solution-design/decision,
  authority, opaque workflow resumption, and profile cases.
- Rewrite `plan-guide/evals.json` to assert the reduced boundary, explicit
  inputs, evidence-based validation, fresh-context continuation, and no
  lifecycle governance.
- Add negative boundary cases showing that UX design, PR review, release,
  incident response, and provider storage are not owned by this skill.
- Keep fixture JSON structurally valid and ensure every referenced file exists.

### 5. Align package metadata and source notes

- Update `plan-guide/SOURCES.md:1-12` to describe planning/code-research sources
  without claiming a shared lifecycle/provider/authority contract with
  `command-workflow`.
- Update descriptions in `package.json`, `.claude-plugin/plugin.json`, and
  `.codex-plugin/plugin.json` only where they describe the removed scope.
- Do not bump versions in this subplan; subplan 05 owns coordinated release
  metadata.
- Preserve the code-quality dependency only if retained code-research routing
  still uses it directly.

### 6. Validate the skill as a standalone capability

- Run formatting and skill-package validation.
- Run cross-package links after the command inventory from subplan 01 exists.
- Search the whole skill package for removed references and semantics.
- Verify retained code-research references and all links from `SKILL.md` remain
  reachable.

## Validation steps

1. `npx moon run skill-plan:ci-check`
2. `npx moon run command-workflow:cross-package-links`
3. `rg -n -- '--profile|authority-reference|mandatory-human|pending-approval|status: approved|experience-design|solution-design|discovery-run|formulation-run|decision-accept|plan-accept|plan-reject' packages/skill/skill-plan` must return no behavioral residue.
4. Validate `eval-queries.json` and `evals.json` with the existing skill-audit
   path.
5. `npx prettier --check packages/skill/skill-plan`
6. `git diff --check`

## Success criteria

- [x] `skill-plan` owns only planning and code-research behavior.
- [x] All 18 out-of-bound reference files are deleted without compatibility
      replacements.
- [x] Retained plan operations accept explicit inline/provider-native inputs and
      do not require a profile or approval state.
- [x] Status and validation cannot authorize or gate later work.
- [x] Evals positively cover the reduced boundary and negatively reject former
      lifecycle responsibilities.
- [x] Package descriptions and source notes match present behavior.
- [x] All skill formatting, audit, and link checks pass.

## Files modified/created

- Modify: `packages/skill/skill-plan/plan-guide/SKILL.md`.
- Modify: `packages/skill/skill-plan/plan-guide/SOURCES.md`.
- Modify: retained files under
  `packages/skill/skill-plan/plan-guide/references/plan-*.md`.
- Delete: the reference families listed in task 2.
- Preserve: `references/code-*.md` and `references/todos.md`, except for direct
  broken-link corrections.
- Modify: `packages/skill/skill-plan/plan-guide/eval-queries.json`.
- Modify: `packages/skill/skill-plan/plan-guide/evals.json`.
- Modify as needed: `packages/skill/skill-plan/package.json`.
- Modify as needed:
  `packages/skill/skill-plan/.claude-plugin/plugin.json`.
- Modify as needed: `packages/skill/skill-plan/.codex-plugin/plugin.json`.

## Dependencies

- Requires the generic command vocabulary established by subplan 01.
- Must finish before subplan 03 documents plan examples.
- Its package and eval outputs feed the final residue and release work in
  subplan 05.

## Estimated duration

One to two focused implementation sessions because the retained references and
evaluation corpus require semantic rewriting, not just file deletion.
