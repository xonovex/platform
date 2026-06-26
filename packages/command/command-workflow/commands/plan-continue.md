---
description: Resume work from an existing plan document with full context loading
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "[document-path]"
---

# /xonovex-workflow:plan-continue — Continue Progress from Plan

## Arguments

`/plan-continue [document-path]`

- `document-path` (optional): Path to plan document. Resolution order if omitted:
  1. Check git config for associated plan (feature worktree)
  2. **Use current conversation context** - if a `/xonovex-workflow:plan-create`, `/xonovex-workflow:plan-research`, or similar command was run earlier in this conversation, continue from that research/plan. Do NOT search for other plans.
  3. Only if no context exists: Ask user which plan to continue

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-continue** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
