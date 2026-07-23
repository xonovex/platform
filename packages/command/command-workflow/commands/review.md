---
description: Review an exact subject through repeatable perspectives and return inline findings
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

# /xonovex-workflow:review — Review

## Arguments

- `subject` (required unless `--request` supplies it): Inline content or one
  provider-native reference.
- `--request` (optional): Structured request file for named subject, evidence, and
  other provider-bound inputs. Do not combine it with the shorthands.
- `--subject-provider`, `--subject-revision`, `--subject-kind` (optional): Provider,
  exact native revision, and semantic kind for the simple subject shorthand.
- `--perspective`, `--role`, `--criterion` (repeatable, optional): Explicit lenses,
  role conveniences, and review criteria. Roles cannot grant authority.
- `--method` (optional): Requested review procedure.
- `--resolution` (optional): Constraint-resolution mode; defaults to `assisted`.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Review** operation with these arguments. Return evidence-linked findings and the
operation-result envelope inline. The skill is the source of truth; do not modify the
subject or persist the review from this command.
