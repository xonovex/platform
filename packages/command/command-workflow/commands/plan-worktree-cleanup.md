---
description: Remove stale and merged feature worktrees, and prune their leftover admin metadata
allowed-tools:
  - Bash
  - Skill
argument-hint: "[--merged] [--stale] [--prune] [--yes] [--dry-run]"
---

# /xonovex-workflow:plan-worktree-cleanup - Remove Stale and Merged Worktrees

## Arguments

`/plan-worktree-cleanup [--merged] [--stale] [--prune] [--yes] [--dry-run]`

- `--merged` (optional): Remove worktrees whose branch is already merged into its base
- `--stale` (optional): Remove worktrees whose branch or directory is gone or abandoned
- `--prune` (optional): Run `git worktree prune` to clear stale administrative metadata
- `--yes` (optional): Skip the confirmation and remove immediately
- `--dry-run` (optional): List what would be removed, change nothing

## Delegation

Load the `git-guide` skill (plugin `xonovex-skill-git`) and perform its
**worktree-cleanup** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas. Do not restate them.
