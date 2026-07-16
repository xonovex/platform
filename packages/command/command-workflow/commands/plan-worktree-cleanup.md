---
description: Remove stale and merged feature worktrees, and prune their leftover admin metadata
allowed-tools:
  - Bash
  - Skill
argument-hint: "[--merged] [--stale] [--prune] [--yes] [--dry-run]"
---

# /xonovex-workflow:plan-worktree-cleanup — Remove Stale and Merged Worktrees

## Arguments

`/plan-worktree-cleanup [--merged] [--stale] [--prune] [--yes] [--dry-run]`

- `--merged` (optional): Remove worktrees whose branch is already merged into its base
- `--stale` (optional): Remove worktrees whose branch or directory is gone or abandoned
- `--prune` (optional): Run `git worktree prune` to clear stale administrative metadata
- `--yes` (optional): Skip the confirmation and remove immediately
- `--dry-run` (optional): List what would be removed, change nothing

## Delegation

Load an installed capability that provides safe Git worktree cleanup and perform its
worktree-cleanup operation. This is a soft dependency: if unavailable, stop and identify
the missing capability instead of pruning or deleting state directly.
