# worktree-merge: Merge Feature Worktree Back to Source

Merge a feature branch from a feature worktree back to its source worktree.

## Goal

- Detect the current feature worktree directory
- Find the source worktree directory
- Retrieve source branch from git config (stored during creation)
- Merge the feature branch into the source branch in the source worktree
- Support both regular merge and squash merge
- Optionally remove the feature worktree after merging

## Core Workflow

1. **Validate Feature Worktree** — verify we're in a `<worktree>-feature-*` directory
2. **Detect Worktree Names** — extract base worktree and feature name from directory
3. **Retrieve Source Branch** — `git config branch.<feature-branch>.mergeBackTo`
4. **Navigate to Source** — find and `cd` to the source worktree
5. **Merge Feature** — regular (`git merge`) or squash (`git merge --squash` + commit); preview-only if requested
6. **Clean Up** — optionally remove the feature worktree after merge (skipped if preview)

## Worktree Detection

Automatically detected:

- Current directory name → feature worktree (e.g. `services-feature-hono-openapi`)
- Base worktree name (e.g. `services`)
- Source worktree directory (sibling, e.g. `../services`)
- Feature branch from `git branch --show-current`
- Source branch from `git config branch.<feature-branch>.mergeBackTo`

## Implementation Steps

1. `pwd` + `basename` to get directory name
2. Validate directory matches `*-feature-*` pattern
3. Extract names: e.g. `services-feature-hono-openapi` → worktree `services`, feature `hono-openapi`
4. `git branch --show-current` for feature branch
5. Validate feature branch format `<worktree>/feature/*`
6. Get source branch: `git config branch.<current-branch>.mergeBackTo` (error if missing)
7. Source worktree path: `../<base-worktree>` (verify exists)
8. `git -C <source-worktree> rev-parse --verify <source-branch>`
9. Check for uncommitted changes in feature worktree (error if dirty)
10. `cd <source-worktree>`
11. `git checkout <source-branch>`
12. Merge:
    - Regular: `git merge <feature-branch>`
    - Squash: `git merge --squash <feature-branch>` then `git commit -m "<message>"`
13. If user requested removal: `git worktree remove <feature-worktree-path>`
14. If user requested remote deletion: `git push origin --delete <feature-branch>`

## Squash Commit Message

```
<type>: <feature-name-as-description>

Squashed commits from <worktree>/feature/<feature-name>
```

Example: `feat: introduce hono openapi\n\nSquashed commits from services/feature/hono-openapi`

## Output

```
Merging feature: services-feature-auth-fix

Detected feature worktree: services-feature-auth-fix
Base worktree: services
Feature branch: services/feature/auth-fix
Source worktree: /home/user/projects/services
Source branch: master (from git config)

Merge type: Regular merge
Merge status: Success (no conflicts)
Worktree cleanup: Kept (run worktree-cleanup to remove)
```

## Error Handling

- Not in a feature worktree (`*-feature-*` pattern) → error
- Not on a feature branch (`<worktree>/feature/*`) → error
- No `mergeBackTo` in git config → error (run worktree-create first)
- Source worktree directory missing → error
- Source branch missing in source worktree → error
- Uncommitted changes in feature worktree → error
- Merge conflicts → error with conflicted files + guidance
- Trying to delete a remote branch that doesn't exist → error
- **Never** proceed with worktree removal if the merge fails

## Gotchas

- Uncommitted changes in the feature worktree silently get included in some merge strategies — always check `git status` clean before merging
- The `--squash` flow loses individual commit history — pick this only when commit-by-commit replay isn't valuable
- `git worktree remove` won't run on a dirty worktree — commit, stash, or use `--force` (the latter is destructive)
- The feature branch persists after `worktree remove` — `git branch -d <feature-branch>` to clean up the ref too
- Forgetting to `git pull` the source worktree before merging can produce stale-base merges — pull first when collaborating
