---
description: Produce one traceable inline revision from explicit feedback without overwriting its source
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--feedback <feedback>...]
  [--subject-provider <provider>] [--subject-revision <revision>]
  [--subject-kind <kind>] [--perspective <perspective>...]
  [--role <lens>...] [--criterion <criterion>...] [--method <method>]
  [--resolution <strict|assisted|automatic>]
---

# /xonovex-workflow:revise — Revise

## Arguments

- `subject` and `--feedback` (required unless `--request` supplies them): Exact source
  plus one or more inline feedback items or provider-native references.
- `--request` (optional): Structured request file for independently bound source,
  feedback, evidence, and criteria. Do not combine it with the shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--perspective`, `--role`, `--criterion` (repeatable, optional): Explicit lenses,
  role conveniences, and constraints the revision must retain.
- `--method` (optional): Requested revision procedure.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Revise** operation with these arguments. Return the successor and its
operation-result envelope inline. The skill is the source of truth; preserve the
source and do not persist the successor from this command.
