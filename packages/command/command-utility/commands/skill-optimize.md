---
description: Optimize a skill (or the whole catalog) to its delta over the weakest model in five phases — Baseline, Scope, Excise, Gate, Ablate: measure priors, set trim depth, cut to the delta, validate.py, then ablate-restore any essential fact lost
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
operation end to end for the target(s), in five phases:

1. **Baseline** — measure the weakest model against each skill's `evals.json` with the skill absent, to map what it already knows.
2. **Scope** — classify the trim depth (aggressive/moderate/conservative) for the skill's type.
3. **Excise** — cut to the knowledge delta; dedupe; fix any defect the cut surfaces.
4. **Gate** — `scripts/validate.py` must pass.
5. **Ablate** — re-measure with the trimmed skill in context and restore any essential fact it no longer conveys.

The skill is the source of truth for the procedure, tiers, and gotchas — do not restate them.
