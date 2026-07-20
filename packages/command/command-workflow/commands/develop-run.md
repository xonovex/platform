---
description: Execute exact Planning assignments and publish independent Development results
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - Task
  - TaskCreate
  - TaskUpdate
  - AskUserQuestion
  - Skill
argument-hint: "<planning-reference...> [--revision <native-revision>] [--profile <reference>] [--workspace <selection>] [--max-concurrency <count>]"
---

# /xonovex-workflow:develop-run — Run Development

## Arguments

- `planning-reference...` (required): One or more opaque child Planning references.
- `--revision` (repeatable): Exact native Planning revision corresponding to each reference.
- `--profile`, `--workspace` (optional): Profile or workspace-provider selections.
- `--max-concurrency` (optional): Bound ready assignments; the profile/provider limit remains authoritative when lower.

## Delegation

Perform **develop-run** with each Planning result's selected implementation skills. Use
workspace, harness, Git, model, agent, and provider skills only when selected, and
preserve partial failures.
