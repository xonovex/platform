---
description: Distil this session's lessons straight into the relevant AGENTS.md files (extract + apply in one pass)
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[category] [--from-reflections] [--persist] [--dry-run] [--agents-file <path>]"
---

# /xonovex-utility:reflect-to-instructions — Convert Insights to AGENTS.md

## Arguments

- `category` (optional): Focus on one category (e.g., `testing`, `typescript`). Default: all
  session insights, each routed to the nearest AGENTS.md via its `applies_to`.
- `--from-reflections`: Source insights from existing `reflections/*.md` files instead of
  extracting from the session (the explicit two-step flow after `reflect-extract`).
- `--persist [<dir>]`: Also write the insights as `reflections/*.md` for an audit trail
  (default: off — apply directly without storing).
- `--dry-run`: Preview without modifying.
- `--agents-file <path>`: Target AGENTS.md (default: auto-detect from `applies_to`).

## Delegation

Load the `reflect-guide` skill (plugin `xonovex-skill-reflect`) and perform its
**integrate-instructions** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas — do not restate them.
