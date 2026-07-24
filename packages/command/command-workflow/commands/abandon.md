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
  [--reason <text>]
  [--criterion <criterion>...]
---

# /xonovex-workflow:abandon — Abandon

## Arguments

- `subject` (required unless `--request` supplies it): Exact work being stopped.
- `--request` (optional): Markdown file containing the subject, reason, and equivalent
  inputs. Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the subject.
- `--reason` (required unless `--request` supplies it): Present-tense reason for
  stopping.
- `--criterion` (repeatable, optional): Retention or recovery constraint.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Abandon** operation with these arguments. Return the reason, usable partial state,
unresolved work, and retry boundary; do not publish, clean, or mutate external state.
