---
description: Record workspace abandonment without mutation
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [reason] [--request <file>] [--context <context>...]
---

# /xonovex-workflow:workspace-abandon — Workspace abandon

## Arguments

- `subject` (required unless `--request` supplies it): The workspace being abandoned.
- `reason` (optional): Why it is being abandoned.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace abandon** operation with these arguments. Record the abandonment and the retry boundary. Remove nothing.
