---
description: >-
  Optimize a skill or catalog to its delta over the weakest model through
  baseline, scope, excise, gate, and ablation
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - Task
  - Skill
argument-hint: "[skill-file|--all] [--model <weakest>] [--tier <tier>] [--dry-run] [--report-only]"
---

# /xonovex-utility:skill-optimize — Trim a skill to its knowledge delta and verify

## Arguments

- `[skill-file]` (required unless `--all`) - Path to a SKILL.md or skill directory
- `--all` (optional) - Optimize every skill in the catalog, one optimize per skill in parallel
- `[--model <m>]` (optional) - Weakest model to measure and ablate against (default `haiku`)
- `[--tier <t>]` (optional) - Trim depth; `auto` classifies per skill (default `auto`)
- `[--dry-run]` (optional) - Preview cuts without writing
- `[--report-only]` (optional) - Ablate and report regressions without restoring

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**optimize** operation with these arguments. The skill is the source of truth for the
procedure, tiers, and gotchas — do not restate them.
