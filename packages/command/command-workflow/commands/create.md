---
description: Create one new result inline without publishing it
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--context <context>...]
---

# /xonovex-workflow:create — Create

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an opaque native reference describing what to produce.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Create** operation with these arguments. Return the result inline. Do not publish it or create any external resource.
