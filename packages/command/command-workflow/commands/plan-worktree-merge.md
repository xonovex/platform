---
description: Merge feature worktree back to source worktree
allowed-tools:
  - Bash
  - Skill
argument-hint: "[--squash] [--remove-worktree] [--delete-remote] [--dry-run]"
---

# /xonovex-workflow:plan-worktree-merge — Merge Feature Worktree Back to Source

## Arguments

`/plan-worktree-merge [--squash] [--remove-worktree] [--delete-remote] [--dry-run]`

- `--squash`: Squash all feature commits into a single commit
- `--remove-worktree`: Remove feature worktree directory after merge (default: keep)
- `--delete-remote`: Also delete remote feature branch (requires remote exists)
- `--dry-run` (optional): Preview changes without modifying files

## Delegation

Load an installed capability that provides safe Git worktree merge and perform its
worktree-merge operation. This is a soft dependency: if unavailable, stop and identify
the missing capability instead of improvising destructive Git operations.
