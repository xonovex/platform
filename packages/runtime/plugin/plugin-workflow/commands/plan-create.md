---
description: "Draft: create a high-level plan from research for user review before detailed subplans"
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

# /xonovex-workflow:plan-create - Create Plan with Research

Create one high-level plan from research for review before expanding it into subplans.

## Arguments

- `spec-file-or-requirements` (optional): Path to spec or inline requirements (defaults to conversation context)
- `--interactive` (optional): Ask context-dependent technical questions during research
- `--depends-on <plan>` (optional): Mark dependency on another plan
- `--dry-run` (optional): Preview without writing files

## Delegation

Load the `plan-guide` skill (plugin `xonovex-workflow`) and perform its **create** operation with these arguments. The skill is the source of truth for the procedure, output format, and gotchas. Do not restate them.
