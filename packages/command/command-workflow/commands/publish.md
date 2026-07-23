---
description: Publish one exact subject to one explicit provider destination with idempotent effect reporting
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
  [subject] [--request <file>] [--subject-provider <provider>]
  [--subject-revision <revision>] [--subject-kind <kind>]
  [--destination-provider <provider>] [--destination-reference <reference>]
  [--destination-revision <revision>] [--criterion <criterion>...]
  [--idempotency-key <key>] [--effect <preview|apply>]
---

# /xonovex-workflow:publish — Publish

## Arguments

- `subject` (required unless `--request` supplies it): Exact inline result, operation
  result, or provider-native source reference.
- `--request` (optional): Structured request file for independently bound source,
  evidence, criteria, and destination. Do not combine it with the shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple source shorthand.
- `--destination-provider`, `--destination-reference` (required for shorthand):
  Provider and opaque native locator for the publication destination.
- `--destination-revision` (optional): Expected destination revision for optimistic
  concurrency.
- `--criterion` (repeatable, optional): Binding publication precondition.
- `--idempotency-key` (required for `apply`): Stable key for safe retry.
- `--effect` (optional): `preview` or `apply`; defaults to `preview`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Publish** operation with these arguments. This is the only core operation that may
persist a domain result. The skill is the source of truth; return the destination's
native locator, revision, and observed effect without implying approval.
