---
description: Create one explicitly targeted workspace from an exact source without lifecycle assumptions
allowed-tools:
  - Read
  - Bash
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  <target> --source <reference> [--branch <reference>]
  [--capability <selection>] [--provider <selection>] [--dry-run]
---

# /xonovex-workflow:workspace-create — Create Workspace

## Goal

- Create one isolated workspace at an explicit target.
- Preserve the source relationship needed for a later merge.
- Keep workspace management independent from every core operation.

## Arguments

- `target` (required): Exact workspace path or provider-native destination.
- `--source` (required): Exact source branch, revision, or provider-native reference.
- `--branch` (optional): Explicit branch reference when the selected capability uses
  branches.
- `--capability`, `--provider` (optional): Independent workspace method and provider
  selections.
- `--dry-run` (optional): Validate and preview creation without changing state.

## Core Workflow

1. Resolve the exact target, source, and any explicit capability or provider. Infer a
   provider only when repository context makes it unambiguous, and report the inference.
2. Load only the selected or unambiguous workspace capability. If an explicit
   capability is unavailable, name it and stop without improvising.
3. Verify that the source exists and that neither the target nor its provider-native
   reference would be overwritten or checked out elsewhere.
4. Preview the exact resources to create. Create them only within the named target.
5. Record the source relationship in provider-native metadata when supported, then
   return the workspace locator and revision.

## Implementation

For a Git worktree, require an explicit branch, verify that source and branch resolve,
reject an existing target or a branch checked out elsewhere, use `git worktree add`,
and record `branch.<branch>.mergeBackTo`. Do not associate a plan or derive status from
one.

## Error Handling

- Missing or broad target, source, or required branch: stop before creating anything.
- Existing target or branch collision: report the exact conflict without overwriting.
- Ambiguous provider: show the viable providers and request one selection.
- Partial creation: report every created resource and the safe recovery action.
