# worktree-merge: Merge Feature Worktree Back to Source

Merge a feature branch from a `<worktree>-feature-*` worktree back into its source branch in the sibling source worktree.

## Steps

1. Validate: pwd basename matches `*-feature-*`, feature branch matches `<worktree>/feature/*`.
2. Source branch: `git config branch.<feature-branch>.mergeBackTo` (error if missing — run `worktree-create` first).
3. Source worktree: sibling `../<base-worktree>` (verify exists and holds the source branch).
4. Require the feature worktree clean (`git status` — error if dirty).
5. In the source worktree, `git checkout <source-branch>` then merge:

```bash
git merge <feature-branch>                 # regular
# or squash:
git merge --squash <feature-branch> && git commit -m "<type>: <feature-name>

Squashed commits from <worktree>/feature/<feature-name>"
```

6. Optional cleanup: `git worktree remove <feature-worktree-path>`; remote delete: `git push origin --delete <feature-branch>`.

**Never** remove the worktree if the merge fails.

## Gotchas

- Uncommitted changes in the feature worktree silently ride along in some merge strategies — confirm clean first
- `--squash` loses individual commit history — pick it only when commit-by-commit replay isn't valuable
- `git worktree remove` refuses a dirty worktree — commit, stash, or `--force` (destructive)
- The feature branch ref persists after `worktree remove` — `git branch -d <feature-branch>` to clean it up
- Merging without pulling the source worktree first can produce stale-base merges when collaborating
