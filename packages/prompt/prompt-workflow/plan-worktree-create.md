---
description: Create a new git worktree for a feature branch
model: haiku
allowed-tools:
  - Bash
  - Glob
  - Read
argument-hint: "[feature-name] [--from <branch>]"
---

# /plan-worktree-create – Create Feature Worktree

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

- In `services` worktree with feature `auth-fix`:
  - Directory: `services-feature-auth-fix`
  - Branch: `services/feature/auth-fix`
- In `api` worktree with feature `new-endpoint`:
  - Directory: `api-feature-new-endpoint`
  - Branch: `api/feature/new-endpoint`

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
2. Associate plan: /plan-worktree-set-plan <plan-dir>/<plan>.md (if working with a plan)
3. Start work: /plan-continue (after associating plan)
4. Make commits: /plan-worktree-commit
5. Validate: /plan-worktree-validate
```

## Error Handling

- Error if feature name not provided
- Error if worktree directory already exists
- Error if branch already exists
- Error if source branch doesn't exist (when using `--from`)
