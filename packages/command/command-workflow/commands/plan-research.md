---
description: "Specialized codebase/web research or read-only code-quality analysis; use research-run for canonical reusable Research results"
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

# /xonovex-workflow:plan-research — Research Codebase and Web for Planning

> Specialized analysis: use `research-run` when the output must be a provider-native canonical Research result.

## Arguments

- `requirements` (required): Description of what to research
- `--interactive` (optional): Ask clarifying questions
- `--save-to <file>` (optional): Save research to file

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-research** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
