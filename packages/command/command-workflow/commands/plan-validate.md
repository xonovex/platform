---
description: Verify that a plan or current work has been fully achieved
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Task
  - TaskCreate
  - TaskUpdate
  - TaskList
  - Skill
argument-hint: "[plan-file] [--detailed]"
---

# /xonovex-workflow:plan-validate — Validate Plan Achievement

## Arguments

`/plan-validate [plan-file] [--detailed]`

- `plan-file` (optional): Path to plan document (if omitted, validates current conversation goal)
- `--detailed` (optional): Comprehensive analysis with full evidence trail

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-validate** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
