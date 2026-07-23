---
description: Author or refresh a skill's evals.json output-eval seed — knowledge probes (prompt + binary assertions) with a tier field
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

# /xonovex-utility:skill-evaluate — Seed a skill's output-eval file

## Arguments

- `[skill-file]` (required) - Path to SKILL.md or skill directory
- `[--count <n>]` (optional) - Target number of evals (default 2-4)

## Delegation

Load the `skill-guide` skill (plugin `xonovex-skill-skill`) and author the skill's
`evals.json` output-eval seed following its **evaluating-outputs** procedure — knowledge
probes with binary assertions, plus a `tier` field, stored at the guide root. The skill
is the source of truth for the shape and gotchas — do not restate them.
