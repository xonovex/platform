# worktree-abandon: Abandon Feature Worktree with Documented Reason

Abandon a feature: update the plan status, document the reason and learnings, and optionally commit current state and remove the worktree.

## Goal

- Update plan status to `abandoned` with reason and learnings
- Optionally commit current state and remove worktree
- Keep worktree by default for later insights extraction

## Core Workflow

1. Verify in a feature worktree (basename matches `*-feature-*`)
2. Get reason from user message or prompt
3. Detect plan via `git config branch.<branch>.plan`
4. Optionally commit: `git add . && git commit -m "wip: abandoned work on <feature>"` (skipped if preview)
5. Update plan frontmatter and add "Abandonment Notes" section (or preview if requested)
6. Keep worktree by default; remove only if user requests removal

## Implementation Steps

1. Verify feature worktree and get current branch
2. Read plan path from git config
3. Get abandonment reason from user message or prompt
4. Optionally commit (if user requested commit and there are uncommitted changes)
5. Update plan (unless user asked to skip plan update):
   - Frontmatter: `status: "abandoned"`, `abandoned_reason`, `abandoned_date`
   - Content: add "Abandonment Notes" with detailed explanation and learnings
6. Optionally remove worktree: `git worktree remove <path>`

## Plan Updates

### Frontmatter

```yaml
status: "abandoned"
abandoned_reason: "Superseded by OAuth 2.0"
abandoned_date: "2026-05-13"
```

### Content

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
1. Extract insights from the abandoned work (see skill-reflect)
2. Review plan's "Abandonment Notes" section
3. Clean up worktree later if desired
```

## Error Handling

- Not in a feature worktree → error
- No reason provided → error or prompt
- Plan doesn't exist → warn (unless user asked to skip plan update)
- Worktree removal fails → error
- Uncommitted changes present and no commit requested → warn (suggest committing first)

## Gotchas

- Abandoning without recording a reason loses the learning — always capture _why_ the approach didn't work
- Default behaviour keeps the worktree intact — insights can still be extracted from the branch later
- A plan marked `abandoned` is the canonical signal downstream operations should respect — don't reuse the same plan file for a fresh attempt without bumping a new path
- `git worktree remove` doesn't delete the branch ref — pair with `git branch -D <feature-branch>` (or tag it as `abandoned/<name>` before deleting) if you want to preserve the work in git history
