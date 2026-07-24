---
description: Produce one traceable inline revision without overwriting its source
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--context <context>...]
  [--feedback <feedback>...]
  [--criterion <criterion>...] [--method <method>]
---

# /xonovex-workflow:revise — Revise

## Arguments

- `subject` (required unless `--request` supplies it): Exact source to revise.
- `--request` (optional): Markdown workflow handoff containing the source, feedback,
  relationships, and equivalent inputs. Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the source; required for a
  provider-native source when the provider exposes one.
- `--context` (repeatable, optional): Canonical explanatory context or an opaque
  context reference to resolve and retain when producing the successor.
- `--feedback` (repeatable, required unless `--request` supplies it): Feedback to
  address.
- `--criterion` (repeatable, optional): Constraint the revision must retain.
- `--method` (optional): Requested subject-specific revision procedure.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Revise** operation with these arguments. Preserve the source and return the successor
inline without publishing it.
