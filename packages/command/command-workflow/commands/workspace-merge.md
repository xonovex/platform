---
description: Validate and integrate a workspace without cleanup
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

# /xonovex-workflow:workspace-merge - Workspace merge

## Arguments

- `subject` (required unless `--request` supplies it): The workspace to integrate.
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
**Workspace merge** operation with these arguments. Integrate into the destination and stop. Merging never implies branch, reference,
or worktree cleanup.
