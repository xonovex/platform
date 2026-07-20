---
description: Abandon one explicit workspace while preserving its reason and recoverable state
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  <target> --reason <text> [--revision <revision>]
  [--capability <selection>] [--provider <selection>]
  [--result <destination-reference>] [--snapshot] [--remove]
  [--delete-reference] [--confirm] [--dry-run]
---

# /xonovex-workflow:workspace-abandon — Abandon Workspace

## Goal

- Stop using one explicit workspace and record why.
- Preserve partial work and recovery information by default.
- Remove workspace or branch state only when explicitly requested.

## Arguments

- `target` (required): Exact workspace path or provider-native reference.
- `--reason` (required): Present-tense reason for abandonment.
- `--revision` (optional): Exact workspace revision being abandoned.
- `--capability`, `--provider` (optional): Independent workspace and provider
  selections.
- `--result` (optional): Explicit destination for the abandonment record.
- `--snapshot` (optional): Preserve current work using the selected capability before
  removal.
- `--remove` (optional): Remove the workspace after recording its state.
- `--delete-reference` (optional): Also delete its provider-native branch or reference.
- `--confirm` (optional): Explicitly authorize the described removals.
- `--dry-run` (optional): Preview updates and removals without applying them.

## Core Workflow

1. Resolve the exact target, revision, and selections. Report an inferred provider only
   when unambiguous; stop on ambiguity.
2. Load the selected workspace and provider capabilities. Name and stop on an
   unavailable explicit capability.
3. Inspect uncommitted or otherwise unsaved state and report how it can be recovered.
   Create a snapshot only when `--snapshot` was selected.
4. Produce the abandonment record with reason, partial state, revision, and recovery
   locator. Persist it only to an explicit `--result`.
5. Keep the workspace by default. Apply `--remove` or `--delete-reference` only after
   exact-scope confirmation; never broaden the target.

## Implementation

For a Git worktree, use Git's worktree operations rather than deleting its directory.
Do not force-remove dirty state. Delete a branch reference only when explicitly selected
and preserve a snapshot or recovery revision first when requested. Do not read or update
a plan.

## Error Handling

- Missing target or reason: stop and request the exact value.
- Dirty state with removal requested but no recoverable snapshot: preserve the workspace.
- Missing authorization: return a dry-run preview and request confirmation.
- Partial removal: report remaining path, metadata, and reference state.
