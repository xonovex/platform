---
description: Create one new inline result without changing or publishing the source
allowed-tools:
  - Read
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
---

# /xonovex-workflow:create — Create

## Arguments

- `subject` (required unless `--request` supplies it): Inline source content or one
  provider-native reference.
- `--request` (optional): Structured request file for named, independently bound
  inputs. Do not combine it with the subject or selection shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--perspective`, `--role`, `--criterion` (repeatable, optional): Explicit lenses,
  role conveniences, and criteria. Roles may only suggest lenses and criteria.
- `--method` (optional): Requested creation procedure.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Create** operation with these arguments. Return its operation-result envelope
inline. The skill is the source of truth for the procedure and errors; do not publish
or persist the result from this command.
