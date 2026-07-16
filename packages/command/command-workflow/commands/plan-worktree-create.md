---
description: Create a new git worktree for a feature branch
allowed-tools:
  - Bash
  - Glob
  - Read
  - Skill
argument-hint: "[feature-name] [--from <branch>]"
---

# /xonovex-workflow:plan-worktree-create — Create Feature Worktree

## Arguments

`/plan-worktree-create [feature-name] [--from <branch>]`

- `feature-name`: Name of the feature (required) - used in directory and branch names
- `--from <branch>`: Create feature from specific branch instead of current branch

## Delegation

Load an installed capability that provides Git worktree creation and perform its
worktree-create operation. This workspace method is a soft dependency: if unavailable,
stop and name the capability to install; do not make Git worktrees a Planning prerequisite.
