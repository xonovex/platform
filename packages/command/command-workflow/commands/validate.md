---
description: Validate an exact subject against provenance-aware criteria and return evidence inline
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

# /xonovex-workflow:validate — Validate

## Arguments

- `subject` (required unless `--request` supplies it): Inline content or one
  provider-native reference.
- `--request` (optional): Structured request file for named subject, evidence, and
  criteria bindings. Do not combine it with the shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--perspective`, `--role`, `--criterion` (repeatable, optional): Explicit lenses,
  role conveniences, and criteria. Validation blocks when no binding criterion can
  be resolved.
- `--method` (optional): Requested validation procedure.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Validate** operation with these arguments. Return one criterion result and evidence
record per binding criterion in the inline operation-result envelope. The skill is
the source of truth; do not revise, accept, publish, or mutate the subject.
