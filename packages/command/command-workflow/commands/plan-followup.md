---
description: "Execution: close out a completed, paused, or handed-over plan with an inline follow-up record"
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Task
  - Skill
argument-hint: "[plan-file]"
---

# /xonovex-workflow:plan-followup - Close Out a Plan

> Lifecycle: research → decide → create → revise ⇄ critique → accept → subplans-create → continue → update → validate → **followup**

## Arguments

`/plan-followup [plan-file]`

- `plan-file` (optional): Path to the plan document (auto-detects from git config if omitted)

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**followup** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas. Do not restate them. The record
is returned inline; hand plan seeds to `/plan-create` and other sections to the
caller's publish step if persistence is wanted.
