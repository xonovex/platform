---
description: >-
  Split a multi-concern skill into several single-owner composable skills, each
  owning one concept and cross-referencing the others by name
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: "[skill-file] [--into <names>] [--dry-run]"
---

# /xonovex-utility:skill-guide-decompose — Decompose a Skill into Composable Skills

## Arguments

- `skill-file` (required): Path to the SKILL.md or skill directory to decompose.
- `--into <names>` (optional): Comma-separated target skill names to split into (inferred from the concern map if omitted).
- `--dry-run` (optional): Preview the proposed split — new skills, moved references, cross-links — without writing.

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**decompose** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
