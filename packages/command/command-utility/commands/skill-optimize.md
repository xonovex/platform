---
description: Optimize a skill (or the whole catalog) to its delta over the weakest model — baseline-measure, tier-aware trim, validate.py gate, and ablation-restore of any essential fact lost
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
argument-hint: "[skill-file|--all] [--model <weakest>] [--tier auto|aggressive|moderate|conservative] [--dry-run] [--report-only]"
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

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and run its **optimize**
operation end to end for the target(s): measure the weakest model's baseline against each
skill's `evals.json` with the skill absent, classify tier, trim to the knowledge delta,
gate with `scripts/validate.py`, then ablate — re-measure with the trimmed skill in context
and restore any essential fact it no longer conveys. The skill is the source of truth for
the procedure, tiers, and gotchas — do not restate them.
