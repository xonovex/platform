---
description: Convert insights from a category into AGENTS.md bullet points
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[category] [--dry-run] [--agents-file <path>]"
---

# /xonovex-utility:reflect-to-instructions — Convert Insights to AGENTS.md

## Arguments

- `category` (required): Category to convert (e.g., `testing`, `typescript`, `workflow`)
- `--dry-run` (optional): Preview without modifying
- `--agents-file <path>` (optional): Target AGENTS.md (default: auto-detect from `applies_to`)

## Delegation

Load the `reflect-guide` skill (plugin `xonovex-skill-reflect`) and perform its
**integrate-instructions** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas — do not restate them.
