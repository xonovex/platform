---
description: >-
  Execution: work a roadmap as supervisor — brief an implementation agent per
  item, verify its work independently, then record and commit it
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
argument-hint: >-
  <roadmap-path> [--model <model>] [--order <sequential|parallel>]
  [--orchestration <agents|workflow>] [--stop-after <count-or-plan>]
  [--commit <per-plan|batch>]
---

# /xonovex-workflow:plan-delegate — Supervise Roadmap Execution by Delegation

> Lifecycle: research → decide → create → revise ⇄ critique → accept → subplans-create → continue / **delegate** → update → validate

## Arguments

- `roadmap-path` (required): Path to the roadmap, plan, or explicit subplan list to work through.
- `--model <model>` (optional): Model for the implementation agents (defaults to the strongest implementation model available; the supervisor keeps its own).
- `--order <sequential|parallel>` (optional): Item ordering (defaults to `sequential`; `parallel` runs a `parallel_group` concurrently when its file sets are disjoint).
- `--orchestration <agents|workflow>` (optional): Delegation mechanism (defaults to `agents`; `workflow` uses a scripted multi-agent orchestration).
- `--stop-after <count-or-plan>` (optional): Stop after N items or after a named plan (defaults to running until the ordering is exhausted or something blocks).
- `--commit <per-plan|batch>` (optional): Commit one conventional commit per accepted item, or batch them at the end (defaults to `per-plan`).

## Delegation

Load the `plan-guide` skill (plugin `xonovex-skill-plan`) and perform its
**plan-delegate** operation with these arguments. The skill is the source of truth for
the procedure, output format, and gotchas — do not restate them.
