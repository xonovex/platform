---
description: Create one new inline result without changing or publishing its source
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--criterion <criterion>...]
  [--method <method>]
---

# /xonovex-workflow:create — Create

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an
  opaque native reference.
- `--request` (optional): Markdown file containing the subject and equivalent inputs.
  Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the subject.
- `--criterion` (repeatable, optional): Constraint the result must satisfy.
- `--method` (optional): Requested subject-specific procedure.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Create** operation with these arguments. Return the result inline; do not publish or
persist it.
