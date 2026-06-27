# plan-update: Update Plan Document with Current Progress

Update a plan document with the latest implementation status, test results, and progress.

## Goal

1. **Identify** the target plan document (auto-detect from git config or user message)
2. **Update the plan** with validation results and current progress
3. **Check child plan status** (if target has child plans)
4. **Auto-update parent** (if target has a parent plan)
5. **Update frontmatter** — `status`, `phase`, `updated` date
6. **Timestamp** the update for progress tracking

## Core Workflow

1. **Identify plan** — auto-detect from git config or use provided path
2. **Check type** — has a parent plan? has child plans?
3. **Update status** — run validation, update frontmatter and status section
4. **Update parent** — if has parent plan, auto-update parent

## Updates

### Frontmatter

`updated`, `status` (in_progress / complete / blocked), `phase`, `completed_date` (if complete)

### Status Section

timestamped update with phase, summary, files modified, validation, next steps

## Status Logic

**Child plans** (`type: plan`, has `parent_plan`):

- Run validation tests → update status (complete / in_progress / blocked)
- Auto-update parent: read `parent_plan`, check all siblings, update parent status

**Parent plans** (`type: plan`, `has_subplans: true`):

- Has child plans: all complete → `complete`; any blocked → `blocked`; else → `in_progress`
- No child plans: run validation → update based on results

## Output

```
Updated plan: plans/auth.md

Status: in_progress → complete
Phase: Implementation → Complete
Validation: PASS (all tests passing)

Parent plan updates: None (this is a parent plan)
Files modified: 5 files updated
```

## Error Handling

- Plan file not found → error
- Invalid format (missing frontmatter) → error
- No changes detected → warning
- Parent plan not found (for child plans) → error

## Gotchas

- Setting `status: complete` without running validation produces optimistic plan state — always validate first
- Auto-updating the parent based on stale child status leaves the parent inconsistent — re-check children before bumping parent
- `completed_date` set without rolling forward across siblings makes timeline reconstruction painful — keep timestamps consistent
- Manual edits to the Status Section get overwritten on next update — use frontmatter `notes` for content that must persist
