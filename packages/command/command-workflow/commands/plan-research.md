---
description: "Pre-plan: research codebase and web for requirements without creating a plan"
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - Task
  - TaskCreate
  - TaskUpdate
  - WebSearch
  - WebFetch
  - AskUserQuestion
  - Skill
argument-hint: "<requirements> [--interactive] [--save-to <file>]"
---

# /xonovex-workflow:plan-research — Research Codebase and Web

> Lifecycle: **research** → decide → create → revise ⇄ critique → subplans-create → continue → update → validate

## Arguments

- `requirements` (required): Description of what to research
- `--interactive` (optional): Ask clarifying questions
- `--save-to <file>` (optional): Save research to file

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-research** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
