---
description: "Execution: expand an approved parent plan into detailed subplans with parallel-group detection"
allowed-tools:
  - Write
  - Read
  - Glob
  - Grep
  - TaskCreate
  - TaskUpdate
  - Skill
argument-hint: "[parent-plan-file] [--by-phase] [--dry-run]"
---

# /xonovex-workflow:plan-subplans-create - Generate Detailed Subplans from Parent Plan

> Lifecycle: research → decide → create → revise ⇄ critique → accept → **subplans-create** → continue → update → validate

## Arguments

- `parent-plan-file` (required): Path to approved parent plan (e.g., `plans/auth.md`)
- `--by-phase` (optional): Split by phase markers instead of logical grouping
- `--dry-run` (optional): Preview without writing files

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**expand** operation with these arguments. The skill is the source of
truth for the procedure, output format, and gotchas. Do not restate them.
