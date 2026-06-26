---
description: >-
  Research alignment between two similar code implementations — read-only
  report that feeds a follow-up plan; makes no changes
allowed-tools:
  - Read
  - Glob
  - Grep
  - Bash
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "[file1] [file2] [--interactive]"
---

# /xonovex-workflow:plan-research-code-align — Research Code Alignment Between Similar Implementations

## Arguments

- `file1` (required): Path to first file, or glob pattern
- `file2` (optional): Path to second file. If omitted and file1 is a glob, compares first two matches
- `--interactive` (optional): Ask clarifying questions about reference implementation

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-research-code-align** operation with these arguments. The skill is the source of
truth for the procedure, output format, and gotchas — do not restate them.
