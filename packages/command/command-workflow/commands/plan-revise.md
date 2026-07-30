---
description: "Draft: apply your annotations and prompt feedback to the plan document (approve separately with plan-accept)"
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[plan-file]"
---

# /xonovex-workflow:plan-revise - Revise Plan from Feedback

> Lifecycle: research → decide → create → **revise** ⇄ critique → accept → subplans-create → continue → update → validate

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**revise** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas. Do not restate them.
