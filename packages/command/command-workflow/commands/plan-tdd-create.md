---
description: >-
  Create a high-level TDD plan with research for user review before detailed
  step generation
allowed-tools:
  - Write
  - Read
  - Glob
  - Grep
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "[spec-file-or-requirements] [--interactive] [--depends-on <plan>] [--dry-run]"
---

# /xonovex-workflow:plan-tdd-create — Create TDD Plan with Research

## Arguments

- `spec-file-or-requirements` (optional): Path to spec or inline requirements (defaults to context)
- `--interactive` (optional): Ask about test approach and step granularity
- `--depends-on <plan>` (optional): Mark dependency on another plan
- `--dry-run` (optional): Preview without writing

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-tdd-create** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas — do not restate them.
