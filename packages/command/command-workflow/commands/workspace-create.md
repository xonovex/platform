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

## Arguments

- `target` (required): Exact workspace path or provider-native destination.
- `--source` (required): Exact source branch, revision, or provider-native reference.
- `--branch` (optional): Explicit branch reference when the selected capability uses
  branches.
- `--capability`, `--provider` (optional): Independent workspace method and provider
  selections.
- `--dry-run` (optional): Validate and preview creation without changing state.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace create** operation with these arguments. The skill is the source of truth
for the procedure, output, and error handling; do not restate them.
