---
description: Verify a skill's trims kept essential knowledge — ablate the removed content against the weakest model and restore anything it fails without
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - Skill
argument-hint: "[skill-file] [--model <weakest>] [--base <git-ref>] [--report-only]"
---

# /xonovex-utility:skill-ablate — Verify no essential knowledge was trimmed

## Arguments

- `[skill-file]` (required) - Path to SKILL.md or skill directory
- `[--model <m>]` (optional) - Model to ablate against; use the weakest you deploy (default `haiku`)
- `[--base <git-ref>]` (optional) - Pre-trim ref to diff removed content against (default `HEAD~1`)
- `[--report-only]` (optional) - Report regressions without restoring

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform the **optimize**
verification (ablation): diff removed content against the base ref, then test with-skill
vs without-skill on the given model via `evals.json` and `scripts/eval-outputs.py` — an
eval the model fails **with** the skill is a fact a trim removed — and restore anything
the current skill no longer conveys. The skill is the source of truth for the procedure —
do not restate it.
