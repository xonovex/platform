---
description: >-
  Augment an existing skill with elements from another skill while preserving
  structure and style
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target-skill] [source-skill] [--aspects <aspects>] [--percentage <percent>]
  [--interactive] [--dry-run]
---

# /xonovex-utility:skill-guide-assimilate — Augment Skill with Another Skill

## Arguments

- `target-skill` (required): Target skill file/directory (augmented)
- `source-skill` (required): Source skill file/directory (provides elements)
- `--aspects <aspects>` (optional): Focus aspects (e.g., "error-handling,validation")
- `--percentage <percent>` (optional): Intensity 10-100 (default: 50)
- `--interactive` (optional): Ask clarifying questions
- `--dry-run` (optional): Preview without modifying

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**merge** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
