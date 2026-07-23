---
description: Record one descriptive inline outcome without granting authority or changing a gate
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--outcome <text>]
  [--subject-provider <provider>] [--subject-revision <revision>]
  [--subject-kind <kind>] [--perspective <perspective>...]
  [--role <lens>...] [--criterion <criterion>...] [--method <method>]
  [--resolution <strict|assisted|automatic>]
---

# /xonovex-workflow:decide — Decide

## Arguments

- `subject` (required unless `--request` supplies it): Decision question or one
  provider-native reference.
- `--request` (optional): Structured request file for independently bound options,
  evidence, and criteria. Do not combine it with the shorthands.
- `--outcome` (optional): Outcome to record; otherwise derive a recommendation.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--perspective`, `--role`, `--criterion` (repeatable, optional): Explicit lenses,
  role conveniences, and decision criteria.
- `--method` (optional): Requested decision procedure.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Decide** operation with these arguments. Return the descriptive decision envelope
inline. The skill is the source of truth; do not approve, reject, merge, publish, or
change a protected gate.
