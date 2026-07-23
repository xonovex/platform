---
description: Record why one workspace was abandoned while preserving every workspace resource
allowed-tools:
  - Read
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target] [--request <file>] [--reason <text>]
  [--workspace-provider <provider>] [--workspace-revision <revision>]
---

# /xonovex-workflow:workspace-abandon — Abandon Workspace

## Arguments

- `target` and `--reason` (required unless `--request` supplies them): Exact workspace
  locator and present-tense reason for stopping.
- `--request` (optional): Structured request file for the workspace, partial state,
  evidence, and recovery information. Do not combine it with the shorthands.
- `--workspace-provider`, `--workspace-revision` (optional): Provider and exact
  revision for the simple workspace shorthand.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace abandon** operation with these arguments. Return an inline abandonment
and recovery record. The skill is the source of truth; do not snapshot, remove, prune,
or mutate the workspace or its provider-native reference.
