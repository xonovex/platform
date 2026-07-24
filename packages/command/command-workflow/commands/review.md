---
description: Review one exact subject and return evidence-linked findings inline
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--perspective <perspective>...]
  [--criterion <criterion>...] [--method <method>]
---

# /xonovex-workflow:review — Review

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an
  opaque native reference.
- `--request` (optional): Markdown file containing the subject and equivalent inputs.
  Do not combine it with shorthand arguments.
- `--perspective` (repeatable, optional): Explicit review lens.
- `--criterion` (repeatable, optional): Review criterion.
- `--method` (optional): Requested subject-specific review procedure.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Review** operation with these arguments. Return evidence-linked findings inline; do
not modify or publish the subject.
