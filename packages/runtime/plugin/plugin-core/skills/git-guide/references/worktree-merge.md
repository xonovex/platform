# worktree-merge: Merge Feature Worktree Back to Source

Preview or apply integration of a feature branch from a `<worktree>-feature-*` worktree into its recorded source branch. Default to `preview`; only explicit `apply` may change refs or worktrees.

## Procedure

1. Resolve and report the exact feature worktree, feature branch, feature revision, recorded source branch, source worktree, source revision, remote, and merge strategy.
2. Validate that the directory and branch naming match, `branch.<feature-branch>.mergeBackTo` exists, both worktrees are clean, the source worktree holds the source branch, and the destination is not a protected mainline.
3. Fetch remote state for an informed preview only when the caller authorizes that external read. Determine whether the feature must be rebased and whether the merge would be fast-forward, regular, squash, conflicting, or empty.
4. In `preview`, return the exact rebase and merge commands, changed refs, conflict risk, validation, and recovery commands without running them.
5. In explicit `apply`, revalidate the previewed revisions, rebase the feature onto the current source, merge with the selected strategy, and run repository validation:

```bash
git merge <feature-branch>                 # regular
# or squash:
git merge --squash <feature-branch> && git commit -m "<type>: <feature-name>

Squashed commits from <worktree>/feature/<feature-name>"
```

6. Verify the destination revision and validation evidence. Preserve the feature worktree, local branch, and remote branch; removal belongs only to `worktree-cleanup`.

## Gotchas

- Uncommitted changes in the feature worktree silently ride along in some merge strategies: confirm clean first
- `--squash` loses individual commit history: pick it only when commit-by-commit replay isn't valuable
- A preview based on stale remote state is incomplete: report when fetch was not authorized or available
- Skipping the rebase produces stale-base merges when collaborating: it is required, not optional
- Never remove a worktree or branch during merge, whether the merge succeeds or fails
