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

Load the `git-guide` skill (plugin `xonovex-skill-git`) and perform its
**worktree-abandon** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas — do not restate them.
