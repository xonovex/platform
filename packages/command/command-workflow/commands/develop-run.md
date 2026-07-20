---
description: Execute exact Planning assignments with the least-adaptive suitable executors and publish independent Development results
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
argument-hint: "<planning-reference...> [--revision <native-revision>] [--profile <reference>] [--workspace <selection>] [--executor <class>] [--max-concurrency <count>]"
---

# /xonovex-workflow:develop-run — Run Development

## Arguments

- `planning-reference...` (required): One or more opaque child Planning references.
- `--revision` (repeatable): Exact native Planning revision corresponding to each reference.
- `--profile`, `--workspace` (optional): Independent profile or workspace-provider selections.
- `--executor` (optional): Explicit `deterministic`, `model`, `agent`, `human`, or `external` workflow selection; it does not imply controls or maturity.
- `--max-concurrency` (optional): Bound ready assignments; the profile/provider limit remains authoritative when lower.

## Delegation

Load `workflow-guide` (plugin `xonovex-skill-workflow`) and perform **develop-run**.
Load each Planning result's implementation skills before mutation. Use workspace, harness,
Git, model, agent, and provider skills only when selected; publish one exact-revision
Development result per assignment and preserve partial failures.
