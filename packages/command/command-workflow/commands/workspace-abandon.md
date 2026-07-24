---
description: Record why one workspace was abandoned while preserving its resources
allowed-tools:
  - Read
  - Glob
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target] [--request <file>] [--target-revision <revision>]
  [--reason <text>]
---

# /xonovex-workflow:workspace-abandon — Abandon Workspace

## Arguments

- `target` (required unless `--request` supplies it): Exact workspace path or opaque
  native reference.
- `--request` (optional): Markdown workflow handoff containing the workspace, reason,
  partial state, and equivalent inputs. Do not combine it with shorthand arguments.
- `--target-revision` (optional): Exact workspace revision when available.
- `--reason` (required unless `--request` supplies it): Present-tense reason for
  stopping.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace abandon** operation with these arguments. Return abandonment and recovery
information without removing or mutating workspace resources.
