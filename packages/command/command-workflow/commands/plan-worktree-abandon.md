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

# /xonovex-workflow:plan-worktree-abandon – Abandon Feature with Documentation

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
abandoned_date: "2026-05-13"
```

**Content**:

```markdown
## Abandonment Notes

**Date**: 2026-05-13
**Reason**: Superseded by OAuth 2.0

[Detailed explanation and learnings]
```

## Output

```
Abandoning feature: auth-custom-jwt

Reason: Superseded by OAuth 2.0
Plan: Updated (status: abandoned)
State: Committed as "wip: abandoned work on auth-custom-jwt" (commit b4c5d6e)
Worktree: Kept (run worktree-cleanup to remove)

Next Steps:
1. Extract insights from the abandoned work (see skill-insights)
2. Review plan's "Abandonment Notes" section
3. Clean up worktree later if desired
```

## Error Handling

- Error: not in feature worktree, no reason, plan doesn't exist, worktree removal fails
- Warning: no plan associated (unless `--no-plan`), uncommitted changes (suggest `--commit`)

## Gotchas

- Abandoning without recording a reason loses the learning — always capture _why_ the approach didn't work
- Default behaviour keeps the worktree intact — insights can still be extracted from the branch later
- A plan marked `abandoned` is the canonical signal downstream operations should respect — don't reuse the same plan file for a fresh attempt without bumping a new path
- `git worktree remove` doesn't delete the branch ref — pair with `git branch -D <feature-branch>` (or tag it as `abandoned/<name>` before deleting) if you want to preserve the work in git history

## Examples

```bash
/xonovex-workflow:plan-worktree-abandon "Superseded by better approach"
/xonovex-workflow:plan-worktree-abandon "Requirements changed" --commit --remove-worktree
/xonovex-workflow:plan-worktree-abandon "Quick experiment" --no-plan
/xonovex-workflow:plan-worktree-abandon "Blocked by dependency" --commit

# Preview changes without modifying files
/xonovex-workflow:plan-worktree-abandon "Superseded by OAuth 2.0" --dry-run
```
