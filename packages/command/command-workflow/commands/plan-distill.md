---
description: "Post-lifecycle: distill a completed plan, branch, or delegated run into a replayable skill suite"
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Write
  - Edit
  - Task
  - Skill
argument-hint: "[plan-source] [skills-dir]"
---

# /xonovex-workflow:plan-distill - Distill Completed Work Into Skills

> Lifecycle: research → decide → create → ... → validate → followup → **distill** (post-lifecycle; consumes the finished work and its records)

## Arguments

`/plan-distill [plan-source] [skills-dir]`

- `plan-source` (optional): The completed plan document or branch to distill (auto-detects from git config if omitted)
- `skills-dir` (optional): Target skills directory for the suite (defaults to the project's skills directory)

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**distill** operation with these arguments. The skill is the source of truth
for the procedure, output format, and gotchas. Do not restate them.
