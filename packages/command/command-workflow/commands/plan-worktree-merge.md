---
description: Merge feature worktree back to source worktree
model: haiku
allowed-tools:
  - Bash
argument-hint: "[--squash] [--remove-worktree] [--delete-remote] [--dry-run]"
---

# /xonovex-workflow:plan-worktree-merge – Merge Feature Worktree Back to Source

Merges a feature branch from a feature worktree back to its source worktree.

## Goal

- Detect current feature worktree directory
- Find the source worktree directory
- Retrieve source branch from git config (stored during creation)
- Merge feature branch into source branch in source worktree
- Support both regular merge and squash merge
- Optionally remove feature worktree after merge

## Arguments

`/plan-worktree-merge [--squash] [--remove-worktree] [--delete-remote] [--dry-run]`

- `--squash`: Squash all feature commits into a single commit
- `--remove-worktree`: Remove feature worktree directory after merge (default: keep)
- `--delete-remote`: Also delete remote feature branch (requires remote exists)
- `--dry-run` (optional): Preview changes without modifying files

## Core Workflow

1. **Validate Feature Worktree**: Verify we're in a `<worktree>-feature-*` directory
2. **Detect Worktree Names**: Extract base worktree name and feature name from directory
3. **Retrieve Source Branch**: Get from `git config branch.<feature-branch>.mergeBackTo`
4. **Navigate to Source**: Find and cd to source worktree directory
5. **Merge Feature**: Regular (`git merge`) or squash (`git merge --squash` + commit), or preview if --dry-run
6. **Clean Up**: Optionally remove feature worktree if `--remove-worktree` (skipped if --dry-run)

## Worktree Detection

The command automatically detects:

- Current directory name to identify feature worktree (e.g., `services-feature-hono-openapi`)
- Base worktree name (e.g., `services`)
- Source worktree directory (sibling directory, e.g., `../services`)
- Feature branch from `git branch --show-current`
- Source branch from `git config branch.<feature-branch>.mergeBackTo`

## Implementation Steps

1. **Get current directory**: `pwd` to get full path, then `basename` to get directory name
2. **Validate feature worktree**: Ensure directory name matches `*-feature-*` pattern
3. **Extract names**:
   - Parse directory name to get base worktree and feature name
   - Example: `services-feature-hono-openapi` -> worktree: `services`, feature: `hono-openapi`
4. **Get current branch**: `git branch --show-current`
5. **Validate feature branch**: Ensure format is `<worktree>/feature/*`
6. **Get source branch**: `git config branch.<current-branch>.mergeBackTo` or error
7. **Find source worktree**: Check for `../<base-worktree>` directory
8. **Verify source exists**: `git -C <source-worktree> rev-parse --verify <source-branch>`
9. **Check for uncommitted changes**: In current feature worktree, error if uncommitted changes exist
10. **Navigate to source**: `cd <source-worktree>`
11. **Checkout source branch**: `git checkout <source-branch>`
12. **Merge feature**:
    - Regular: `git merge <feature-branch>`
    - Squash: `git merge --squash <feature-branch>` then `git commit -m "<message>"`
13. **Remove worktree**: `git worktree remove <feature-worktree-path>` (only if `--remove-worktree`)
14. **Delete remote branch**: `git push origin --delete <feature-branch>` (if `--delete-remote`)

## Squash Commit Message

When using `--squash`, generate commit message in format:

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

## Examples

```bash
# Merge feature branch back to source
/xonovex-workflow:plan-worktree-merge

# Squash and merge
/xonovex-workflow:plan-worktree-merge --squash

# Merge and clean up worktree
/xonovex-workflow:plan-worktree-merge --remove-worktree --delete-remote

# Preview merge without making changes
/xonovex-workflow:plan-worktree-merge --dry-run
```

## Error Handling

- Error if not in a feature worktree directory (`*-feature-*` pattern)
- Error if not on a feature branch (`<worktree>/feature/*`)
- No `mergeBackTo` in git config → error (run worktree-create first)
- Error if source worktree directory doesn't exist
- Error if source branch doesn't exist in source worktree
- Error if uncommitted changes exist in feature worktree
- Error if merge conflicts occur - show conflicted files and guidance
- Error if trying to delete remote branch that doesn't exist
- **Never** proceed with worktree removal if the merge fails

## Gotchas

- Uncommitted changes in the feature worktree silently get included in some merge strategies — always check `git status` clean before merging
- The `--squash` flow loses individual commit history — pick this only when commit-by-commit replay isn't valuable
- `git worktree remove` won't run on a dirty worktree — commit, stash, or use `--force` (the latter is destructive)
- The feature branch persists after `worktree remove` — `git branch -d <feature-branch>` to clean up the ref too
- Forgetting to `git pull` the source worktree before merging can produce stale-base merges — pull first when collaborating
