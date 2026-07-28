---
description: Persist one exact result to one explicit destination
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
  [subject] [--request <file>] [--destination <destination>]
  [--expected-revision <revision>] [--effect <inspect|preview|apply>]
  [--idempotency-key <key>]
---

# /xonovex-workflow:publish — Publish

## Arguments

- `subject` (required unless `--request` supplies it): The exact result to persist.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--destination` (required unless `--request` supplies it): Exact destination.
- `--expected-revision` (optional): Destination revision the write expects. Blocks
  the write when the destination has moved.
- `--effect` (optional): `inspect`, `preview`, or `apply`. Defaults to `preview`.
- `--idempotency-key` (optional): Stable retry key. Required for an externally
  submitted `apply` when the provider supports idempotency.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Publish** operation with these arguments. Persist only the supplied result to the named destination. Preview before apply and
return the destination revision.
