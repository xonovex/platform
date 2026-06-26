---
description: Update a plan document with the latest implementation status and test results
allowed-tools:
  - Read
  - Edit
  - Bash
  - Glob
  - TaskUpdate
  - TaskList
  - Skill
argument-hint: "[document-path] [--dry-run]"
---

# /xonovex-workflow:plan-update — Update Plan Progress

## Arguments

`/plan-update [document-path] [--dry-run]`

- `document-path` (optional): Path to plan document (auto-detects from git config if omitted)
- `--dry-run` (optional): Preview changes without modifying files

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-update** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
