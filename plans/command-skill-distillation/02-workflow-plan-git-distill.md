---
type: plan
has_subplans: false
parent_plan: plans/command-skill-distillation.md
parallel_group: 2
status: complete
feature: command-skill-distillation
dependencies:
  plans:
    - plans/command-skill-distillation/00-mechanism-pilot.md
  files:
    - packages/command/command-workflow/commands/plan-create.md
    - packages/command/command-workflow/commands/plan-clarify.md
    - packages/command/command-workflow/commands/plan-refine.md
    - packages/command/command-workflow/commands/plan-continue.md
    - packages/command/command-workflow/commands/plan-update.md
    - packages/command/command-workflow/commands/plan-validate.md
    - packages/command/command-workflow/commands/plan-subplans-create.md
    - packages/command/command-workflow/commands/plan-tdd-create.md
    - packages/command/command-workflow/commands/plan-research.md
    - packages/command/command-workflow/commands/plan-research-code-align.md
    - packages/command/command-workflow/commands/plan-research-code-harden.md
    - packages/command/command-workflow/commands/plan-research-code-simplify.md
    - packages/command/command-workflow/commands/plan-worktree-create.md
    - packages/command/command-workflow/commands/plan-worktree-merge.md
    - packages/command/command-workflow/commands/plan-worktree-abandon.md
    - packages/command/command-workflow/commands/git-commit.md
    - packages/command/command-workflow/.claude-plugin/plugin.json
    - packages/command/command-workflow/.codex-plugin/plugin.json
skills_to_consult: [command-guide, git-guide, pull-request-guide, code-review-guide]
validation:
  type_check: n/a
  lint: pass
  build: pass
  tests: n/a
  integration: documented
---

# 02 — Workflow plan-* and git Distillation

## Objective

Distill the 16 tier-1 `command-workflow` commands (the 12 `plan-*` commands, the 3
`plan-worktree-*` commands, and `git-commit` — NOT the 5 `pr-*` commands) to the thin
delegator shape proven by the mechanism pilot (subplan 00). Each command collapses to
frontmatter + an `## Arguments` flag contract + a `## Delegation` block that loads its
guideline skill and names the operation; everything else (Goal / Workflow / Output /
Examples / Error Handling / Gotchas) is deleted because it duplicates the skill reference.
Then declare the skill presence as a hard install-time requirement by adding
`dependencies: ["xonovex-skill-plan", "xonovex-skill-git"]` to BOTH
`command-workflow/.claude-plugin/plugin.json` and `.codex-plugin/plugin.json`.

The operation names map 1:1 to existing reference files (verified present): `plan-*`
commands map to `xonovex-skill-plan` / `plan-guide` references of the same name; the three
worktree commands and `git-commit` map to `xonovex-skill-git` / `git-guide` references
`worktree-create`, `worktree-merge`, `worktree-abandon`, and `commit` respectively.

## Tasks

1. **Distill the 8 `plan-*` lifecycle commands** (all delegate to plugin
   `xonovex-skill-plan`, skill `plan-guide`).
   Files and their operation (the reference file is the same-named one under
   `packages/skill/skill-plan/plan-guide/references/`):
   - `packages/command/command-workflow/commands/plan-create.md` → operation `plan-create`
   - `packages/command/command-workflow/commands/plan-clarify.md` → operation `plan-clarify`
   - `packages/command/command-workflow/commands/plan-refine.md` → operation `plan-refine`
   - `packages/command/command-workflow/commands/plan-continue.md` → operation `plan-continue`
   - `packages/command/command-workflow/commands/plan-update.md` → operation `plan-update`
   - `packages/command/command-workflow/commands/plan-validate.md` → operation `plan-validate`
   - `packages/command/command-workflow/commands/plan-subplans-create.md` → operation `plan-subplans-create`
   - `packages/command/command-workflow/commands/plan-tdd-create.md` → operation `plan-tdd-create`

   For each: keep the existing `description` and `argument-hint` verbatim; add `Skill` to
   `allowed-tools`; keep the `## Arguments` flag/default contract verbatim (this is the
   only substantive command-unique content — e.g. `plan-create` keeps its
   `spec-file-or-requirements` / `--interactive` / `--depends-on <plan>` / `--dry-run`
   list); delete the `## Prerequisites`, `## Goal`, `## Core Workflow`,
   `## Implementation Details`, `## Output`, `## Examples`, `## Error Handling`, and
   `## Gotchas` sections. Replace the body below `## Arguments` with the standard
   `## Delegation` block. Target shape (illustrated for `plan-create`):

   ```markdown
   ---
   description: >-
     Create a high-level plan with research for user review before detailed
     subplans
   allowed-tools:
     - Write
     - Read
     - Glob
     - Grep
     - TaskCreate
     - TaskUpdate
     - AskUserQuestion
     - Skill
   argument-hint: "[spec-file-or-requirements] [--interactive] [--depends-on <plan>] [--dry-run]"
   ---

   # /xonovex-workflow:plan-create — Create Plan with Research

   ## Arguments

   - `spec-file-or-requirements` (optional): Path to spec or inline requirements (defaults to conversation context)
   - `--interactive` (optional): Ask context-dependent technical questions during research
   - `--depends-on <plan>` (optional): Mark dependency on another plan
   - `--dry-run` (optional): Preview without writing files

   ## Delegation

   Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
   **plan-create** operation with these arguments. The skill is the source of truth for
   the procedure, output format, and gotchas — do not restate them.
   ```

