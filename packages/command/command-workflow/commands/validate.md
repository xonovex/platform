---
description: Validate one exact subject against binding criteria and return evidence inline
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--criterion <criterion>...]
  [--perspective <perspective>...] [--method <method>]
---

# /xonovex-workflow:validate — Validate

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an
  opaque native reference.
- `--request` (optional): Markdown file containing the subject, criteria, and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the subject.
- `--criterion` (repeatable, required unless `--request` supplies it): Binding
  criterion to evaluate.
- `--perspective` (repeatable, optional): Advisory validation lens.
- `--method` (optional): Requested subject-specific validation procedure.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Validate** operation with these arguments. Return one evidence-backed result per
binding criterion; do not revise, accept, publish, or mutate the subject.
