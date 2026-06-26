---
description: >-
  Create or update a skill by extracting patterns from codebase and project
  instructions
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - AskUserQuestion
  - Task
  - Skill
argument-hint: "[skill-name] [source-path] [--update] [--interactive] [--dry-run]"
---

# /xonovex-utility:skill-guide-extract — Extract Skill from Codebase

## Arguments

- `skill-name` (required): Name for skill (e.g., `example-guide`)
- `source-path` (required): Path to analyze (e.g., `packages/example`)
- `--update` (optional): Update existing skill instead of creating new
- `--interactive` (optional): Ask which patterns to include before writing
- `--dry-run` (optional): Preview without writing files

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**extract-from-codebase** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas — do not restate them.
