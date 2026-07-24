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
  [--feedback <feedback>...]
  [--criterion <criterion>...] [--method <method>]
---

# /xonovex-workflow:revise — Revise

## Arguments

- `subject` (required unless `--request` supplies it): Exact source to revise.
- `--request` (optional): Markdown file containing the source, feedback, and equivalent
  inputs. Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the source.
- `--feedback` (repeatable, required unless `--request` supplies it): Feedback to
  address.
- `--criterion` (repeatable, optional): Constraint the revision must retain.
- `--method` (optional): Requested subject-specific revision procedure.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Revise** operation with these arguments. Preserve the source and return the successor
inline without publishing it.
