---
type: plan
has_subplans: false
parent_plan: plans/agentic/platform-workflow-improvement.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - plans/agentic/platform-workflow-improvement/03-command-distillation.md
  files:
    - packages/command/command-workflow/commands/plan-list.md
    - packages/command/command-workflow/commands/plan-continue.md
    - packages/command/command-workflow/commands/plan-update.md
    - packages/command/command-workflow/commands/plan-create.md
    - packages/command/command-workflow/commands/plan-worktree-create.md
    - packages/command/command-workflow/commands/pr-create.md
    - packages/skill/skill-plan/plan-guide/**
    - packages/skill/skill-git/git-guide/references/worktree-create.md
    - packages/skill/skill-pull-request/pull-request-guide/references/create.md
skills_to_consult:
  - command-guide
  - plan-guide
  - git-guide
  - skill-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 04 — Workflow Plumbing: Discovery, Association, One Argument Name

## Objective

Close the lifecycle's navigation gaps (parent decision 6): a `plan-list`
command backed by a new plan-guide operation, automatic plan association
(`plan-worktree-create` sets the git config `plan-continue` already
reads; `pr-create` gains `--plan`), and one positional argument name —
`[plan-file]` — across the plan commands.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); read cited files before
editing. Line refs below came from research agents — verify on read.

1. `plan-continue.md` (~:29) reads `git config branch.<branch>.plan`;
   `git-guide/references/worktree-create.md` (~:60) mentions setting it
   only as a manual "next step"; `plan-worktree-create.md` doesn't
   mention it at all. The association mechanism exists but nothing sets
   it.
2. `pr-create.md` has `--work-item` but no `--plan`; a PR loses its link
   to the originating plan.
3. No discovery: plan commands auto-detect from conversation or "most
   recent `plans/*.md`". This repo's plans live at mixed depths:
   `plans/<plan>.md`, category dirs like `plans/agentic/<plan>.md`,
   subplan dirs `.../<feature>/NN-*.md`, and `plans/old/` (archive —
   always exclude it). Discovery must be recursive and keyed on
   frontmatter, not directory depth.
4. Positional-argument drift (post the 2026-07-03 rename, parent
   decision 7): `plan-continue` and `plan-update` use
   `[document-path]`; `plan-revise`/`plan-critique`/`plan-validate` use
   `[plan-file]` (already correct); `plan-decide` uses
   `[topic-or-plan-file]` (already conforms); `plan-create` uses
   `[spec-file-or-requirements]`.
5. Subplan 03 already re-touched pr-create.md — this subplan builds on
   its thin form (hence group 3).
6. Frontmatter conventions for plans (status, parallel groups) are
   established by the plan-guide skill and visible in this repo's
   `plans/agent-security-hardening.md` — plan-list parses those fields.

## Tasks

1. **plan-guide `plan-list` operation**:
   `packages/skill/skill-plan/plan-guide/references/plan-list.md` —
   read-only: glob `plans/**/*.md` recursively (always skip
   `plans/old/`), keep files whose frontmatter has `type: plan` or
   `type: roadmap`, group parent → subplans via each subplan's
   `parent_plan` field (not directory depth), parse `status`,
   `parallel_group`, `updated`; output a table with status rollup;
   flags `--status <filter>` and `--json`. Add to SKILL.md progressive
   disclosure.
2. **New command**
   `packages/command/command-workflow/commands/plan-list.md` — thin
   delegation to plan-guide **plan-list**, standard wording, `Read`,
   `Glob`, `Grep`, `Skill` in allowed-tools.
3. **Association on worktree create**: update
   `git-guide/references/worktree-create.md` to make
   `git config branch.<branch>.plan <plan-file>` a step of the operation
   (when a plan file is provided), not a "next step" note; surface the
   plan argument in `plan-worktree-create.md`'s Arguments.
4. **`pr-create --plan <plan-file>`**: add to argument-hint + Arguments;
   in pull-request-guide `create.md` (from 02), route it — include a
   plan reference line in the PR body and set
   `branch.<branch>.plan` if unset.
5. **Standardize `[plan-file]`**: rename the positional in
   `plan-continue` and `plan-update` (document-path), `plan-create`
   (spec-file-or-requirements → `[spec-or-plan-file]` — it accepts
   inline requirements too; keep semantics, align naming). Update each
   command's Arguments section and the matching plan-guide reference
   argument docs.
6. **Register `plan-list`** in command-workflow's plugin.json is NOT
   needed (commands are directory-discovered), but verify
   `plugin-validate` (from 01) and `skill-validate` stay green; run
   `#command:format`.

## Validation Steps

- `npx moon ci :ci-check` green.
- Smoke: in a scratch worktree, `plan-worktree-create` sets the branch
  config; `plan-continue` picks the plan up without prompting;
  `plan-list` renders this repo's real plans (parent + 6 subplans, old/
  excluded); `pr-create --dry-run --plan plans/agentic/platform-workflow-improvement.md`
  previews a body containing the plan reference.
- Grep: no command references `[document-path]` / `[input-file]`
  anymore.

## Success Criteria

- [ ] `plan-list` command + operation exist; output covers parents,
      subplans, statuses; `plans/old/` excluded.
- [ ] `plan-worktree-create` sets `branch.<branch>.plan`; `pr-create
      --plan` records the association and body reference.
- [ ] One positional naming scheme across plan commands, mirrored in
      plan-guide references.
- [ ] ci-check green; smoke tests pass.

## Files Modified/Created

- Created: `plan-guide/references/plan-list.md`,
  `command-workflow/commands/plan-list.md`
- Modified: `plan-guide/SKILL.md` + affected operation references,
  `git-guide/references/worktree-create.md`,
  `pull-request-guide/references/create.md`, 5 plan command files,
  `plan-worktree-create.md`, `pr-create.md`

## Dependencies

Requires 03 (same command files; 03 must settle their thin form first).

## Estimated Duration

2 days.
