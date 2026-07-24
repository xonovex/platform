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
  [subject] [--request <file>] [--criterion <criterion>...]
  [--method <method>] [--effect <inspect|preview|apply>]
---

# /xonovex-workflow:execute — Execute

## Arguments

- `subject` (required unless `--request` supplies it): Bounded instructions, a path,
  or an opaque native reference.
- `--request` (optional): Markdown file containing the subject and equivalent inputs.
  Do not combine it with shorthand arguments.
- `--criterion` (repeatable, optional): Completion criterion.
- `--method` (optional): Requested subject-specific execution procedure.
- `--effect` (optional): `inspect`, `preview`, or `apply`; defaults to `inspect`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Execute** operation with these arguments. Report every planned, applied, failed, or
unknown effect; do not publish results or manage workspaces implicitly.
