---
description: Stop work and return the reason, partial state, and retry boundary
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [reason] [--request <file>] [--context <context>...]
---

# /xonovex-workflow:abandon — Abandon

## Arguments

- `subject` (required unless `--request` supplies it): The work being abandoned.
- `reason` (optional): Why the work is stopping.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Abandon** operation with these arguments. Report partial state and the safe retry boundary. Change nothing and clean up
nothing.
