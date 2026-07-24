---
description: Record one descriptive inline outcome without changing a protected gate
allowed-tools:
  - Read
  - Glob
  - Grep
  - AskUserQuestion
  - Skill
argument-hint: >-
  [subject] [--request <file>] [--subject-revision <revision>]
  [--option <option>...] [--evidence <reference>...] [--outcome <text>]
  [--criterion <criterion>...] [--method <method>]
---

# /xonovex-workflow:decide — Decide

## Arguments

- `subject` (required unless `--request` supplies it): Decision question, options, or
  an opaque native reference.
- `--request` (optional): Markdown workflow handoff containing the question, options,
  evidence entries, and equivalent inputs. Do not combine it with shorthand arguments.
- `--subject-revision` (optional): Exact native revision of the decision subject;
  required when binding evidence is revision-pinned and the provider exposes one.
- `--option` (repeatable, optional): Explicit option to compare.
- `--evidence` (repeatable, optional): Opaque evidence reference to preserve and
  evaluate.
- `--outcome` (optional): Outcome to record; otherwise derive a recommendation.
- `--criterion` (repeatable, optional): Decision criterion.
- `--method` (optional): Requested subject-specific decision procedure.

## Delegation

Load the `workflow-guide` skill (plugin `xonovex-skill-workflow`) and perform its
**Decide** operation with these arguments. Return a descriptive decision; do not
approve, reject, merge, publish, or change a protected gate.
