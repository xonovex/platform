---
description: >-
  Augment an existing slash command with elements from another slash command while preserving structure and style
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: >-
  [target-command] [source-command] [--aspects <aspects>] [--percentage <percent>] [--interactive] [--dry-run]
---

# /xonovex-agentic:slashcommand-assimilate - Augment Slash Command

Add useful guidance from another command while preserving the target argument contract and delegation.

## Arguments

- `target-command` (required): Target command file (augmented)
- `source-command` (required): Source command file (provides elements)
- `--aspects <aspects>` (optional): Focus aspects (e.g., "workflow,validation,error-handling")
- `--percentage <percent>` (optional): Intensity 10-100 (default: 50)
- `--interactive` (optional): Ask clarifying questions
- `--dry-run` (optional): Preview without modifying

## Delegation

Load the `command-guide` skill (plugin `xonovex-agentic`) and perform its **merge** operation with these arguments. The skill is the source of truth for the procedure, output format, and gotchas. Do not restate them.
