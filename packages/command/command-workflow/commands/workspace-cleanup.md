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

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace cleanup** operation with these arguments. The skill is the source of truth
for the procedure, output, and error handling; do not restate them.
