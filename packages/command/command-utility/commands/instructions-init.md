---
description: Create an AGENTS.md file for a directory by analyzing its structure and contents
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
  - Skill
argument-hint: "[directory] [--dry-run] [--recursive]"
---

# /xonovex-utility:instructions-init — Create AGENTS.md

## Arguments

- `directory` (required): Target directory
- `--dry-run` (optional): Preview without writing
- `--recursive` (optional): Also create AGENTS.md for subdirectories with unique content

## Delegation

Load the `instruction-guide` skill (plugin `xonovex-skill-instruction`) and perform its
**init** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