2. **Distill the 4 `plan-research*` commands** (all delegate to plugin
   `xonovex-skill-plan`, skill `plan-guide`).
   - `packages/command/command-workflow/commands/plan-research.md` → operation `plan-research`
   - `packages/command/command-workflow/commands/plan-research-code-align.md` → operation `plan-research-code-align`
   - `packages/command/command-workflow/commands/plan-research-code-harden.md` → operation `plan-research-code-harden`
   - `packages/command/command-workflow/commands/plan-research-code-simplify.md` → operation `plan-research-code-simplify`

   Same transformation as task 1: preserve `description`/`argument-hint`, add `Skill` to
   `allowed-tools`, keep the `## Arguments` block verbatim, delete all procedural sections,
   and end with the `## Delegation` block naming the matching `plan-guide` operation. These
   are read-only research commands — the delegation wording is identical; only the operation
   name differs.

3. **Distill the 3 worktree commands** (all delegate to plugin `xonovex-skill-git`, skill
   `git-guide`). Note the command→operation rename (command drops the `plan-` prefix):
   - `packages/command/command-workflow/commands/plan-worktree-create.md` → operation `worktree-create`
   - `packages/command/command-workflow/commands/plan-worktree-merge.md` → operation `worktree-merge`
   - `packages/command/command-workflow/commands/plan-worktree-abandon.md` → operation `worktree-abandon`

   Same transformation: keep `description`/`argument-hint`, add `Skill` to `allowed-tools`,
   keep `## Arguments` verbatim (e.g. `plan-worktree-create` keeps `feature-name` and
   `--from <branch>`), delete `## Goal`, `## Core Workflow`, `## Naming Convention`,
   `## Implementation Steps`, `## Output`, `## Error Handling`, `## Gotchas`, and end with:

   ```markdown
   ## Delegation

   Load the `git-guide` skill (plugin `xonovex-skill-git`) and perform its
   **worktree-create** operation with these arguments. The skill is the source of truth for
   the procedure, output format, and gotchas — do not restate them.
   ```

4. **Distill `git-commit`** (delegates to plugin `xonovex-skill-git`, skill `git-guide`,
   operation `commit`).
   - `packages/command/command-workflow/commands/git-commit.md` → operation `commit`

   Keep the existing `description` and the long multi-flag `argument-hint` verbatim; add
   `Skill` to `allowed-tools` (currently `Bash`, `Read`); keep the `## Arguments` block with
   all 8 flags verbatim (`message`, `--type`, `--path`, `--remote`, `--branch`, `--push`,
   `--dry-run`, `--interactive`); delete `## Goal`, `## Core Workflow`,
   `## Smart Suggestion Logic`, `## Implementation Steps`, `## Commit Format`, `## Output`,
   `## Error Handling`, `## Gotchas`, `## Examples`; end with the `## Delegation` block
   naming the **commit** operation.

5. **Add the dependencies to BOTH manifests.** Add the same bare-string `dependencies`
   array to each file (both currently have no `dependencies` key — insert after `author`):
   - `packages/command/command-workflow/.claude-plugin/plugin.json`
   - `packages/command/command-workflow/.codex-plugin/plugin.json`

   ```json
   "dependencies": ["xonovex-skill-plan", "xonovex-skill-git"]
   ```

   Both skills live in the same `xonovex-marketplace`, so bare names suffice and `version`
   is optional. IMPORTANT: subplan 04 (pr-command-distill) ALSO edits these two manifests to
   append `xonovex-skill-pull-request`, `xonovex-skill-code-review`, and
   `xonovex-skill-git-host` to this same array — 04 must land AFTER 02 and extend (not
   replace) the array. Keep this array on its own line / stable formatting so 04's diff is a
   clean append.

6. **Validate.** Run prettier `fmt-check`, `moon build` for `command-workflow`, JSON-parse
   both manifests, and integration-invoke at least one command per family (lifecycle,
   research, worktree, git-commit), confirming the named skill loads via the `Skill` tool and
   output matches the pre-distillation behavior. See Validation Steps.

## Validation Steps

- **type_check**: N/A (markdown + JSON only; no app code). Mark satisfied once JSON parses.
- **lint (fmt)**: `npx moon run command-workflow:fmt-check` — prettier format check on the
  touched package passes for all 16 rewritten command files and both manifests.
