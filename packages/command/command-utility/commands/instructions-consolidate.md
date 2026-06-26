---
description: Consolidate project instructions by removing redundant files and standardizing format
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: "[--dry-run] [--path <directory>]"
---

# /xonovex-utility:instructions-consolidate — Consolidate project instruction files

## Arguments

- `--dry-run` (optional): Preview without modifying
- `--path <directory>` (optional): Root directory to scan (defaults to workspace root)

## Delegation

Load the `instruction-guide` skill (plugin `xonovex-skill-instruction`) and perform its
**consolidate** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
