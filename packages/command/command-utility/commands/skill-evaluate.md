---
description: Author or refresh a skill's evals.json output-eval seed, knowledge probes (prompt + binary assertions) with a tier field
allowed-tools:
  - Read
  - Write
  - Glob
  - Grep
  - Bash
  - TodoWrite
  - Skill
argument-hint: "[skill-file] [--count <n>]"
---

# /xonovex-utility:skill-evaluate - Seed a skill's output-eval file

## Arguments

- `[skill-file]` (required) - Path to SKILL.md or skill directory
- `[--count <n>]` (optional) - Target number of evals (default 2-4)

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and perform its
**evaluating-outputs** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas. Do not restate them.
Prefer an installed test-design skill when its routing description fits the requested
evaluation work; continue with baseline evaluation design when none is installed.
