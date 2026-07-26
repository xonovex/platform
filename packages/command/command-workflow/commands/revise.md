---
description: Revise one subject into a traceable successor
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [feedback] [--request <file>] [--revision <revision>]
  [--context <context>...]
---

# /xonovex-workflow:revise — Revise

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an opaque native reference.
- `feedback` (optional): The change to apply, as explicit feedback.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--revision` (optional): Exact native revision of the subject. Required for a
  protected provider-native subject when the provider exposes one.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Revise** operation with these arguments. Return the successor inline and preserve the source. Do not publish it.
