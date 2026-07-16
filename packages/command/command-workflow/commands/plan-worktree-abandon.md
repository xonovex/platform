---
description: Document and abandon a feature with reason and learnings
allowed-tools:
  - Bash
  - Read
  - Edit
  - Write
  - Skill
argument-hint: "[reason] [--remove-worktree] [--no-plan] [--commit] [--dry-run]"
---

# /xonovex-workflow:plan-worktree-abandon — Abandon Feature with Documentation

## Arguments

- `reason` (optional): Concise reason for abandonment (prompted if not provided)
- `--remove-worktree` (optional): Remove feature worktree after documenting
- `--no-plan` (optional): Skip plan update (for features without plans)
- `--commit` (optional): Commit current state before abandoning
- `--dry-run` (optional): Preview changes without modifying files

## Delegation

Load an installed capability that provides safe Git worktree abandonment and perform its
worktree-abandon operation. This is a soft dependency: if unavailable, stop and identify
the missing capability instead of deleting or rewriting state directly.
