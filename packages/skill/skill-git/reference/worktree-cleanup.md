# worktree-cleanup: Clean Up Stale and Merged Worktrees

**Guideline:** Remove stale and merged feature worktrees to keep workspace clean and organized.

**Rationale:** Prevents workspace clutter from old feature worktrees by identifying merged worktrees (fully integrated) and stale worktrees (30+ days without changes).

**Example:**

```bash
# List all worktrees and their status
git worktree list

# Output:
# /home/user/project             9a3c5e2 [master]
# /home/user/project-feature-auth 7f2b1d8 [feature/auth-redesign] (merged)
# /home/user/project-feature-db   c4a8e1f [feature/db-migration] (30+ days no commits)

# Cleanup will identify and remove merged/stale worktrees
# Merged: feature/auth-redesign (fully merged into master)
# Stale: feature/db-migration (70 days since last commit)

# Remove merged worktree
git worktree remove ../project-feature-auth
git branch -d feature/auth-redesign

# Remove stale worktree (but keep if uncommitted changes)
git worktree remove ../project-feature-db --force
git branch -d feature/db-migration

# Verify cleanup
git worktree list
# Only main worktree remains
```

**Techniques:**

- List worktrees: `git worktree list`
- Identify feature worktrees with `*-feature-*` pattern
- Check merge status: `git branch --merged <source>`
- Categorize: merged, stale (>30 days no commits), active
- Remove selected: `git worktree remove <path>` + `git branch -d <branch>`
