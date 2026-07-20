---
description: Remove only explicitly selected stale or merged workspaces after a reviewable preview
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  <target>... [--capability <selection>] [--provider <selection>]
  [--remove-reference] [--prune] [--force] [--confirm] [--dry-run]
---

# /xonovex-workflow:workspace-cleanup — Clean Up Workspaces

## Goal

- Inspect and clean only explicitly named workspace targets.
- Distinguish merged, stale, active, and dirty state before removal.
- Make every destructive effect previewable and confirmable.

## Arguments

- `target` (repeatable, required): Exact workspace path or provider-native reference.
- `--capability`, `--provider` (optional): Independent cleanup and provider
  selections.
- `--remove-reference` (optional): Remove the associated branch or native reference
  after the workspace.
- `--prune` (optional): Prune stale provider administration metadata after listing
  every affected record.
- `--force` (optional): Include an exact dirty or unmerged target in the preview.
- `--confirm` (optional): Explicitly authorize the previewed cleanup set.
- `--dry-run` (optional): Inventory and preview only.

## Core Workflow

1. Resolve every exact target and selection. Never discover additional removal targets
   implicitly. Report an inferred provider only when unambiguous and stop on ambiguity.
2. Load the selected workspace and provider capabilities. Name and stop on an
   unavailable explicit capability.
3. Classify each target as merged, stale, active, dirty, or unknown and show its
   recovery revision. Keep active, dirty, unmerged, and unknown targets by default.
4. Preview the exact workspace, reference, and administration records that would be
   removed. Broad patterns, unresolved variables, and implicit home or repository roots
   are invalid targets.
5. Remove only the confirmed set. Return what was removed, retained, recoverable, and
   not recoverable.

## Implementation

For Git worktrees, inventory with `git worktree list`, remove with
`git worktree remove`, use `git branch -d` for merged references, and reserve forced
removal or `git branch -D` for an explicitly confirmed exact target. Run
`git worktree prune` only after listing the stale metadata it affects.

## Error Handling

- Missing, broad, or unresolved target: stop without removing anything.
- Dirty or unmerged target without `--force`: retain it and explain why.
- Missing authorization: return the dry-run preview and request confirmation.
- Partial cleanup: report the remaining workspace, reference, and metadata separately.
