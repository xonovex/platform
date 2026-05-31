---
description: Resume work from an existing plan document with full context loading
argument-hint: "[document-path]"
---

# /xonovex-workflow:plan-continue – Continue Progress from Plan

Resumes work from an existing plan document by loading its context and outlining the next steps. Auto-detects associated plan from git config if available.

## Goal

**Start working IMMEDIATELY. No questions, no asking for permission, no confirmations.**

**ONE plan only. When complete, STOP. User re-runs `/xonovex-workflow:plan-continue` for next plan.**

- Auto-detect plan from git config or use provided path
- Find first actionable child plan (if parent has subplans)
- Execute that single plan to completion
- Mark complete, report success, STOP

## Arguments

`/plan-continue [document-path]`

- `document-path` (optional): Path to plan document. Resolution order if omitted:
  1. Check git config for associated plan (feature worktree)
  2. **Use current conversation context** - if a `/xonovex-workflow:plan-create`, `/xonovex-workflow:plan-research`, or similar command was run earlier in this conversation, continue from that research/plan. Do NOT search for other plans.
  3. Only if no context exists: Ask user which plan to continue

## Core Workflow

1. **Resolve Plan**: Check git config, then conversation context (do NOT search for plans)
2. **Load Parent Plan**: Parse plan document and extract metadata
3. **Find Next Actionable Plan**: Find first child plan with `pending` or `in_progress` status
4. **Load and Execute**: Read the target plan fully and implement it
5. **Complete and STOP**: Mark plan complete, update status, report success, then STOP

## Implementation Steps

**IMPORTANT: Work on ONE plan at a time. Do NOT read all child plans upfront.**

1. **Resolve Plan** (if no document-path provided):
   - First, check git config: `git config branch.$(git branch --show-current).plan`
   - If no git config, **use conversation context** - look at what was discussed/researched earlier in this conversation (e.g., `/plan-research-*` output, `/xonovex-workflow:plan-create` result). Implement that work directly.
   - Only if truly no context: Ask user which plan to work on. Do NOT glob/search for plans unprompted.

2. **Load Parent Plan Metadata-Only**: Read only the first 30 lines (front matter section)
   - Extract `has_subplans`, `status`, and `parallel_group` fields
   - Do NOT read the entire plan file

3. **Find Next Actionable Child Plan** (if `has_subplans: true`):
   - List child plan files in `plans/<plan-name>/` directory
   - For each child in order (01, 02, 03...), read only the first 30 lines to get `status`
   - Stop at FIRST plan with `status: "in_progress"` or `status: "pending"`
   - Do NOT read remaining child plans

4. **Determine Work Target**:
   - If parent plan `status: "complete"`: Warn user and exit
   - If found actionable child plan: That becomes the work target
   - If all checked child plans are complete: Recommend `/xonovex-workflow:plan-update` to mark parent complete
   - If no child plans: Parent plan is the work target

5. **Load Work Target Plan**: Read ONLY the selected plan in full (not all plans)
6. **Analyze Current Status**:
   - Extract phase, status, completed tasks from selected document
   - Identify pending tasks and blockers
7. **Validate State**:
   - Check if files mentioned in plan exist
   - Run type checking if applicable
   - Run linting to check code style
   - Run tests if test commands specified
   - Report baseline status (typecheck, lint, build, tests)
8. **Consult Skills**: Before writing ANY code, read the plan frontmatter `skills_to_consult` array and invoke EACH listed skill. For each invoked skill, also read the relevant progressive-disclosure documents it points to (its `references/*.md`), not just the top-level `SKILL.md` summary. This is a hard gate — implementation must not begin until the applicable skills and their reference docs have been loaded.
9. **Generate Next Steps**:
   - Identify highest priority pending tasks for current plan only
   - Extract file locations and required changes
   - Create implementation order based on dependencies
10. **Track Tasks**: Use the environment's task/todo tracking tool to create one entry per plan task. Update their status as work progresses — this prevents tasks from being silently skipped.
11. **Execute Plan**: Implement all tasks in the plan.
12. **On Completion**:

- Re-read the plan file (not from memory — read the file)
- Verify each task was done; do any that were missed
- Run each success criterion command; fix failures
- Run validation (typecheck, lint, build, tests)
- Report per-task DONE/SKIPPED and per-criterion PASS/FAIL
- Update plan status to `complete` in front matter
- **STOP** - do not continue to next plan automatically

## Output

**Initial Briefing**:

```
Continuing plan: plans/authentication.md

Plan: Authentication System
Progress: 2/4 child plans complete
Current: 03-api-routes.md (in_progress)

Validation:
  [PASS] Typecheck: PASS
  [PASS] Lint: PASS
  [PASS] Build: PASS
  [WARN] Tests: 2 failing (expected - not yet implemented)
```

**Completion Report**:

```
Completed plan: plans/authentication/03-api-routes.md

Implementation Summary:
- Added 3 API routes (login, logout, refresh)
- Created session middleware with JWT validation
- Implemented auth guards for protected routes

Validation:
  [PASS] Typecheck: PASS
  [PASS] Lint: PASS
  [PASS] Build: PASS
  [PASS] Tests: PASS (12 tests)

Status: in_progress -> complete
Parent: plans/authentication.md (3/4 complete)
```

## Error Handling

- If no git config and no conversation context: Ask user which plan to continue
- Error if specified document-path doesn't exist
- **Warning if plan `status: "complete"`** (all work done, recommend `/xonovex-workflow:plan-validate` or `/xonovex-workflow:plan-update`)
- **Info if all child plans complete** (recommend `/xonovex-workflow:plan-update` to mark parent complete)
- Error if plan file is malformed (invalid front matter)
- Warning if no plan associated with feature worktree

## Examples

```bash
# Auto-detect from git config (in feature worktree)
/xonovex-workflow:plan-continue

# Explicit path
/xonovex-workflow:plan-continue plans/authentication.md

# Continue specific subplan
/xonovex-workflow:plan-continue plans/authentication/subplan-02.md
```

## Gotchas

- Reading all child plans upfront wastes context — metadata-only scan first, then load only the target
- Skipping the validation step at start hides regressions introduced by previous plans — always baseline before working
- Auto-continuing to the next plan after completion silently chains work — STOP after one plan and let the user decide
- `skills_to_consult` listed in the plan must actually be invoked, and each skill's relevant reference docs read — they encode the project conventions the implementation must follow, and the `SKILL.md` summary alone often omits the detail
- Skipping the completion verification (re-reading the plan, checking each task and criterion) is the #1 cause of incomplete subplans — later tasks get compressed out of context and silently forgotten
- Use the environment's task/todo tracking tool — a pending task is visible, an untracked one is invisible
