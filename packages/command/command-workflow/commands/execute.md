---
description: Carry out previously specified work under an explicit effect mode
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
  [--context <context>...] [--effect <inspect|preview|apply>]
  [--idempotency-key <key>]
---

# /xonovex-workflow:execute — Execute

## Arguments

- `subject` (required unless `--request` supplies it): The plan, accepted request, decision, or feedback to carry out.
- `--request` (optional): Markdown workflow handoff carrying the subject and
  equivalent inputs. Do not combine it with shorthand arguments.
- `--revision` (optional): Exact native revision of the subject. Required for a
  protected provider-native subject when the provider exposes one.
- `--context` (repeatable, optional): Explanatory context, or an opaque reference to
  resolve. Context constrains the work; it is never evidence or approval.
- `--effect` (optional): `inspect`, `preview`, or `apply`. Defaults to `inspect`.
- `--idempotency-key` (optional): Stable retry key. Required for an externally
  submitted `apply` when the provider supports idempotency.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Execute** operation with these arguments. Execute expects work that was already specified. With no such antecedent, use
Create or a freeform session instead. Report every planned, applied, failed, or
unknown effect; do not publish or manage workspaces implicitly.
