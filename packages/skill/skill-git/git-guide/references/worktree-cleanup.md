# worktree-cleanup: Clean Up Stale and Merged Worktrees

Preview or apply removal of exact feature worktree and branch targets. Default to
`preview`; never turn a broad pattern or age heuristic directly into a destructive
command.

## Procedure

1. Inventory worktrees and branches, then resolve exact requested targets. A
   `*-feature-*` pattern or more than 30 days without commits is only a candidate
   signal.
2. For each target, inspect dirty state, current revision, upstream publication,
   merge status against the recorded source, and whether another worktree uses the
   branch.
3. Protect active, dirty, unmerged, unpushed, current, and unresolved targets by
   default. Report the recovery revision and commands before any deletion.
4. In `preview`, list the exact worktree paths, local refs, remote refs, commands,
   recovery information, and reasons for inclusion or exclusion without mutating.
5. In explicit `apply`, revalidate the previewed identities and state, remove only
   the authorized worktree paths and refs, then verify they are gone.
6. Require separate explicit authorization for forced removal or remote ref deletion.

## Apply Commands

```bash
git worktree remove <exact-path>
git branch -d <exact-branch>
git push <exact-remote> --delete <exact-branch>
```

## Gotchas

- `git worktree remove` does not delete the branch ref
- `-D`, `--force`, and remote deletion can make unmerged work difficult to recover
- `git worktree prune` removes stale administrative entries; it is not a substitute for resolving exact cleanup targets
