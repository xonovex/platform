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
  [subject] [--request <file>] [--destination <reference>]
  [--expected-revision <revision>] [--criterion <criterion>...]
  [--idempotency-key <key>] [--effect <preview|apply>]
---

# /xonovex-workflow:publish — Publish

## Arguments

- `subject` (required unless `--request` supplies it): Exact inline result, path, or
  opaque native source reference.
- `--request` (optional): Markdown file containing the subject, destination, and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--destination` (required unless `--request` supplies it): Exact native destination
  reference.
- `--expected-revision` (optional): Expected destination revision for concurrency.
- `--criterion` (repeatable, optional): Binding publication precondition.
- `--idempotency-key` (required for `apply`): Stable retry key.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Publish** operation with these arguments. Return the observed destination reference,
revision, and effects without implying approval.
