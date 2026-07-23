---
description: Execute one bounded subject with an explicit effect mode and return observed effects inline
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
  [--perspective <perspective>...] [--role <lens>...]
  [--criterion <criterion>...] [--method <method>]
  [--resolution <strict|assisted|automatic>]
  [--effect <inspect|preview|apply>]
---

# /xonovex-workflow:execute — Execute

## Arguments

- `subject` (required unless `--request` supplies it): Bounded instructions or one
  provider-native reference.
- `--request` (optional): Structured request file for named, independently bound
  inputs. Do not combine it with the shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--perspective`, `--role`, `--criterion` (repeatable, optional): Explicit lenses,
  role conveniences, and completion criteria.
- `--method` (optional): Requested execution procedure.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.
- `--effect` (optional): `inspect`, `preview`, or `apply`; defaults to `inspect`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Execute** operation with these arguments. Return the result envelope and every
planned, applied, failed, or unknown effect inline. The skill is the source of truth;
do not publish results or manage workspaces implicitly.
