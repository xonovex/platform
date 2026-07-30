---
description: Record one descriptive decision without granting authority
allowed-tools:
  - Read
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--context <context>...]
---

# /xonovex-workflow:decide - Decide

## Arguments

- `subject` (required unless `--request` supplies it): The question, with its options, as inline content, a path, or an opaque
  native reference.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Decide** operation with these arguments. Record one outcome with its reason. The result is descriptive: it authorizes no
publication, integration, or other protected effect.
