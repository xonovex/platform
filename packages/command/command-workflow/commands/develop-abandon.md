---
description: Stop one Development assignment while preserving its reason, partial result, evidence, cleanup state, and retry boundary
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - Task
  - AskUserQuestion
  - Skill
argument-hint: "<development-or-planning-reference> --reason <text> [--revision <native-revision>] [--cleanup <none|safe|selected>]"
---

# /xonovex-workflow:develop-abandon — Abandon Development

## Arguments

- `development-or-planning-reference` (required): Opaque assignment/result reference.
- `--reason` (required): Present-tense reason for stopping this assignment revision.
- `--revision` (optional): Exact native revision to abandon.
- `--cleanup` (optional): Preserve by default; apply only explicitly selected provider-safe cleanup.

## Delegation

Load `workflow-guide` and perform **develop-abandon**. Load the selected workspace or Git
skill for cancellation and cleanup mechanics. Publish an abandoned Development result;
do not erase partial work or silently reuse its identity for a later attempt.
