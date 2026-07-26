---
description: Review one exact subject and return evidence-linked findings inline
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--revision <revision>]
  [--context <context>...]
---

# /xonovex-workflow:review — Review

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an opaque native reference.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--revision` (optional): Exact native revision of the subject. Required for a
  protected provider-native subject when the provider exposes one.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Review** operation with these arguments. Return evidence-linked findings inline; do not modify or publish the subject.
When context is supplied, assess the subject without it first, then report what the
context changed.
