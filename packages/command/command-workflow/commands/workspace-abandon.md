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
  [--remove-reference] [--confirm] [--dry-run]
---

# /xonovex-workflow:workspace-abandon — Abandon Workspace

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
- `--remove-reference` (optional): Also remove its provider-native branch or reference.
- `--confirm` (optional): Explicitly authorize the described removals.
- `--dry-run` (optional): Preview updates and removals without applying them.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace abandon** operation with these arguments. The skill is the source of truth
for the procedure, output, and error handling; do not restate them.
