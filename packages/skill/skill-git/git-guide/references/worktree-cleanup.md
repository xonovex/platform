# worktree-cleanup: Clean Up Stale and Merged Worktrees

Remove feature worktrees (dir pattern `*-feature-*`) that are merged or stale (>30 days without commits); keep active ones.

List the worktrees and branch refs that would be removed, and get confirmation, before removing any — `git worktree remove --force` discards uncommitted work and `git branch -D` discards unmerged commits. Preview-only when asked; skip the confirmation only when explicitly told to.

```bash
git worktree list                          # inventory
git branch --merged <source>               # which feature branches are merged
git worktree remove <path>                 # add --force only if it has uncommitted changes
git branch -d <branch>                     # remove the merged ref too (-D to force)
```
