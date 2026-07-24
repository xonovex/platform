---
description: Stop work and return an inline abandonment record without cleanup
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--context <context>...]
  [--reason <text>]
  [--criterion <criterion>...]
---

# /xonovex-workflow:abandon — Abandon

## Arguments

- `subject` (required unless `--request` supplies it): Exact work being stopped.
- `--request` (optional): Markdown workflow handoff containing the subject, reason,
  partial state, relationships, and equivalent inputs. Do not combine it with shorthand
  arguments.
- `--subject-revision` (optional): Exact native revision of the subject when available.
- `--context` (repeatable, optional): Canonical explanatory context or an opaque
  context reference to preserve with recovery information.
- `--reason` (required unless `--request` supplies it): Present-tense reason for
  stopping.
- `--criterion` (repeatable, optional): Retention or recovery constraint.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Abandon** operation with these arguments. Return the reason, usable partial state,
unresolved work, and retry boundary; do not publish, clean, or mutate external state.
