---
description: Convert insights from a category into a progressive disclosure skill
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[category] [--dry-run] [--force] [--output <path>]"
---

# /xonovex-utility:insights-skills-integrate — Convert Insights to Skill

## Arguments

- `category` (required) - Category to convert (e.g., `hono`, `typescript`, `workflow`)
- `--dry-run` - Preview without writing
- `--force` - Overwrite existing skill instead of merging
- `--output <path>` - Custom output path (default: `.claude/skills/{category}/SKILL.md`)

## Delegation

Load the `insights-guide` skill (plugin `xonovex-skill-insights`) and perform its
**integrate-skills** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
