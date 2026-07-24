---
description: Review one exact subject and return evidence-linked findings inline
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
  [--perspective <perspective>...]
  [--criterion <criterion>...] [--method <method>] [--independent]
---

# /xonovex-workflow:review — Review

## Arguments

- `subject` (required unless `--request` supplies it): Inline content, a path, or an
  opaque native reference.
- `--request` (optional): Markdown workflow handoff containing the subject and
  equivalent inputs, including evidence and required or preferred capabilities. Do not
  combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the subject; required when
  revision-pinned review is a criterion and the provider exposes one.
- `--context` (repeatable, optional): Canonical explanatory context or an opaque
  context reference to resolve without accepting it as evidence, approval, or a
  substitute for independent review.
- `--perspective` (repeatable, optional): Explicit review lens.
- `--criterion` (repeatable, optional): Review criterion.
- `--method` (optional): Requested subject-specific review procedure.
- `--independent` (optional): Inspect the pinned subject without creator context or
  prior findings first, then assess supplied context in a second pass and report its
  effect. Use separate invocations for independently accountable perspectives.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Review** operation with these arguments. Return evidence-linked findings inline; do
not modify or publish the subject.
