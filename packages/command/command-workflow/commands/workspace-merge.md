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
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--squash] [--remove] [--confirm] [--dry-run]
---

# /xonovex-workflow:workspace-merge — Merge Workspace

## Arguments

- `target` (required): Exact workspace path or provider-native source reference.
- `--into` (required): Exact destination workspace, branch, or native reference.
- `--revision` (optional): Exact source revision to merge.
- `--criteria` (repeatable, optional): Explicit conditions the workspace must satisfy
  before the merge.
- `--method` (optional): Validation or merge procedure selection.
- `--capability` (repeatable, optional): Independent workspace, validation, and merge
  capability selections.
- `--provider` (optional): Provider selection for native references and effects.
- `--squash` (optional): Request one consolidated change when supported.
- `--remove` (optional): Remove the source workspace only after a successful merge.
- `--confirm` (optional): Explicitly authorize the described merge effect.
- `--dry-run` (optional): Validate and preview without merging.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace merge** operation with these arguments. The skill is the source of truth
for the procedure, output, and error handling; do not restate them.
