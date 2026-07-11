# plan-update: Update Plan Document with Current Progress

Update a plan document with the latest implementation status, validation results, and progress. Auto-detect the target from git config or the user message.

## Core Workflow

1. **Identify plan** — git config or provided path; check whether it has a parent and/or child plans
2. **Run validation** and update the frontmatter + status section
3. **Auto-update parent** — if the target has a `parent_plan`, re-check all siblings and roll up

## Updates

- **Frontmatter** — `updated`, `status` (in_progress / complete / blocked), `phase`, `completed_date` (if complete)
- **Status Section** — timestamped update with phase, summary, files modified, validation, next steps

## Status Logic

- **Child plan** (has `parent_plan`) — run validation → set status; then read `parent_plan`, check all siblings, roll up
- **Parent plan** (`has_subplans: true`) — with children: all complete → `complete`, any blocked → `blocked`, else `in_progress`; no children: run validation → set from results

## Gotchas

- Setting `status: complete` without running validation produces optimistic plan state — always validate first
- Rolling up the parent from stale child status leaves it inconsistent — re-check children before bumping
- Manual edits to the Status Section get overwritten on next update — use frontmatter `notes` for content that must persist
