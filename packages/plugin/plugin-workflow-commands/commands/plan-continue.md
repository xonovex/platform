---
description: Resume work from an existing plan document with full context loading
allowed-tools:
  - Read
  - Bash
  - Glob
  - TaskCreate
  - TaskUpdate
  - TaskList
  - Task
argument-hint: "[document-path]"
---

# /plan-continue – Continue Progress from Plan

Resumes work from an existing plan document by loading its context and outlining the next steps. Auto-detects associated plan from git config if available.

## Goal

**Start working IMMEDIATELY. No questions, no asking for permission, no confirmations.**

**ONE plan only. When complete, STOP. User re-runs `/plan-continue` for next plan.**

- Auto-detect plan from git config or use provided path
- Find first actionable child plan (if parent has subplans)
- Execute that single plan to completion
- Mark complete, report success, STOP

## Arguments

`/plan-continue [document-path]`

- `document-path` (optional): Path to plan document. Resolution order if omitted:
  1. Check git config for associated plan (feature worktree)
  2. **Use current conversation context** - if a `/plan-create`, `/plan-research-*`, or similar command was run earlier in this conversation, continue from that research/plan. Do NOT search for other plans.
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
   - If no git config, **use conversation context** - look at what was discussed/researched earlier in this conversation (e.g., `/plan-research-*` output, `/plan-create` result). Implement that work directly.
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
   - If all checked child plans are complete: Recommend `/plan-update` to mark parent complete
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
8. **Consult Skills**: Check plan frontmatter for `skills_to_consult` array and invoke relevant skills before implementation
9. **Generate Next Steps**:
   - Identify highest priority pending tasks for current plan only
   - Extract file locations and required changes
   - Create implementation order based on dependencies
10. **Create Tasks**: Use TaskCreate to set up actionable tasks with status tracking; update with TaskUpdate as work progresses
11. **Execute Plan**: Implement all tasks in the plan
12. **On Completion**:

- Update plan status to `complete` in front matter
- Run validation (typecheck, lint, build, tests)
- Report completion summary
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

Next Steps:
1. Implement POST /api/auth/login route
2. Add session middleware
3. Create auth guards

Files to modify:
- src/routes/auth/login.ts
- src/middleware/session.ts
- src/guards/auth.ts
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

Next Steps:
1. Run /plan-continue for next child plan
2. Or /plan-status to check overall progress
```

## Error Handling

- If no git config and no conversation context: Ask user which plan to continue
- Error if specified document-path doesn't exist
- **Warning if plan `status: "complete"`** (all work done, recommend `/plan-validate` or `/plan-update`)
- **Info if all child plans complete** (recommend `/plan-update` to mark parent complete)
- Error if plan file is malformed (invalid front matter)
- Warning if no plan associated with feature worktree

## Examples

```bash
# Auto-detect from git config (in feature worktree)
/plan-continue

# Explicit path
/plan-continue plans/authentication.md

# Continue specific subplan
/plan-continue plans/authentication/subplan-02.md
```
