---
description: Execute one bounded subject under an explicit effect mode
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
  [--context <context>...]
  [--criterion <criterion>...]
  [--method <method>] [--idempotency-key <key>]
  [--effect <inspect|preview|apply>]
---

# /xonovex-workflow:execute — Execute

## Arguments

- `subject` (required unless `--request` supplies it): Bounded instructions, a path,
  or an opaque native reference.
- `--request` (optional): Markdown workflow handoff containing the subject, effects,
  relationships, and equivalent inputs. Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the subject; required for a
  protected provider-native subject when the provider exposes one.
- `--context` (repeatable, optional): Canonical explanatory context or an opaque
  context reference to resolve; active context constrains the work and is preserved
  in the result.
- `--criterion` (repeatable, optional): Completion criterion.
- `--method` (optional): Requested subject-specific execution procedure.
- `--idempotency-key` (optional): Stable retry key. Required for an externally
  submitted `apply` when the selected provider supports idempotency.
- `--effect` (optional): `inspect`, `preview`, or `apply`; defaults to `inspect`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Execute** operation with these arguments. Report every planned, applied, failed, or
unknown effect; do not publish results or manage workspaces implicitly.
