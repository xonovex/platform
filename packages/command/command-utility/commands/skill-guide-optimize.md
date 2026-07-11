---
description: Trim a skill to its delta over baseline model knowledge — tier-aware cuts, intra-skill dedup, defect fixes, validate.py gate
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - Skill
argument-hint: "[skill-file] [--tier auto|aggressive|moderate|conservative] [--dry-run]"
---

# /xonovex-utility:skill-guide-optimize — Trim a skill to its knowledge delta

## Arguments

- `[skill-file]` (required) - Path to SKILL.md or skill directory
- `[--tier <t>]` (optional) - Trim depth; `auto` classifies the skill (default `auto`)
- `[--dry-run]` (optional) - Preview cuts without writing

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**optimize** operation with these arguments. The skill is the source of truth for the
procedure, tiers, and gotchas — do not restate them.
