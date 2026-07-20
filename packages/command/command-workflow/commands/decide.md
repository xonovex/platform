---
description: Record one descriptive outcome and rationale without granting authority or changing a gate
allowed-tools:
  - Read
  - Write
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  <subject> [--outcome <text>] [--reference <reference>...]
  [--revision <revision>] [--kind <selection>] [--perspective <selection>]
  [--criteria <criteria>...] [--method <selection>]
  [--capability <selection>...] [--provider <selection>]
  [--result <destination-reference>]
---

# /xonovex-workflow:decide — Decide

## Arguments

- `subject` (required): Inline decision question or one opaque provider-native
  reference.
- `--outcome` (optional): Explicit outcome to record; otherwise derive a recommendation
  from the supplied evidence and criteria.
- `--reference` (repeatable): Evidence, options, constraints, or opaque references.
- `--revision` (optional): Exact native revision of a referenced subject.
- `--kind`, `--perspective`, `--method`, `--capability`, `--provider`
  (optional): Independent, open selections.
- `--criteria` (repeatable, optional): Explicit decision criteria.
- `--result` (optional): Explicit provider-native destination reference.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Decide** operation with these arguments. The skill is the source of truth for the
procedure, output, and error handling; do not restate them.
