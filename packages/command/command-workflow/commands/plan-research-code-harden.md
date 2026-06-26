---
description: >-
  Research code-hardening opportunities (type safety, validation, logging,
  error handling) — read-only report that feeds a follow-up plan; makes no
  changes
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "[path] [--aspects <type-safety,logging,validation>]"
---

# /xonovex-workflow:plan-research-code-harden — Research Code Hardening Opportunities

## Arguments

- `path` (required): Directory to analyze
- `--aspects` (optional): Comma-separated aspects (type-safety, logging, validation, error-handling, testing, or custom)

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-research-code-harden** operation with these arguments. The skill is the source
of truth for the procedure, output format, and gotchas — do not restate them.