- **build**: `npx moon run command-workflow:build` — the `command-workflow` project builds
  green. Also confirm both manifests are valid JSON:
  `python3 -m json.tool packages/command/command-workflow/.claude-plugin/plugin.json` and the
  `.codex-plugin` counterpart.
- **tests**: N/A (no unit tests for command markdown).
- **integration**: With the pilot mechanism proven (subplan 00), invoke one command per
  family and confirm the delegated skill loads at run time and output is unchanged:
  - lifecycle: `/xonovex-workflow:plan-validate` loads `plan-guide` (plugin
    `xonovex-skill-plan`).
  - research: `/xonovex-workflow:plan-research` loads `plan-guide`.
  - worktree: `/xonovex-workflow:plan-worktree-create` loads `git-guide` (plugin
    `xonovex-skill-git`) and performs `worktree-create`.
  - commit: `/xonovex-workflow:git-commit --dry-run` loads `git-guide` and performs `commit`.
  Verify each command, read in isolation, names the exact skill + operation and that no
  command body restates Workflow/Output/Gotchas.

## Success Criteria

- [ ] All 16 tier-1 `command-workflow` commands are thin delegators (~15–25 lines): frontmatter + `## Arguments` + `## Delegation`, with no restated Goal/Workflow/Output/Examples/Error Handling/Gotchas sections.
- [ ] Every rewritten command has `Skill` in `allowed-tools` and an unchanged `description` and `argument-hint`.
- [ ] Each `## Delegation` block names the correct plugin + skill + operation: the 12 `plan-*` → `xonovex-skill-plan`/`plan-guide`/same-named op; the 3 worktree → `xonovex-skill-git`/`git-guide`/`worktree-{create,merge,abandon}`; `git-commit` → `xonovex-skill-git`/`git-guide`/`commit`.
- [ ] Both `command-workflow` manifests declare `dependencies: ["xonovex-skill-plan", "xonovex-skill-git"]` and remain valid JSON.
- [ ] `command-workflow:fmt-check` and `command-workflow:build` pass.
- [ ] One command per family, when invoked, loads its skill at run time and produces output unchanged from before distillation.
- [ ] The dependencies array is formatted so subplan 04 can cleanly append three more skills.

## Files Modified / Created

- `packages/command/command-workflow/commands/plan-create.md` (modified)
- `packages/command/command-workflow/commands/plan-clarify.md` (modified)
- `packages/command/command-workflow/commands/plan-refine.md` (modified)
- `packages/command/command-workflow/commands/plan-continue.md` (modified)
- `packages/command/command-workflow/commands/plan-update.md` (modified)
- `packages/command/command-workflow/commands/plan-validate.md` (modified)
- `packages/command/command-workflow/commands/plan-subplans-create.md` (modified)
- `packages/command/command-workflow/commands/plan-tdd-create.md` (modified)
- `packages/command/command-workflow/commands/plan-research.md` (modified)
- `packages/command/command-workflow/commands/plan-research-code-align.md` (modified)
- `packages/command/command-workflow/commands/plan-research-code-harden.md` (modified)
- `packages/command/command-workflow/commands/plan-research-code-simplify.md` (modified)
- `packages/command/command-workflow/commands/plan-worktree-create.md` (modified)
- `packages/command/command-workflow/commands/plan-worktree-merge.md` (modified)
- `packages/command/command-workflow/commands/plan-worktree-abandon.md` (modified)
- `packages/command/command-workflow/commands/git-commit.md` (modified)
- `packages/command/command-workflow/.claude-plugin/plugin.json` (modified — add `dependencies`)
- `packages/command/command-workflow/.codex-plugin/plugin.json` (modified — add `dependencies`)

## Dependencies

- **Must land after subplan 00 (mechanism-pilot).** 00 proves the runtime contract this
  subplan relies on: that `plugin.json` `dependencies` auto-installs the skill, that a thin
  command loads its skill via the `Skill` tool at run time, that Codex `.codex-plugin`
  honours both, and that the `--plugin-dir` dev loop resolves dependencies. Do not start
  until 00 passes.
- **Independent of subplan 01 (utility-distill).** 01 edits only `command-utility`; this
  subplan edits only `command-workflow`. They share no files and can run in parallel
  worktrees.
- **Subplan 04 (pr-command-distill) must land AFTER this subplan.** Both 02 and 04 edit
  `command-workflow/.claude-plugin/plugin.json` and `.codex-plugin/plugin.json`. 02
  introduces the `dependencies` array (`xonovex-skill-plan`, `xonovex-skill-git`); 04
  appends `xonovex-skill-pull-request`, `xonovex-skill-code-review`, and the new
  `xonovex-skill-git-host`. Serializing 02→04 avoids a manifest merge conflict on that array.

## Estimated Duration

~half a day — 16 near-mechanical command rewrites (delete restated sections, keep
frontmatter + Arguments, append the Delegation block) plus a two-line edit to each of the
two manifests, then the fmt/build/integration validation pass.
