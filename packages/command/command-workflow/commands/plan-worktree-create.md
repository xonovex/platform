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

Load the `git-guide` skill (plugin `xonovex-skill-git`) and perform its
**worktree-create** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas — do not restate them.
