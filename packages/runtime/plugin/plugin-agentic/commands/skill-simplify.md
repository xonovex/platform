---
description: >-
  Make skills project-independent, remove redundancy, condense SKILL.md to bullet list with examples in reference files
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - TodoWrite
  - Skill
argument-hint: "[skill-file] [--dry-run] [--target-reduction <percent>]"
---

# /xonovex-agentic:skill-simplify - Condense verbose skill files

Remove redundant guidance and move detailed examples into references while preserving the skill behavior.

## Arguments

- `[skill-file]` (required): Path to SKILL.md file or skill directory
- `[--dry-run]` (optional): Preview without modifying
- `[--target-reduction <percent>]` (optional): Override default 70% (range: 50-90)

## Delegation

Load the `skill-guide` skill (plugin `xonovex-agentic`) and perform its **simplify** operation with these arguments. The skill is the source of truth for the procedure, output format, and gotchas. Do not restate them.
