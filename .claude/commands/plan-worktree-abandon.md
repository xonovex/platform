---
description: Document and abandon a feature with reason and learnings
model: haiku
allowed-tools:
  - Bash
  - Read
  - Edit
  - Write
argument-hint: "[reason] [--remove-worktree] [--no-plan] [--commit] [--dry-run]"
---

# /plan-worktree-abandon – Abandon Feature with Documentation

Abandon feature by updating plan status, documenting reason, and optionally cleaning up worktree.

## Goal

- Update plan status to "abandoned" with reason and learnings
- Optionally commit current state and remove worktree
- Keep worktree by default for later insights extraction

## Arguments

- `reason` (optional): Concise reason for abandonment (prompted if not provided)
- `--remove-worktree` (optional): Remove feature worktree after documenting
- `--no-plan` (optional): Skip plan update (for features without plans)
- `--commit` (optional): Commit current state before abandoning
- `--dry-run` (optional): Preview changes without modifying files

## Core Workflow

1. Verify in feature worktree (basename matches `*-feature-*`)
2. Get reason from argument or prompt user
3. Detect plan via `git config branch.<branch>.plan`
4. Optionally commit: `git add . && git commit -m "wip: abandoned work on <feature>"` (skipped if --dry-run)
5. Update plan frontmatter and add "Abandonment Notes" section (or preview if --dry-run)
6. Keep worktree by default (or remove if `--remove-worktree`, skipped if --dry-run)

## Implementation Steps

1. Verify feature worktree and get current branch
2. Read plan path from git config
3. Get abandonment reason (arg or prompt)
4. Optionally commit (if `--commit` and uncommitted changes)
5. Update plan (unless `--no-plan`):
   - Frontmatter: `status: "abandoned"`, `abandoned_reason`, `abandoned_date`
   - Content: Add "Abandonment Notes" with detailed explanation and learnings
6. Optionally remove worktree: `git worktree remove <path>`

## Plan Updates

**Frontmatter**:

```yaml
status: "abandoned"
abandoned_reason: "Superseded by OAuth 2.0"
abandoned_date: "2025-11-08"
```

**Content**:

```markdown
## Abandonment Notes

**Date**: 2025-11-08
**Reason**: Superseded by OAuth 2.0

[Detailed explanation and learnings]
```

## Output

```
Abandoning feature: auth-custom-jwt

Reason: Superseded by OAuth 2.0
Plan: Updated (status: abandoned)
State: Committed as "wip: abandoned work on auth-custom-jwt" (commit b4c5d6e)
Worktree: Kept (cleanup with /plan-worktree-cleanup)

Next Steps:
1. Extract insights: /insights-extract abandoned - Capture learnings from abandoned work
2. Review plan: Check plan's "Abandonment Notes" section for documented reason
3. Clean up worktree: /plan-worktree-cleanup - Remove abandoned worktree (if kept)
4. Alternative approach: Create new plan/feature if different approach needed
5. Share learnings: Document insights in team knowledge base if applicable
```

## Error Handling

- Error: not in feature worktree, no reason, plan doesn't exist, worktree removal fails
- Warning: no plan associated (unless `--no-plan`), uncommitted changes (suggest `--commit`)

## Examples

```bash
/plan-worktree-abandon "Superseded by better approach"
/plan-worktree-abandon "Requirements changed" --commit --remove-worktree
/plan-worktree-abandon "Quick experiment" --no-plan
/plan-worktree-abandon "Blocked by dependency" --commit

# Preview changes without modifying files
/plan-worktree-abandon "Superseded by OAuth 2.0" --dry-run
```
