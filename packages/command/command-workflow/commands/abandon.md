---
description: Stop work on one subject while preserving its reason, partial result, and retry boundary
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> --reason <text> [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>] [--cleanup <selection>] [--dry-run]
---

# /xonovex-workflow:abandon — Abandon

## Arguments

- `subject` (required): Inline work description or one opaque provider-native
  reference.
- `--reason` (required): Present-tense reason for stopping.
- `--reference` (repeatable): Partial results, evidence, or opaque references.
- `--revision` (optional): Exact native revision being abandoned.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Retention or cleanup constraints.
- `--result` (optional): Explicit destination for the abandonment record.
- `--cleanup` (optional): Explicit cleanup capability or scope; preserve by default.
- `--dry-run` (optional): Preview provider updates or cleanup.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Abandon** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
