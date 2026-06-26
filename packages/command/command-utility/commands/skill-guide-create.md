---
description: Create a guideline skill from a provided document or URL
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - AskUserQuestion
  - TodoWrite
  - Skill
argument-hint: "[source] [--name <name>] [--dry-run]"
---

# /xonovex-utility:skill-guide-create — Create Guideline Skill from Document

## Arguments

- `source` (required): URL or file path to the source document
- `--name` (required): Skill name in kebab-case (e.g., `react-guide`, `go-guide`)
- `--dry-run` (optional): Preview generated structure without writing files

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**create** operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
