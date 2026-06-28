---
description: Distil this session's lessons straight into the relevant guideline skills (extract + apply in one pass)
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[category] [--from-reflections] [--persist] [--dry-run] [--force] [--output <path>]"
---

# /xonovex-utility:reflect-to-skill — Convert Insights to Skill

## Arguments

- `category` (optional) - Focus on one category (e.g., `testing`, `typescript`). Default: all
  session insights, each routed to the existing skill that owns its domain.
- `--from-reflections` - Source insights from existing `reflections/*.md` files instead of
  extracting from the session (the explicit two-step flow after `reflect-extract`).
- `--persist [<dir>]` - Also write the insights as `reflections/*.md` for an audit trail
  (default: off — apply directly without storing).
- `--dry-run` - Preview without writing.
- `--force` - Overwrite an existing skill instead of merging.
- `--output <path>` - Custom output path for a newly created skill.

## Delegation

Load the `reflect-guide` skill (plugin `xonovex-skill-reflect`) and perform its
**integrate-skills** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
