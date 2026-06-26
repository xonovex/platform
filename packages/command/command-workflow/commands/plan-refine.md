---
description: >-
  Process user annotations in a plan document and refine iteratively until
  approved
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[plan-file] [--final]"
---

# /xonovex-workflow:plan-refine — Refine Plan from Annotations

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)
- `--final` (optional): Treat this as the final pass — after resolving annotations, mark plan as `approved` in frontmatter

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-refine** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
