---
description: >-
  Augment project instructions with elements from another project's instructions
  while preserving structure and style
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target-instructions] [source-instructions] [--aspects <aspects>]
  [--percentage <percent>] [--interactive] [--dry-run]
---

# /xonovex-utility:instructions-assimilate — Augment Project Instructions

## Arguments

- `target-instructions` (required): Target AGENTS.md (augmented)
- `source-instructions` (required): Source AGENTS.md (provides patterns)
- `--aspects <aspects>` (optional): Focus aspects (e.g., "workflow,structure,integration")
- `--percentage <percent>` (optional): Intensity 10-100 (default: 45)
- `--interactive` (optional): Ask clarifying questions
- `--dry-run` (optional): Preview without modifying

## Delegation

Load the `instruction-guide` skill (plugin `xonovex-skill-instruction`) and perform its
**merge** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
