---
description: Create only the named isolated workspace resources
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
  [subject] [--request <file>] [--source <source>] [--revision <revision>]
  [--effect <inspect|preview|apply>] [--idempotency-key <key>]
---

# /xonovex-workflow:workspace-create — Workspace create

## Arguments

- `subject` (required unless `--request` supplies it): Name of the workspace to create.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--source` (required unless `--request` supplies it): Exact source to isolate.
- `--revision` (optional): Exact native revision of the subject. Required for a
  protected provider-native subject when the provider exposes one.
- `--effect` (optional): `inspect`, `preview`, or `apply`. Defaults to `preview`.
- `--idempotency-key` (optional): Stable retry key. Required for an externally
  submitted `apply` when the provider supports idempotency.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Workspace create** operation with these arguments. Create only the named isolation resources. Perform none of the work that will run
inside them.
