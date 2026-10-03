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

# /xonovex-agentic:instructions-init - Create AGENTS.md

Create AGENTS.md from the directory structure, tools, and project conventions.

## Arguments

- `directory` (required): Target directory
- `--dry-run` (optional): Preview without writing
- `--recursive` (optional): Also create AGENTS.md for subdirectories with unique content

## Delegation

Load the `instruction-guide` skill (plugin `xonovex-agentic`) and perform its **init** operation with these arguments. The skill is the source of truth for the procedure, output format, and gotchas. Do not restate them.
