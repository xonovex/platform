---
description: Validate one exact subject against binding criteria and return evidence inline
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Task
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--context <context>...]
  [--criterion <criterion>...]
  [--perspective <perspective>...] [--method <method>] [--independent]
---

# /xonovex-workflow:validate — Validate

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an
  opaque native reference.
- `--request` (optional): Markdown workflow handoff containing the subject, criteria,
  evidence, and required or preferred capabilities. Do not combine it with shorthand
  arguments.
- `--subject-revision` (optional): Exact native revision of the subject; required when
  a binding criterion needs revision-pinned evidence and the provider exposes one.
- `--context` (repeatable, optional): Canonical explanatory context or an opaque
  context reference to resolve without converting it into evidence or a binding
  outcome.
- `--criterion` (repeatable, required unless `--request` supplies it): Binding
  criterion to evaluate.
- `--perspective` (repeatable, optional): Advisory validation lens.
- `--method` (optional): Requested subject-specific validation procedure.
- `--independent` (optional): Evaluate the pinned subject without implementation or
  prior-review context first, then assess supplied context in a second pass and report
  its effect. Use separate invocations for independently accountable roles.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Validate** operation with these arguments. Return one evidence-backed result per
binding criterion; do not revise, accept, publish, or mutate the subject.
