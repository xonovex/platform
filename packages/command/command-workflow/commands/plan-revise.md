---
description: "Draft: apply your annotations and prompt feedback to the plan document; --final marks it approved"
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[plan-file] [--final]"
---

# /xonovex-workflow:plan-revise — Revise Plan from Feedback

> Lifecycle: research → decide → create → **revise** ⇄ critique → subplans-create → continue → update → validate

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)
- `--final` (optional): Treat this as the final pass — after resolving feedback, mark plan as `approved` in frontmatter

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-revise** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
