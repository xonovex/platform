---
description: Evaluate binding criteria independently and report evidence
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [criteria] [--request <file>] [--revision <revision>]
  [--context <context>...]
---

# /xonovex-workflow:validate — Validate

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an opaque native reference.
- `criteria` (optional): The binding criteria to evaluate, inline or as a path.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--revision` (optional): Exact native revision of the subject. Required for a
  protected provider-native subject when the provider exposes one.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Validate** operation with these arguments. Evaluate each criterion against the pinned subject and report pass, fail, or
blocked with its evidence. The result authorizes nothing.
