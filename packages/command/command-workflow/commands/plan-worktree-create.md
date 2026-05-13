---
description: Create a new git worktree for a feature branch
model: haiku
allowed-tools:
  - Bash
  - Glob
  - Read
argument-hint: "[feature-name] [--from <branch>]"
---

# /xonovex-workflow:plan-worktree-create – Create Feature Worktree

Creates a new git worktree directory with a feature branch, allowing isolated development without affecting the current worktree.

## Goal

- Detect current worktree name from directory path
- Create a new sibling worktree directory with a feature branch
- Follow naming pattern: `<worktree>-feature-<name>` directory with `<worktree>/feature/<name>` branch
- Store source branch association in git config

## Arguments

`/plan-worktree-create [feature-name] [--from <branch>]`

- `feature-name`: Name of the feature (required) - used in directory and branch names
- `--from <branch>`: Create feature from specific branch instead of current branch

## Core Workflow

1. **Detect Worktree**: Extract worktree name from current directory path
2. **Get Source Branch**: Use `--from` argument or current branch via `git branch --show-current`
3. **Create Worktree**: Use `git worktree add` to create new directory with feature branch
4. **Store Association**: Save source branch in git config

## Naming Convention

- Worktree directory: `<worktree>-feature-<feature-name>`
- Branch name: `<worktree>/feature/<feature-name>`

Examples:

| In worktree | Feature        | Directory                   | Branch                      |
| ----------- | -------------- | --------------------------- | --------------------------- |
| `services`  | `auth-fix`     | `services-feature-auth-fix` | `services/feature/auth-fix` |
| `api`       | `new-endpoint` | `api-feature-new-endpoint`  | `api/feature/new-endpoint`  |

## Implementation Steps

1. **Validate**: Ensure feature name is provided
2. **Get current directory**: Use `pwd` to get full path
3. **Extract worktree name**: Parse basename from path (e.g., `/path/to/services` -> `services`)
4. **Get source branch**: Use `--from` argument or `git branch --show-current`
5. **Sanitize feature name**: Convert to kebab-case, remove special chars
6. **Construct names**:
   - Worktree dir: `../<worktree>-feature-<sanitized-name>`
   - Branch name: `<worktree>/feature/<sanitized-name>`
7. **Create worktree**: `git worktree add <worktree-dir> -b <branch-name> <source-branch>`
8. **Store association**:
   - `cd <worktree-dir>`
   - `git config branch.<branch-name>.mergeBackTo <source-branch>`

## Output

```
Created feature worktree: services-feature-auth-fix

Detected worktree: services
Source branch: master

Created worktree: /home/user/projects/services-feature-auth-fix
Created branch: services/feature/auth-fix

Stored associations:
- Source branch: master

Next Steps:
1. Navigate: cd ../services-feature-auth-fix
2. Associate plan (if working with a plan): set git config branch.<branch>.plan
3. Start work in the new worktree
```

## Error Handling

- Error if feature name not provided
- Error if worktree directory already exists
- Error if branch already exists
- Error if source branch doesn't exist (when using `--from`)

## Gotchas

- The directory naming pattern `<worktree>-feature-<name>` is what downstream operations (merge, abandon) detect — non-conforming names break the workflow
- `mergeBackTo` git config is the only record of the source branch — without it, `worktree-merge` can't find where to merge back
- Creating a worktree from a dirty source branch carries those uncommitted changes into the feature worktree only if they're staged — verify clean state first
- A second worktree can't check out the same branch as another worktree — try a different source branch or move the existing worktree
