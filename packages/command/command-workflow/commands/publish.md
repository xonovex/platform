---
description: Publish one exact subject to one explicit destination
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--destination <reference>]
  [--expected-revision <revision>] [--criterion <criterion>...]
  [--idempotency-key <key>] [--effect <preview|apply>]
---

# /xonovex-workflow:publish — Publish

## Arguments

- `subject` (required unless `--request` supplies it): Exact inline result, path, or
  opaque native source reference.
- `--request` (optional): Markdown workflow handoff containing the subject,
  destination, preconditions, retry identity, and equivalent inputs. Do not combine it
  with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the subject; required for
  `apply` when the provider exposes one.
- `--destination` (required unless `--request` supplies it): Exact native destination
  reference.
- `--expected-revision` (optional): Expected destination revision for concurrency;
  required for `apply` when an existing destination exposes one.
- `--criterion` (repeatable, optional): Binding publication precondition.
- `--idempotency-key` (optional): Stable retry key. Required for an externally
  submitted `apply` when the selected provider supports idempotency.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Publish** operation with these arguments. Return the observed destination reference,
revision, and effects without implying approval.
