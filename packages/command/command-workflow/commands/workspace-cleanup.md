---
description: Preview and remove only the exact named workspace resources
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
  [subject] [--request <file>] [--revision <revision>]
  [--effect <inspect|preview|apply>] [--idempotency-key <key>]
---

# /xonovex-workflow:workspace-cleanup — Workspace cleanup

## Arguments

- `subject` (required unless `--request` supplies it): Exact resources to remove.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--revision` (optional): Exact native revision of the subject. Required for a
  protected provider-native subject when the provider exposes one.
- `--effect` (optional): `inspect`, `preview`, or `apply`. Defaults to `inspect`.
- `--idempotency-key` (optional): Stable retry key. Required for an externally
  submitted `apply` when the provider supports idempotency.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace cleanup** operation with these arguments. Remove only the exact named resources. Never widen the set; preview before apply.
