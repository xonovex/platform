# worktree-cleanup: Clean Up Stale and Merged Worktrees

Remove feature worktrees (dir pattern `*-feature-*`) that are merged or stale (>30 days without commits); keep active ones.

```bash
git worktree list                          # inventory
git branch --merged <source>               # which feature branches are merged
git worktree remove <path>                 # add --force only if it has uncommitted changes
git branch -d <branch>                     # remove the merged ref too (-D to force)
```
