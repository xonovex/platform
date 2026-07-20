---
description: Validate and merge one explicit workspace into one explicit destination
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  <target> --into <destination-reference> [--revision <revision>]
  [--method <selection>] [--capability <selection>] [--provider <selection>]
  [--squash] [--remove] [--confirm] [--dry-run]
---

# /xonovex-workflow:workspace-merge — Merge Workspace

## Goal

- Validate one exact workspace before merging it into one exact destination.
- Bring the source up to date before integration.
- Preserve the workspace whenever validation, rebase, or merge fails.

## Arguments

- `target` (required): Exact workspace path or provider-native source reference.
- `--into` (required): Exact destination workspace, branch, or native reference.
- `--revision` (optional): Exact source revision to merge.
- `--method`, `--capability`, `--provider` (optional): Independent merge,
  workspace, and provider selections.
- `--squash` (optional): Request one consolidated change when supported.
- `--remove` (optional): Remove the source workspace only after a successful merge.
- `--confirm` (optional): Explicitly authorize the described merge effect.
- `--dry-run` (optional): Validate and preview without merging.

## Core Workflow

1. Resolve the exact target, destination, revision, and selections. Report an inferred
   provider only when it is unambiguous; stop when more than one provider fits.
2. Load the selected workspace, validation, merge, and provider capabilities. Name and
   stop on any unavailable explicit capability.
3. Verify source/destination identity, source metadata, and a clean source workspace.
   Run the repository's typecheck, lint, build, and tests before merging.
4. Update the source against the current destination using the selected capability,
   resolve no conflicts by guesswork, and preview the final merge.
5. Merge only when `--confirm` or the original request already authorizes the exact
   source, destination, and effect. Remove the workspace only after success and only
   when `--remove` was explicit.

## Implementation

For a Git worktree, verify the worktree and branch relationship, require
`branch.<branch>.mergeBackTo`, reject a mismatch with `--into`, fetch, rebase the
feature branch onto the current destination, then merge. Never remove the worktree when
validation, rebase, or merge fails.

## Error Handling

- Dirty, unvalidated, stale, or mismatched source: stop and report the exact condition.
- Conflict: leave the workspace intact and identify the unresolved files.
- Missing authorization: return the dry-run preview and request confirmation.
- Failed merge: do not remove the workspace or its provider-native reference.
