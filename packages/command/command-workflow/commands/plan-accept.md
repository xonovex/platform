---
description: "Draft: approve a plan for execution, setting status approved after a final sanity check"
allowed-tools:
  - Read
  - Edit
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: "[plan-file]"
---

# /xonovex-workflow:plan-accept - Approve a Plan

> Lifecycle: research → decide → create → revise ⇄ critique → **accept** → subplans-create → continue → update → validate

## Arguments

- `plan-file` (optional): Path to plan document (auto-detects from git config or most recent plan in `plans/`)

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**accept** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas. Do not restate them.
