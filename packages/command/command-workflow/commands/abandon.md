---
description: Stop work and return an inline abandonment record without cleanup or provider mutation
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--reason <text>]
  [--subject-provider <provider>] [--subject-revision <revision>]
  [--subject-kind <kind>] [--criterion <criterion>...]
  [--resolution <strict|assisted|automatic>]
---

# /xonovex-workflow:abandon — Abandon

## Arguments

- `subject` and `--reason` (required unless `--request` supplies them): Exact work
  being stopped and the present-tense reason.
- `--request` (optional): Structured request file for the subject, partial results,
  evidence, reason, and criteria. Do not combine it with the shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--criterion` (repeatable, optional): Retention or recovery constraint.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Abandon** operation with these arguments. Return the reason, partial state,
unresolved work, and retry boundary inline. The skill is the source of truth; do not
publish, delete, clean, snapshot, or mutate provider state.
