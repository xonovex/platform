---
description: Update a plan document with the latest implementation status and test results
allowed-tools:
  - Read
  - Edit
  - Bash
  - Glob
  - TaskUpdate
  - TaskList
argument-hint: "[document-path] [--dry-run]"
---

# /plan-update – Update Plan Progress

Updates a plan document with the latest implementation status, test results, and progress.

## Goal

1. **Identifies** the target plan document (auto-detect from git config or argument)
2. **Updates the plan** with validation results and current progress
3. **Checks child plan status** (if target has child plans)
4. **Auto-updates parent** (if target has a parent plan)
5. **Updates frontmatter** (`status`, `phase`, `updated` date)
6. **Timestamps** the update for progress tracking

## Arguments

`/plan-update [document-path] [--dry-run]`

- `document-path` (optional): Path to plan document (auto-detects from git config if omitted)
- `--dry-run` (optional): Preview changes without modifying files

## Core Workflow

1. **Identify Plan**: Auto-detect from git config or use provided path
2. **Check Type**: Determine if has parent plan or child plans
3. **Update Status**: Run validation, update frontmatter and status section
4. **Update Parent**: If has parent plan, auto-update parent

## Updates

**Frontmatter**: `updated`, `status` (in_progress/complete/blocked), `phase`, `completed_date` (if complete)

**Status Section**: Timestamped update with phase, summary, files modified, validation, next steps

## Status Logic

**Child plans** (`type: "plan"`, has `parent_plan`):

- Run validation tests -> update status (complete/in_progress/blocked)
- Auto-update parent: Read `parent_plan` -> check all siblings -> update parent status

**Parent plans** (`type: "plan"`, `has_subplans: true`):

- If has child plans: All complete -> `complete` | Any blocked -> `blocked` | Else -> `in_progress`
- If no child plans: Run validation -> update based on results

## Output

```
Updated plan: plans/auth.md

Status: in_progress -> complete
Phase: Implementation -> Complete
Validation: PASS (all tests passing)

Parent plan updates: None (this is a parent plan)
Files modified: 5 files updated

Next Steps:
1. Check status: Review updated plan status and validation results
2. If child plans remain: /plan-continue plans/<plan>.md - Continue with next pending child plan
3. If all child plans complete: /plan-worktree-validate - Run pre-merge validation
4. If blocked: Fix identified issues, then /plan-validate to verify
5. Track progress: /plan-status plans/<plan>.md - Monitor overall completion
```

## Error Handling

- Error if plan file not found
- Error if plan format invalid (missing frontmatter)
- Warning if no changes detected
- Error if parent plan not found (for child plans)

## Examples

```bash
# Auto-detect from git config (in feature worktree)
/plan-update

# Explicit path
/plan-update plans/auth.md

# Update child plan (auto-updates parent)
/plan-update plans/auth/02-implementation.md

# Preview changes
/plan-update plans/auth.md --dry-run
```
